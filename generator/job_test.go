package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

func initJob() config.InitJob {
	maxRetries := uint(3)
	return config.InitJob{
		BaseJob: config.BaseJob{
			Name: "migrate",
			Runnable: config.Runnable{
				Image:     config.Image{Repository: "my-registry/migrate", Tag: "1.0.0"},
				Command:   []string{"./migrate"},
				Args:      []string{"up"},
				Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
			},
			MaxRetries: &maxRetries,
		},
	}
}

func TestGenerateJob_TypeMeta(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	if j.APIVersion != "batch/v1" || j.Kind != "Job" {
		t.Errorf("TypeMeta = %q/%q, want batch/v1/Job", j.APIVersion, j.Kind)
	}
}

func TestGenerateJob_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	if j.Name != "migrate" {
		t.Errorf("Name = %q, want %q", j.Name, "migrate")
	}
	if j.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", j.Namespace, "myapp")
	}
}

func TestGenerateJob_BackoffLimit(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	if j.Spec.BackoffLimit == nil || *j.Spec.BackoffLimit != 3 {
		t.Errorf("BackoffLimit = %v, want 3", j.Spec.BackoffLimit)
	}
}

func TestGenerateJob_NilMaxRetries_DefaultsToZero(t *testing.T) {
	job := config.InitJob{
		BaseJob: config.BaseJob{
			Name: "migrate",
			Runnable: config.Runnable{
				Image:     config.Image{Repository: "my-registry/migrate", Tag: "1.0.0"},
				Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
			},
			MaxRetries: nil,
		},
	}
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(job, cfg)
	if j.Spec.BackoffLimit == nil || *j.Spec.BackoffLimit != 0 {
		t.Errorf("BackoffLimit = %v, want 0", j.Spec.BackoffLimit)
	}
}

func TestGenerateJob_PodTemplateAnnotation(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	ann := j.Spec.Template.Annotations
	if ann["kubectl.kubernetes.io/default-container"] != "main" {
		t.Errorf("default-container annotation = %q, want %q", ann["kubectl.kubernetes.io/default-container"], "main")
	}
}

func TestGenerateJob_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	if j.Labels["app.kubernetes.io/name"] != "migrate" {
		t.Errorf("label name = %q, want %q", j.Labels["app.kubernetes.io/name"], "migrate")
	}
	if j.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", j.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if j.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", j.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	if got, want := len(j.Labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}
	// Pod template labels must match
	for k, v := range j.Labels {
		if j.Spec.Template.Labels[k] != v {
			t.Errorf("pod template label[%q] = %q, want %q", k, j.Spec.Template.Labels[k], v)
		}
	}
}

func TestGenerateJob_RestartPolicy(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	if j.Spec.Template.Spec.RestartPolicy != "OnFailure" {
		t.Errorf("RestartPolicy = %q, want OnFailure", j.Spec.Template.Spec.RestartPolicy)
	}
}
