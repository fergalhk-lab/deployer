package generator

import (
	genv1alpha1 "github.com/external-secrets/external-secrets/apis/generators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GeneratePasswordGenerator(gs config.GeneratedSecret, cfg config.Config) *genv1alpha1.Password {
	symbols := 0
	if gs.Symbols {
		symbols = 1
	}

	return &genv1alpha1.Password{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "generators.external-secrets.io/v1alpha1",
			Kind:       "Password",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gs.Name,
			Namespace: cfg.Name,
			Labels:    BuildLabels(gs.Name, cfg.Name),
		},
		Spec: genv1alpha1.PasswordSpec{
			Length:      gs.Length,
			Symbols:     &symbols,
			AllowRepeat: true,
		},
	}
}
