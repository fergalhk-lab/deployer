package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runnable() config.Runnable {
	rawVal := "info"
	return config.Runnable{
		Image:   config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
		Command: []string{"./api"},
		Args:    []string{"--port=8080"},
		Resources: config.Resources{
			CPU:    "100m",
			Memory: "128Mi",
		},
		Env: []config.Env{
			{Name: "LOG_LEVEL", RawValue: &rawVal},
			{Name: "IGNORED", RawValue: nil}, // nil RawValue should be skipped
		},
	}
}

func TestBuildContainer_Name(t *testing.T) {
	c := generator.BuildContainer(runnable())
	assert.Equal(t, "main", c.Name)
}

func TestBuildContainer_Image(t *testing.T) {
	c := generator.BuildContainer(runnable())
	assert.Equal(t, "my-registry/api:1.0.0", c.Image)
}

func TestBuildContainer_CommandAndArgs(t *testing.T) {
	c := generator.BuildContainer(runnable())
	require.Len(t, c.Command, 1)
	assert.Equal(t, "./api", c.Command[0])
	require.Len(t, c.Args, 1)
	assert.Equal(t, "--port=8080", c.Args[0])
}

func TestBuildContainer_EnvSkipsNilRawValue(t *testing.T) {
	c := generator.BuildContainer(runnable())
	require.Len(t, c.Env, 1)
	assert.Equal(t, "LOG_LEVEL", c.Env[0].Name)
	assert.Equal(t, "info", c.Env[0].Value)
}

func TestBuildContainer_CPURequestOnly(t *testing.T) {
	c := generator.BuildContainer(runnable())
	assert.Contains(t, c.Resources.Requests, corev1.ResourceCPU)
	assert.NotContains(t, c.Resources.Limits, corev1.ResourceCPU)
	wantCPU := resource.MustParse("100m")
	assert.True(t, c.Resources.Requests[corev1.ResourceCPU].Equal(wantCPU),
		"CPU request = %v, want %v", c.Resources.Requests[corev1.ResourceCPU], wantCPU)
}

func TestBuildContainer_MemoryRequestAndLimit(t *testing.T) {
	c := generator.BuildContainer(runnable())
	assert.Contains(t, c.Resources.Requests, corev1.ResourceMemory)
	assert.Contains(t, c.Resources.Limits, corev1.ResourceMemory)
	wantMem := resource.MustParse("128Mi")
	assert.True(t, c.Resources.Requests[corev1.ResourceMemory].Equal(wantMem),
		"memory request = %v, want %v", c.Resources.Requests[corev1.ResourceMemory], wantMem)
	assert.True(t, c.Resources.Limits[corev1.ResourceMemory].Equal(wantMem),
		"memory limit = %v, want %v", c.Resources.Limits[corev1.ResourceMemory], wantMem)
}

func TestBuildPodSpec_ContainerIsMain(t *testing.T) {
	spec := generator.BuildPodSpec(runnable(), "")
	require.Len(t, spec.Containers, 1)
	assert.Equal(t, "main", spec.Containers[0].Name)
	assert.Equal(t, "my-registry/api:1.0.0", spec.Containers[0].Image)
}

func TestBuildPodSpec_ServiceAccountName(t *testing.T) {
	spec := generator.BuildPodSpec(runnable(), "myservice")
	assert.Equal(t, "myservice", spec.ServiceAccountName)
}

func runnableWithIAM() config.Runnable {
	r := runnable()
	r.IAMRoleARN = "arn:aws:iam::123456789012:role/my-role"
	return r
}

func runnableWithGeneratedSecret() config.Runnable {
	secretName := "app-key"
	return config.Runnable{
		Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
		Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
		Env: []config.Env{
			{Name: "APP_KEY", FromGeneratedSecret: &secretName},
		},
	}
}

func TestBuildContainer_IAMEnvVars(t *testing.T) {
	c := generator.BuildContainer(runnableWithIAM())
	var roleARN, tokenFile string
	for _, e := range c.Env {
		switch e.Name {
		case "AWS_ROLE_ARN":
			roleARN = e.Value
		case "AWS_WEB_IDENTITY_TOKEN_FILE":
			tokenFile = e.Value
		}
	}
	assert.Equal(t, "arn:aws:iam::123456789012:role/my-role", roleARN)
	assert.Equal(t, "/var/run/secrets/eks.amazonaws.com/serviceaccount/token", tokenFile)
}

func TestBuildContainer_IAMEnvVarsAfterUserEnv(t *testing.T) {
	c := generator.BuildContainer(runnableWithIAM())
	n := len(c.Env)
	require.GreaterOrEqual(t, n, 3)
	assert.Equal(t, "LOG_LEVEL", c.Env[0].Name)
	assert.Equal(t, "AWS_ROLE_ARN", c.Env[n-2].Name)
	assert.Equal(t, "AWS_WEB_IDENTITY_TOKEN_FILE", c.Env[n-1].Name)
}

func TestBuildContainer_NoIAMEnvVarsWithoutRoleARN(t *testing.T) {
	c := generator.BuildContainer(runnable())
	var envNames []string
	for _, e := range c.Env {
		envNames = append(envNames, e.Name)
	}
	assert.NotContains(t, envNames, "AWS_ROLE_ARN")
	assert.NotContains(t, envNames, "AWS_WEB_IDENTITY_TOKEN_FILE")
}

func TestBuildPodSpec_IAMVolume(t *testing.T) {
	spec := generator.BuildPodSpec(runnableWithIAM(), "myservice")
	require.Len(t, spec.Volumes, 1)
	vol := spec.Volumes[0]
	assert.Equal(t, "aws-iam-token", vol.Name)
	require.NotNil(t, vol.Projected)
	require.Len(t, vol.Projected.Sources, 1)
	src := vol.Projected.Sources[0]
	require.NotNil(t, src.ServiceAccountToken)
	assert.Equal(t, "sts.amazonaws.com", src.ServiceAccountToken.Audience)
	require.NotNil(t, src.ServiceAccountToken.ExpirationSeconds)
	assert.Equal(t, int64(86400), *src.ServiceAccountToken.ExpirationSeconds)
	assert.Equal(t, "token", src.ServiceAccountToken.Path)
}

func TestBuildPodSpec_IAMVolumeMount(t *testing.T) {
	spec := generator.BuildPodSpec(runnableWithIAM(), "myservice")
	require.Len(t, spec.Containers, 1)
	mounts := spec.Containers[0].VolumeMounts
	require.Len(t, mounts, 1)
	assert.Equal(t, "aws-iam-token", mounts[0].Name)
	assert.Equal(t, "/var/run/secrets/eks.amazonaws.com/serviceaccount", mounts[0].MountPath)
	assert.True(t, mounts[0].ReadOnly)
}

func TestBuildPodSpec_NoIAMWithoutRoleARN(t *testing.T) {
	spec := generator.BuildPodSpec(runnable(), "myservice")
	assert.Empty(t, spec.Volumes)
	require.Len(t, spec.Containers, 1)
	assert.Empty(t, spec.Containers[0].VolumeMounts)
}

func TestBuildContainer_EnvFromGeneratedSecret(t *testing.T) {
	c := generator.BuildContainer(runnableWithGeneratedSecret())
	require.Len(t, c.Env, 1)
	e := c.Env[0]
	assert.Equal(t, "APP_KEY", e.Name)
	assert.Empty(t, e.Value)
	require.NotNil(t, e.ValueFrom)
	require.NotNil(t, e.ValueFrom.SecretKeyRef)
	assert.Equal(t, "app-key", e.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "value", e.ValueFrom.SecretKeyRef.Key)
}

func TestBuildContainer_EnvFromGeneratedSecretOptionalIsNil(t *testing.T) {
	c := generator.BuildContainer(runnableWithGeneratedSecret())
	require.Len(t, c.Env, 1)
	assert.Nil(t, c.Env[0].ValueFrom.SecretKeyRef.Optional,
		"optional field must be nil (required by default)")
}
