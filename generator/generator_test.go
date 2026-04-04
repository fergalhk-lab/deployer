package generator_test

import (
	"flag"
	"os"
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
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
					Image:     config.Image{Repository: "my-registry/api", Tag: "1.0.0"},
					Command:   []string{"./api"},
					Args:      []string{"--port=8080"},
					Resources: config.Resources{CPU: "100m", Memory: "128Mi"},
					Env:       []config.Env{{Name: "LOG_LEVEL", RawValue: &rawVal}},
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
	}
}

func TestGenerate_GoldenFile(t *testing.T) {
	cfg := fullConfig()
	opts := generator.Options{IngressClass: "traefik"}

	got, err := generator.Generate(cfg, opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	goldenPath := "testdata/expected.yaml"

	if *update {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %q: %v (run with -update to generate)", goldenPath, err)
	}

	if string(got) != string(expected) {
		t.Errorf("output differs from golden file\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

// Ensure Generate returns valid YAML by round-tripping through yaml.Unmarshal
func TestGenerate_ValidYAML(t *testing.T) {
	cfg := fullConfig()
	opts := generator.Options{IngressClass: "traefik"}

	got, err := generator.Generate(cfg, opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var m map[string]interface{}
	if err := yaml.Unmarshal(got, &m); err != nil {
		t.Errorf("output is not valid YAML: %v", err)
	}
}
