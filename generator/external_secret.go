package generator

import (
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateExternalSecret(gs config.GeneratedSecret, cfg config.Config) *esv1beta1.ExternalSecret {
	refreshInterval := metav1.Duration{Duration: 0}

	return &esv1beta1.ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "external-secrets.io/v1beta1",
			Kind:       "ExternalSecret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gs.Name,
			Namespace: cfg.Name,
			Labels:    BuildLabels(gs.Name, cfg.Name),
		},
		Spec: esv1beta1.ExternalSecretSpec{
			RefreshInterval: &refreshInterval,
			Target: esv1beta1.ExternalSecretTarget{
				Name:           gs.Name,
				CreationPolicy: esv1beta1.CreatePolicyOwner,
			},
			DataFrom: []esv1beta1.ExternalSecretDataFromRemoteRef{
				{
					SourceRef: &esv1beta1.StoreGeneratorSourceRef{
						GeneratorRef: &esv1beta1.GeneratorRef{
							APIVersion: "generators.external-secrets.io/v1alpha1",
							Kind:       "Password",
							Name:       gs.Name,
						},
					},
				},
			},
		},
	}
}
