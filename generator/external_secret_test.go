package generator_test

import (
	"testing"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateExternalSecret_TypeMeta(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, "external-secrets.io/v1beta1", es.APIVersion)
	assert.Equal(t, "ExternalSecret", es.Kind)
}

func TestGenerateExternalSecret_ObjectMeta(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, "app-key", es.Name)
	assert.Equal(t, "myapp", es.Namespace)
	assert.Equal(t, "deployer", es.Labels["app.kubernetes.io/managed-by"])
	assert.Equal(t, "app-key", es.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", es.Labels["app.kubernetes.io/part-of"])
}

func TestGenerateExternalSecret_RefreshIntervalZero(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	require.NotNil(t, es.Spec.RefreshInterval)
	assert.Equal(t, int64(0), es.Spec.RefreshInterval.Duration.Nanoseconds())
}

func TestGenerateExternalSecret_TargetName(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, "app-key", es.Spec.Target.Name)
}

func TestGenerateExternalSecret_TargetCreationPolicy(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	assert.Equal(t, esv1beta1.CreatePolicyOwner, es.Spec.Target.CreationPolicy)
}

func TestGenerateExternalSecret_DataFromGeneratorRef(t *testing.T) {
	es := generator.GenerateExternalSecret(testGeneratedSecret(), testAppConfig())
	require.Len(t, es.Spec.DataFrom, 1)
	sourceRef := es.Spec.DataFrom[0].SourceRef
	require.NotNil(t, sourceRef)
	require.NotNil(t, sourceRef.GeneratorRef)
	assert.Equal(t, "generators.external-secrets.io/v1alpha1", sourceRef.GeneratorRef.APIVersion)
	assert.Equal(t, "Password", sourceRef.GeneratorRef.Kind)
	assert.Equal(t, "app-key", sourceRef.GeneratorRef.Name)
}
