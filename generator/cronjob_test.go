package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "batch/v1", cj.APIVersion)
	assert.Equal(t, "CronJob", cj.Kind)
}

func TestGenerateCronJob_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	assert.Equal(t, "cleanup", cj.Name)
	assert.Equal(t, "myapp", cj.Namespace)
}

func TestGenerateCronJob_Schedule(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	assert.Equal(t, "0 2 * * *", cj.Spec.Schedule)
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
	require.NotNil(t, cj.Spec.JobTemplate.Spec.BackoffLimit)
	assert.Equal(t, int32(2), *cj.Spec.JobTemplate.Spec.BackoffLimit)
}

func TestGenerateCronJob_NilMaxRetries_DefaultsToZero(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	require.NotNil(t, cj.Spec.JobTemplate.Spec.BackoffLimit)
	assert.Equal(t, int32(0), *cj.Spec.JobTemplate.Spec.BackoffLimit)
}

func TestGenerateCronJob_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	assert.Equal(t, "cleanup", cj.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", cj.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", cj.Labels["app.kubernetes.io/managed-by"])
	assert.Len(t, cj.Labels, 3)
	podLabels := cj.Spec.JobTemplate.Spec.Template.Labels
	for k, v := range cj.Labels {
		assert.Equal(t, v, podLabels[k], "pod template label %q", k)
	}
}

func TestGenerateCronJob_PodTemplateAnnotation(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	assert.Equal(t, "main", cj.Spec.JobTemplate.Spec.Template.Annotations["kubectl.kubernetes.io/default-container"])
}

func TestGenerateCronJob_RestartPolicy(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	cj := generator.GenerateCronJob(cronJob(), cfg)
	assert.Equal(t, corev1.RestartPolicyOnFailure, cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy)
}
