package generator_test

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "networking.k8s.io/v1", ing.APIVersion)
	assert.Equal(t, "Ingress", ing.Kind)
}

func TestGenerateIngress_IngressClass(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	require.NotNil(t, ing.Spec.IngressClassName)
	assert.Equal(t, "traefik", *ing.Spec.IngressClassName)
}

func TestGenerateIngress_Host(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	require.Len(t, ing.Spec.Rules, 1)
	assert.Equal(t, "api.example.com", ing.Spec.Rules[0].Host)
}

func TestGenerateIngress_PathAndType(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	paths := ing.Spec.Rules[0].HTTP.Paths
	require.Len(t, paths, 1)
	assert.Equal(t, "/", paths[0].Path)
	require.NotNil(t, paths[0].PathType)
	assert.Equal(t, networkingv1.PathTypePrefix, *paths[0].PathType)
}

func TestGenerateIngress_BackendService(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	require.NotEmpty(t, ing.Spec.Rules, "ingress has no rules")
	require.NotNil(t, ing.Spec.Rules[0].HTTP, "ingress has no HTTP rules")
	require.NotEmpty(t, ing.Spec.Rules[0].HTTP.Paths, "ingress has no paths")
	backend := ing.Spec.Rules[0].HTTP.Paths[0].Backend.Service
	assert.Equal(t, "api", backend.Name)
	assert.Equal(t, int32(8080), backend.Port.Number)
}

func TestGenerateIngress_Namespace(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	assert.Equal(t, "api", ing.Name)
	assert.Equal(t, "myapp", ing.Namespace)
}

func TestGenerateIngress_Labels(t *testing.T) {
	svc := svcWithPublicIngress()
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	assert.Equal(t, "api", ing.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", ing.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", ing.Labels["app.kubernetes.io/managed-by"])
	assert.Len(t, ing.Labels, 3)
}

func ptr[T any](v T) *T { return &v }

func TestGenerateIngress_PrefixPath(t *testing.T) {
	svc := config.Service{
		Name: "api",
		Runnable: config.Runnable{
			Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
			Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
		},
		Ingress: &config.Ingress{
			Port: 8080,
			Public: &config.PublicIngress{
				Domain: "api.example.com",
				Path:   &config.IngressPath{Prefix: ptr("/api/")},
			},
		},
	}
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	paths := ing.Spec.Rules[0].HTTP.Paths
	require.Len(t, paths, 1)
	assert.Equal(t, "/api/", paths[0].Path)
	require.NotNil(t, paths[0].PathType)
	assert.Equal(t, networkingv1.PathTypePrefix, *paths[0].PathType)
}

func TestGenerateIngress_LiteralPath(t *testing.T) {
	svc := config.Service{
		Name: "api",
		Runnable: config.Runnable{
			Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
			Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
		},
		Ingress: &config.Ingress{
			Port: 8080,
			Public: &config.PublicIngress{
				Domain: "api.example.com",
				Path:   &config.IngressPath{Literal: ptr("/foo/bar")},
			},
		},
	}
	cfg := config.Config{Name: "myapp"}
	opts := generator.Options{IngressClass: "traefik"}
	ing := generator.GenerateIngress(svc, cfg, opts)
	paths := ing.Spec.Rules[0].HTTP.Paths
	require.Len(t, paths, 1)
	assert.Equal(t, "/foo/bar", paths[0].Path)
	require.NotNil(t, paths[0].PathType)
	assert.Equal(t, networkingv1.PathTypeExact, *paths[0].PathType)
}
