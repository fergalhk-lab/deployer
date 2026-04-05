package generator_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "batch/v1", j.APIVersion)
	assert.Equal(t, "Job", j.Kind)
}

func TestGenerateJob_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	assert.Equal(t, "migrate", j.Name)
	assert.Equal(t, "myapp", j.Namespace)
}

func TestGenerateJob_BackoffLimit(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	require.NotNil(t, j.Spec.BackoffLimit)
	assert.Equal(t, int32(3), *j.Spec.BackoffLimit)
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
	require.NotNil(t, j.Spec.BackoffLimit)
	assert.Equal(t, int32(0), *j.Spec.BackoffLimit)
}

func TestGenerateJob_PodTemplateAnnotation(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	assert.Equal(t, "main", j.Spec.Template.Annotations["kubectl.kubernetes.io/default-container"])
}

func TestGenerateJob_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	assert.Equal(t, "migrate", j.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", j.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", j.Labels["app.kubernetes.io/managed-by"])
	assert.Len(t, j.Labels, 3)
	// Pod template labels must match
	for k, v := range j.Labels {
		assert.Equal(t, v, j.Spec.Template.Labels[k], "pod template label %q", k)
	}
}

func TestGenerateJob_RestartPolicy(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	j := generator.GenerateJob(initJob(), cfg)
	assert.Equal(t, corev1.RestartPolicyOnFailure, j.Spec.Template.Spec.RestartPolicy)
}
