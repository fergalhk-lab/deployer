package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func svcWithIngress() config.Service {
	return config.Service{
		Name: "api",
		Runnable: config.Runnable{
			Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
			Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
		},
		Ingress: &config.Ingress{Port: 8080},
	}
}

func TestGenerateService_TypeMeta(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	assert.Equal(t, "v1", s.APIVersion)
	assert.Equal(t, "Service", s.Kind)
}

func TestGenerateService_Namespace(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	assert.Equal(t, "myapp", s.Namespace)
	assert.Equal(t, "api", s.Name)
}

func TestGenerateService_ClusterIP(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	assert.Equal(t, corev1.ServiceTypeClusterIP, s.Spec.Type)
}

func TestGenerateService_Port(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	require.Len(t, s.Spec.Ports, 1)
	assert.Equal(t, int32(8080), s.Spec.Ports[0].Port)
	assert.Equal(t, corev1.ProtocolTCP, s.Spec.Ports[0].Protocol)
}

func TestGenerateService_SelectorUsesName(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	sel := s.Spec.Selector
	assert.Equal(t, "api", sel["app.kubernetes.io/name"])
	assert.NotContains(t, sel, "app.kubernetes.io/part-of")
	assert.NotContains(t, sel, "app.kubernetes.io/managed-by")
	assert.Len(t, sel, 1)
}

func TestGenerateService_Labels(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	assert.Equal(t, "api", s.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", s.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", s.Labels["app.kubernetes.io/managed-by"])
	assert.Len(t, s.Labels, 3)
}
