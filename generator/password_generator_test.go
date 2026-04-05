package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGeneratedSecret() config.GeneratedSecret {
	return config.GeneratedSecret{Name: "app-key", Length: 32, Symbols: false}
}

func testAppConfig() config.Config {
	return config.Config{Name: "myapp"}
}

func TestGeneratePasswordGenerator_TypeMeta(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, "generators.external-secrets.io/v1alpha1", p.APIVersion)
	assert.Equal(t, "Password", p.Kind)
}

func TestGeneratePasswordGenerator_ObjectMeta(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, "app-key", p.Name)
	assert.Equal(t, "myapp", p.Namespace)
	assert.Equal(t, "deployer", p.Labels["app.kubernetes.io/managed-by"])
	assert.Equal(t, "app-key", p.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", p.Labels["app.kubernetes.io/part-of"])
}

func TestGeneratePasswordGenerator_Length(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, 32, p.Spec.Length)
}

func TestGeneratePasswordGenerator_SymbolsFalse(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	require.NotNil(t, p.Spec.Symbols)
	assert.Equal(t, 0, *p.Spec.Symbols)
}

func TestGeneratePasswordGenerator_SymbolsTrue(t *testing.T) {
	gs := config.GeneratedSecret{Name: "app-key", Length: 32, Symbols: true}
	p := generator.GeneratePasswordGenerator(gs, testAppConfig())
	require.NotNil(t, p.Spec.Symbols)
	assert.Equal(t, 1, *p.Spec.Symbols)
}

func TestGeneratePasswordGenerator_AllowRepeat(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	assert.True(t, p.Spec.AllowRepeat)
}

func TestGeneratePasswordGenerator_SecretKeys(t *testing.T) {
	p := generator.GeneratePasswordGenerator(testGeneratedSecret(), testAppConfig())
	require.Len(t, p.Spec.SecretKeys, 1)
	assert.Equal(t, "value", p.Spec.SecretKeys[0])
}

