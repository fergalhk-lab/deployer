package generator_test

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

var update = flag.Bool("update", false, "update golden files")

func fullConfig() config.Config {
	rawVal := "info"
	maxRetries := uint(3)
	return config.Config{
		Name: "myapp",
		Services: []config.Service{
			{
				Name: "api",
				Runnable: config.Runnable{
					Image:      config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
					Command:    []string{"./api"},
					Args:       []string{"--port=8080"},
					Resources:  config.Resources{CPU: "100m", Memory: "128Mi"},
					Env: []config.Env{
						{Name: "LOG_LEVEL", RawValue: &rawVal},
						{Name: "DB_PASSWORD", FromSecret: &config.SecretRef{Name: "my-secret", Key: "password"}},
					},
					IAMRoleARN: "arn:aws:iam::123456789012:role/api-role",
				},
				Ingress: &config.Ingress{
					Port:   8080,
					Public: &config.PublicIngress{Domain: "api.example.com"},
				},
			},
			{
				Name: "worker",
				Runnable: config.Runnable{
					Image:     config.Image{Repository: "my-registry/worker", Tag: "1.0.0"},
					Resources: config.Resources{CPU: "200m", Memory: "256Mi"},
				},
			},
		},
		InitJobs: []config.InitJob{
			{
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
			},
		},
		CronJobs: []config.CronJob{
			{
				BaseJob: config.BaseJob{
					Name: "cleanup",
					Runnable: config.Runnable{
						Image:     config.Image{Repository: "my-registry/cleanup", Tag: "1.0.0"},
						Command:   []string{"./cleanup"},
						Resources: config.Resources{CPU: "50m", Memory: "64Mi"},
					},
				},
				Schedule: "0 2 * * *",
			},
		},
	}
}

func TestGenerate_GoldenFile(t *testing.T) {
	cfg := fullConfig()
	opts := generator.Options{IngressClass: "traefik"}

	got, err := generator.Generate(cfg, opts)
	require.NoError(t, err, "Generate")

	goldenPath := "testdata/expected.yaml"

	if *update {
		require.NoError(t, os.MkdirAll("testdata", 0755), "mkdir testdata")
		require.NoError(t, os.WriteFile(goldenPath, got, 0644), "write golden")
		return
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden file %q (run with -update to generate)", goldenPath)

	assert.Equal(t, string(expected), string(got))
}

// Ensure Generate returns valid YAML by round-tripping through yaml.Unmarshal
func TestGenerate_ValidYAML(t *testing.T) {
	cfg := fullConfig()
	opts := generator.Options{IngressClass: "traefik"}

	got, err := generator.Generate(cfg, opts)
	require.NoError(t, err, "Generate")

	parts := bytes.Split(got, []byte("---\n"))
	for i, part := range parts {
		if len(bytes.TrimSpace(part)) == 0 {
			continue
		}
		var m map[string]interface{}
		assert.NoError(t, yaml.Unmarshal(part, &m), "document %d is not valid YAML", i)
	}
}
