package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/generator"
	"github.com/stretchr/testify/assert"
)

func TestBuildLabels(t *testing.T) {
	labels := generator.BuildLabels("api", "myapp")

	assert.Len(t, labels, 3)
	assert.Equal(t, "api", labels["app.kubernetes.io/name"])
	assert.Equal(t, "myapp", labels["app.kubernetes.io/part-of"])
	assert.Equal(t, "deployer", labels["app.kubernetes.io/managed-by"])
}
