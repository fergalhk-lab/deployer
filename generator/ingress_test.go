package generator_test

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

func svcWithPublicIngress() config.Service {
	return config.Service{
		Name: "api",
		Runnable: config.Runnable{
			Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
			Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
		},
		Ingress: &config.Ingress{
			Port:   8080,
			Public: &config.PublicIngress{Domain: "api.example.com"},
		},
	}
}

func TestGenerateIngress_TypeMeta(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if ing.APIVersion != "networking.k8s.io/v1" || ing.Kind != "Ingress" {
		t.Errorf("TypeMeta = %q/%q, want networking.k8s.io/v1/Ingress", ing.APIVersion, ing.Kind)
	}
}

func TestGenerateIngress_IngressClass(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if ing.Spec.IngressClassName == nil || *ing.Spec.IngressClassName != "traefik" {
		t.Errorf("IngressClassName = %v, want traefik", ing.Spec.IngressClassName)
	}
}

func TestGenerateIngress_Host(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if len(ing.Spec.Rules) != 1 || ing.Spec.Rules[0].Host != "api.example.com" {
		t.Errorf("host = %v, want api.example.com", ing.Spec.Rules)
	}
}

func TestGenerateIngress_PathAndType(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	paths := ing.Spec.Rules[0].HTTP.Paths
	if len(paths) != 1 {
		t.Fatalf("len(paths) = %d, want 1", len(paths))
	}
	if paths[0].Path != "/" {
		t.Errorf("path = %q, want /", paths[0].Path)
	}
	if paths[0].PathType == nil || *paths[0].PathType != networkingv1.PathTypePrefix {
		t.Errorf("pathType = %v, want Prefix", paths[0].PathType)
	}
}

func TestGenerateIngress_BackendService(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if len(ing.Spec.Rules) == 0 || ing.Spec.Rules[0].HTTP == nil || len(ing.Spec.Rules[0].HTTP.Paths) == 0 {
		t.Fatal("ingress has no rules/paths to check backend")
	}
	backend := ing.Spec.Rules[0].HTTP.Paths[0].Backend.Service
	if backend.Name != "api" {
		t.Errorf("backend service name = %q, want %q", backend.Name, "api")
	}
	if backend.Port.Number != 8080 {
		t.Errorf("backend port = %d, want 8080", backend.Port.Number)
	}
}

func TestGenerateIngress_Namespace(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if ing.Name != "api" {
		t.Errorf("Name = %q, want %q", ing.Name, "api")
	}
	if ing.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", ing.Namespace, "myapp")
	}
}

func TestGenerateIngress_Labels(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	if ing.Labels["app.kubernetes.io/name"] != "api" {
		t.Errorf("label name = %q, want %q", ing.Labels["app.kubernetes.io/name"], "api")
	}
	if ing.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", ing.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if ing.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", ing.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	if got, want := len(ing.Labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}
}
