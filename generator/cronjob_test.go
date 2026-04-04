package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

func cronJob() config.CronJob {
	return config.CronJob{
		BaseJob: config.BaseJob{
			Name: "cleanup",
			Runnable: config.Runnable{
				Image:     config.Image{Repository: "my-registry/cleanup", Tag: "1.0.0"},
				Command:   []string{"./cleanup"},
				Resources: config.Resources{CPU: "50m", Memory: "64Mi"},
			},
		},
		Schedule: "0 2 * * *",
	}
}

func TestGenerateCronJob_TypeMeta(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if cj.APIVersion != "batch/v1" || cj.Kind != "CronJob" {
		t.Errorf("TypeMeta = %q/%q, want batch/v1/CronJob", cj.APIVersion, cj.Kind)
	}
}

func TestGenerateCronJob_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if cj.Name != "cleanup" {
		t.Errorf("Name = %q, want %q", cj.Name, "cleanup")
	}
	if cj.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", cj.Namespace, "myapp")
	}
}

func TestGenerateCronJob_Schedule(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if cj.Spec.Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want %q", cj.Spec.Schedule, "0 2 * * *")
	}
}

func TestGenerateCronJob_BackoffLimit(t *testing.T) {
	maxRetries := uint(2)
	job := config.CronJob{
		BaseJob: config.BaseJob{
			Name:       "cleanup",
			Runnable:   config.Runnable{Image: config.Image{Repository: "r", Tag: "t"}, Resources: config.Resources{CPU: "50m", Memory: "64Mi"}},
			MaxRetries: &maxRetries,
		},
		Schedule: "0 * * * *",
	}
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(job, cfg)
	if got := cj.Spec.JobTemplate.Spec.BackoffLimit; got == nil || *got != 2 {
		t.Errorf("BackoffLimit = %v, want 2", got)
	}
}

func TestGenerateCronJob_NilMaxRetries_DefaultsToZero(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if got := cj.Spec.JobTemplate.Spec.BackoffLimit; got == nil || *got != 0 {
		t.Errorf("BackoffLimit = %v, want 0", got)
	}
}

func TestGenerateCronJob_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if cj.Labels["app.kubernetes.io/name"] != "cleanup" {
		t.Errorf("label name = %q, want %q", cj.Labels["app.kubernetes.io/name"], "cleanup")
	}
	if cj.Labels["app.kubernetes.io/part-of"] != "myapp" {
		t.Errorf("label part-of = %q, want %q", cj.Labels["app.kubernetes.io/part-of"], "myapp")
	}
	if cj.Labels["app.kubernetes.io/managed-by"] != "deployer" {
		t.Errorf("label managed-by = %q, want %q", cj.Labels["app.kubernetes.io/managed-by"], "deployer")
	}
	if got, want := len(cj.Labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}
	podLabels := cj.Spec.JobTemplate.Spec.Template.Labels
	for k, v := range cj.Labels {
		if podLabels[k] != v {
			t.Errorf("pod template label[%q] = %q, want %q", k, podLabels[k], v)
		}
	}
}

func TestGenerateCronJob_PodTemplateAnnotation(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	ann := cj.Spec.JobTemplate.Spec.Template.Annotations
	if ann["kubectl.kubernetes.io/default-container"] != "main" {
		t.Errorf("default-container annotation = %q, want %q", ann["kubectl.kubernetes.io/default-container"], "main")
	}
}

func TestGenerateCronJob_RestartPolicy(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	if cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy != "OnFailure" {
		t.Errorf("RestartPolicy = %q, want OnFailure", cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy)
	}
}
