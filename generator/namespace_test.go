package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNamespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	ns := generator.GenerateNamespace(cfg)

	assert.Equal(t, "myapp", ns.Name)
	assert.Equal(t, "v1", ns.APIVersion)
	assert.Equal(t, "Namespace", ns.Kind)
	assert.Equal(t, "myapp", ns.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", ns.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", ns.Labels["app.kubernetes.io/managed-by"])
	assert.Len(t, ns.Labels, 3)
}
