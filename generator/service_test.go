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
	if s.Name != "api" {
		t.Errorf("Name = %q, want %q", s.Name, "api")
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
	if s.Spec.Ports[0].Protocol != corev1.ProtocolTCP {
		t.Errorf("Protocol = %q, want TCP", s.Spec.Ports[0].Protocol)
	}
}

func TestGenerateService_SelectorUsesName(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	sel := s.Spec.Selector
	if sel["app.kubernetes.io/name"] != "api" {
		t.Errorf("selector name = %q, want %q", sel["app.kubernetes.io/name"], "api")
	}
	if _, ok := sel["app.kubernetes.io/part-of"]; ok {
		t.Error("selector should not contain part-of label")
	}
	if _, ok := sel["app.kubernetes.io/managed-by"]; ok {
		t.Error("selector should not contain managed-by label")
	}
	if len(sel) != 1 {
		t.Errorf("selector len = %d, want 1", len(sel))
	}
}

func TestGenerateService_Labels(t *testing.T) {
	svc := svcWithIngress()
	cfg := config.Config{Name: "myapp"}
	s := generator.GenerateService(svc, cfg)
	if s.Labels["app.kubernetes.io/name"] != "api" {
		t.Errorf("label name = %q, want %q", s.Labels["app.kubernetes.io/name"], "api")
	}
	if s.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", s.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if s.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", s.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	if got, want := len(s.Labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}
}
