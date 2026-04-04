package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerateServiceAccount_TypeMeta(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	assert.Equal(t, "v1", sa.APIVersion)
	assert.Equal(t, "ServiceAccount", sa.Kind)
}

func TestGenerateServiceAccount_Name(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	assert.Equal(t, "api", sa.Name)
}

func TestGenerateServiceAccount_Namespace(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	assert.Equal(t, "myapp", sa.Namespace)
}

func TestGenerateServiceAccount_Labels(t *testing.T) {
	cfg := config.Config{Name: "myapp"}
	sa := generator.GenerateServiceAccount("api", cfg)
	assert.Equal(t, "api", sa.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", sa.Labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", sa.Labels["app.kubernetes.io/managed-by"])
}
