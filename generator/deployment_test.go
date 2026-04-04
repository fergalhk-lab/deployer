package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
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
	if d.APIVersion != "apps/v1" || d.Kind != "Deployment" {
		t.Errorf("TypeMeta = %q/%q, want apps/v1/Deployment", d.APIVersion, d.Kind)
	}
}

func TestGenerateDeployment_Namespace(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	if d.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", d.Namespace, "myapp")
	}
	if d.Name != "api" {
		t.Errorf("Name = %q, want %q", d.Name, "api")
	}
}

func TestGenerateDeployment_OneReplica(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	if d.Spec.Replicas == nil || *d.Spec.Replicas != 1 {
		t.Errorf("Replicas = %v, want 1", d.Spec.Replicas)
	}
}

func TestGenerateDeployment_Labels(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	if d.Labels["app.kubernetes.io/name"] != "api" {
		t.Errorf("label name = %q, want %q", d.Labels["app.kubernetes.io/name"], "api")
	}
	if d.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", d.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if d.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", d.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	// Pod template labels must match ObjectMeta labels
	for k, v := range d.Labels {
		if d.Spec.Template.Labels[k] != v {
			t.Errorf("pod template label[%q] = %q, want %q", k, d.Spec.Template.Labels[k], v)
		}
	}
}

func TestGenerateDeployment_SelectorUsesName(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	sel := d.Spec.Selector.MatchLabels
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

func TestGenerateDeployment_PodTemplateAnnotation(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	ann := d.Spec.Template.Annotations
	if ann["kubectl.kubernetes.io/default-container"] != "main" {
		t.Errorf("default-container annotation = %q, want %q", ann["kubectl.kubernetes.io/default-container"], "main")
	}
}

func TestGenerateDeployment_ContainerImage(t *testing.T) {
	svc, cfg := svcAndCfg()
	d := generator.GenerateDeployment(svc, cfg)
	containers := d.Spec.Template.Spec.Containers
	if len(containers) != 1 || containers[0].Image != "my-registry/api:1.0.0" {
		t.Errorf("container image = %q, want %q", containers[0].Image, "my-registry/api:1.0.0")
	}
}
