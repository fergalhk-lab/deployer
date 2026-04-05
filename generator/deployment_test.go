package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func svcAndCfg() (config.Service, config.Config) {
	rawVal := "info"
	svc := config.Service{
		Name: "api",
		Runnable: config.Runnable{
			Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
			Command:   []string{"./api"},
			Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
			Env:       []config.Env{{Name: "LOG_LEVEL", RawValue: &rawVal}},
		},
	}
	cfg := config.Config{Name: "myapp"}
	return svc, cfg
}

func TestGenerateDeployment_TypeMeta(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	assert.Equal(t, "apps/v1", d.APIVersion)
	assert.Equal(t, "Deployment", d.Kind)
}

func TestGenerateDeployment_Namespace(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	assert.Equal(t, "myapp", d.Namespace)
	assert.Equal(t, "api", d.Name)
}

func TestGenerateDeployment_OneReplica(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	require.NotNil(t, d.Spec.Replicas)
	assert.Equal(t, int32(1), *d.Spec.Replicas)
}

func TestGenerateDeployment_Labels(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	assert.Equal(t, "api", d.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", d.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", d.Labels["app.kubernetes.io/managed-by"])
	// Pod template labels must match ObjectMeta labels
	for k, v := range d.Labels {
		assert.Equal(t, v, d.Spec.Template.Labels[k], "pod template label %q", k)
	}
}

func TestGenerateDeployment_SelectorUsesName(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	sel := d.Spec.Selector.MatchLabels
	assert.Equal(t, "api", sel["app.kubernetes.io/name"])
	assert.NotContains(t, sel, "app.kubernetes.io/part-of")
	assert.NotContains(t, sel, "app.kubernetes.io/managed-by")
	assert.Len(t, sel, 1)
}

func TestGenerateDeployment_PodTemplateAnnotation(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	assert.Equal(t, "main", d.Spec.Template.Annotations["kubectl.kubernetes.io/default-container"])
}

func TestGenerateDeployment_ContainerImage(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	containers := d.Spec.Template.Spec.Containers
	require.Len(t, containers, 1)
	assert.Equal(t, "my-registry/api:1.0.0", containers[0].Image)
}

func svcWithIngressForDeployment() (config.Service, config.Config) {
	svc, cfg := svcAndCfg()
	svc.Ingress = &config.Ingress{Port: 8080}
	return svc, cfg
}

func ptrDeploy[T any](v T) *T { return &v }

// Case 1: ingress present, no healthcheck field → probe on ingress port, path /readyz
func TestGenerateDeployment_ReadinessProbe_IngressDefault(t *testing.T) {
	svc, cfg := svcWithIngressForDeployment()
	d := generator.GenerateDeployment(svc, cfg)
	require.Len(t, d.Spec.Template.Spec.Containers, 1)
	probe := d.Spec.Template.Spec.Containers[0].ReadinessProbe
	require.NotNil(t, probe)
	require.NotNil(t, probe.HTTPGet)
	assert.Equal(t, "/readyz", probe.HTTPGet.Path)
	assert.Equal(t, int32(8080), probe.HTTPGet.Port.IntVal)
}

// Case 2: ingress present, healthcheck.path set only → probe on ingress port, custom path
func TestGenerateDeployment_ReadinessProbe_IngressCustomPath(t *testing.T) {
	svc, cfg := svcWithIngressForDeployment()
	svc.HealthCheck = &config.HealthCheck{Path: "/healthz"}
	d := generator.GenerateDeployment(svc, cfg)
	probe := d.Spec.Template.Spec.Containers[0].ReadinessProbe
	require.NotNil(t, probe)
	require.NotNil(t, probe.HTTPGet)
	assert.Equal(t, "/healthz", probe.HTTPGet.Path)
	assert.Equal(t, int32(8080), probe.HTTPGet.Port.IntVal)
}

// Case 3: ingress present, healthcheck.port &0 → no probe
func TestGenerateDeployment_ReadinessProbe_ExplicitlyDisabled(t *testing.T) {
	svc, cfg := svcWithIngressForDeployment()
	svc.HealthCheck = &config.HealthCheck{Port: ptrDeploy(uint16(0))}
	d := generator.GenerateDeployment(svc, cfg)
	assert.Nil(t, d.Spec.Template.Spec.Containers[0].ReadinessProbe)
}

// Case 4: ingress present, healthcheck.port &N → probe on port N
func TestGenerateDeployment_ReadinessProbe_ExplicitPort(t *testing.T) {
	svc, cfg := svcWithIngressForDeployment()
	svc.HealthCheck = &config.HealthCheck{Port: ptrDeploy(uint16(9090))}
	d := generator.GenerateDeployment(svc, cfg)
	probe := d.Spec.Template.Spec.Containers[0].ReadinessProbe
	require.NotNil(t, probe)
	require.NotNil(t, probe.HTTPGet)
	assert.Equal(t, int32(9090), probe.HTTPGet.Port.IntVal)
	assert.Equal(t, "/readyz", probe.HTTPGet.Path)
}

// Case 5: no ingress, no healthcheck → no probe
func TestGenerateDeployment_ReadinessProbe_NoIngressNoHealthcheck(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	assert.Nil(t, d.Spec.Template.Spec.Containers[0].ReadinessProbe)
}

// Case 6: no ingress, healthcheck.port &N → probe on port N
func TestGenerateDeployment_ReadinessProbe_NoIngressExplicitPort(t *testing.T) {
	svc, cfg := svcAndCfg()
	svc.HealthCheck = &config.HealthCheck{Port: ptrDeploy(uint16(8080))}
	d := generator.GenerateDeployment(svc, cfg)
	probe := d.Spec.Template.Spec.Containers[0].ReadinessProbe
	require.NotNil(t, probe)
	require.NotNil(t, probe.HTTPGet)
	assert.Equal(t, int32(8080), probe.HTTPGet.Port.IntVal)
	assert.Equal(t, "/readyz", probe.HTTPGet.Path)
}
