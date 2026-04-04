package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

func TestGenerateNamespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	ns := generator.GenerateNamespace(cfg)

	if ns.Name != "myapp" {
		t.Errorf("Name = %q, want %q", ns.Name, "myapp")
	}
	if ns.APIVersion != "v1" {
		t.Errorf("APIVersion = %q, want %q", ns.APIVersion, "v1")
	}
	if ns.Kind != "Namespace" {
		t.Errorf("Kind = %q, want %q", ns.Kind, "Namespace")
	}
	if ns.Labels["app.kubernetes.io/name"] != "myapp" {
		t.Errorf("label name = %q, want %q", ns.Labels["app.kubernetes.io/name"], "myapp")
	}
	if ns.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", ns.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if ns.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", ns.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	if got, want := len(ns.Labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}
}
