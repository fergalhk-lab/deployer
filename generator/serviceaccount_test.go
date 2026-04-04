package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

func TestGenerateServiceAccount_TypeMeta(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	if sa.APIVersion != "v1" || sa.Kind != "ServiceAccount" {
		t.Errorf("TypeMeta = %q/%q, want v1/ServiceAccount", sa.APIVersion, sa.Kind)
	}
}

func TestGenerateServiceAccount_Name(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	if sa.Name != "api" {
		t.Errorf("Name = %q, want %q", sa.Name, "api")
	}
}

func TestGenerateServiceAccount_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	if sa.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", sa.Namespace, "myapp")
	}
}

func TestGenerateServiceAccount_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	if sa.Labels["app.kubernetes.io/name"] != "api" {
		t.Errorf("label name = %q, want %q", sa.Labels["app.kubernetes.io/name"], "api")
	}
	if sa.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", sa.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if sa.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", sa.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
}
