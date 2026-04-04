package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
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
	if c.Name != "main" {
		t.Errorf("Name = %q, want %q", c.Name, "main")
	}
}

func TestBuildContainer_Image(t *testing.T) {
	c := generator.BuildContainer(runnable())
	if c.Image != "my-registry/api:1.0.0" {
		t.Errorf("Image = %q, want %q", c.Image, "my-registry/api:1.0.0")
	}
}

func TestBuildContainer_CommandAndArgs(t *testing.T) {
	c := generator.BuildContainer(runnable())
	if len(c.Command) != 1 || c.Command[0] != "./api" {
		t.Errorf("Command = %v, want [./api]", c.Command)
	}
	if len(c.Args) != 1 || c.Args[0] != "--port=8080" {
		t.Errorf("Args = %v, want [--port=8080]", c.Args)
	}
}

func TestBuildContainer_EnvSkipsNilRawValue(t *testing.T) {
	c := generator.BuildContainer(runnable())
	if len(c.Env) != 1 {
		t.Fatalf("len(Env) = %d, want 1", len(c.Env))
	}
	if c.Env[0].Name != "LOG_LEVEL" || c.Env[0].Value != "info" {
		t.Errorf("Env[0] = %+v, want {Name:LOG_LEVEL Value:info}", c.Env[0])
	}
}

func TestBuildContainer_CPURequestOnly(t *testing.T) {
	c := generator.BuildContainer(runnable())
	if _, ok := c.Resources.Requests[corev1.ResourceCPU]; !ok {
		t.Error("expected CPU request to be set")
	}
	if _, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
		t.Error("expected no CPU limit")
	}
	wantCPU := resource.MustParse("100m")
	if !c.Resources.Requests[corev1.ResourceCPU].Equal(wantCPU) {
		t.Errorf("CPU request = %v, want %v", c.Resources.Requests[corev1.ResourceCPU], wantCPU)
	}
}

func TestBuildContainer_MemoryRequestAndLimit(t *testing.T) {
	c := generator.BuildContainer(runnable())
	if _, ok := c.Resources.Requests[corev1.ResourceMemory]; !ok {
		t.Error("expected memory request to be set")
	}
	if _, ok := c.Resources.Limits[corev1.ResourceMemory]; !ok {
		t.Error("expected memory limit to be set")
	}
	if c.Resources.Requests[corev1.ResourceMemory] != c.Resources.Limits[corev1.ResourceMemory] {
		t.Error("expected memory request and limit to be equal")
	}
	wantMem := resource.MustParse("128Mi")
	if !c.Resources.Requests[corev1.ResourceMemory].Equal(wantMem) {
		t.Errorf("memory request = %v, want %v", c.Resources.Requests[corev1.ResourceMemory], wantMem)
	}
	if !c.Resources.Limits[corev1.ResourceMemory].Equal(wantMem) {
		t.Errorf("memory limit = %v, want %v", c.Resources.Limits[corev1.ResourceMemory], wantMem)
	}
}

func TestBuildPodSpec_ContainerIsMain(t *testing.T) {
	spec := generator.BuildPodSpec(runnable(), "test")
	if len(spec.Containers) != 1 {
		t.Fatalf("len(Containers) = %d, want 1", len(spec.Containers))
	}
	if spec.Containers[0].Name != "main" {
		t.Errorf("container name = %q, want %q", spec.Containers[0].Name, "main")
	}
	if spec.Containers[0].Image != "my-registry/api:1.0.0" {
		t.Errorf("container image = %q, want %q", spec.Containers[0].Image, "my-registry/api:1.0.0")
	}
}

func TestBuildPodSpec_ServiceAccountName(t *testing.T) {
	spec := generator.BuildPodSpec(runnable(), "myservice")
	if spec.ServiceAccountName != "myservice" {
		t.Errorf("ServiceAccountName = %q, want %q", spec.ServiceAccountName, "myservice")
	}
}

func runnableWithIAM() config.Runnable {
	r := runnable()
	r.IAMRoleARN = "arn:aws:iam::123456789012:role/my-role"
	return r
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
	if roleARN != "arn:aws:iam::123456789012:role/my-role" {
		t.Errorf("AWS_ROLE_ARN = %q, want %q", roleARN, "arn:aws:iam::123456789012:role/my-role")
	}
	if tokenFile != "/var/run/secrets/eks.amazonaws.com/serviceaccount/token" {
		t.Errorf("AWS_WEB_IDENTITY_TOKEN_FILE = %q, want %q", tokenFile, "/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
	}
}

func TestBuildContainer_IAMEnvVarsAfterUserEnv(t *testing.T) {
	c := generator.BuildContainer(runnableWithIAM())
	// User env (LOG_LEVEL) should come before IAM env vars
	if len(c.Env) < 3 {
		t.Fatalf("len(Env) = %d, want >= 3", len(c.Env))
	}
	if c.Env[0].Name != "LOG_LEVEL" {
		t.Errorf("Env[0].Name = %q, want LOG_LEVEL", c.Env[0].Name)
	}
}

func TestBuildContainer_NoIAMEnvVarsWithoutRoleARN(t *testing.T) {
	c := generator.BuildContainer(runnable())
	for _, e := range c.Env {
		if e.Name == "AWS_ROLE_ARN" || e.Name == "AWS_WEB_IDENTITY_TOKEN_FILE" {
			t.Errorf("unexpected IAM env var %q in non-IAM runnable", e.Name)
		}
	}
}
