package generator_test

import (
	"testing"

	"github.com/fergalhk-lab/deployer/generator"
)

func TestBuildLabels(t *testing.T) {
	labels := generator.BuildLabels("api", "myapp")

	if got, want := len(labels), 3; got != want {
		t.Errorf("len(labels) = %d, want %d", got, want)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"app.kubernetes.io/name", "api"},
		{"app.kubernetes.io/part-of", "myapp"},
		{"app.kubernetes.io/managed-by", "deployer"},
	}
	for _, tt := range tests {
		if got := labels[tt.key]; got != tt.want {
			t.Errorf("labels[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}
