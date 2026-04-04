package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
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
	if s.APIVersion != "v1" || s.Kind != "Service" {
		t.Errorf("TypeMeta = %q/%q, want v1/Service", s.APIVersion, s.Kind)
	}
}

func TestGenerateService_Namespace(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	if s.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", s.Namespace, "myapp")
	}
}

func TestGenerateService_ClusterIP(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	if s.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Errorf("Type = %q, want ClusterIP", s.Spec.Type)
	}
}

func TestGenerateService_Port(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	if len(s.Spec.Ports) != 1 || s.Spec.Ports[0].Port != 8080 {
		t.Errorf("port = %v, want 8080", s.Spec.Ports)
	}
}

func TestGenerateService_SelectorUsesName(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	if s.Spec.Selector["app.kubernetes.io/name"] != "api" {
		t.Errorf("selector = %v, want name=api", s.Spec.Selector)
	}
}
