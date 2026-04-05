package generator

import (
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/fergalhk-lab/deployer/config"
)

// externalSecretManifest is a minimal marshaling struct that omits secretStoreRef,
// which older external-secrets versions reject when empty.
type externalSecretManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              externalSecretManifestSpec `json:"spec"`
}

type externalSecretManifestSpec struct {
	RefreshInterval *metav1.Duration                            `json:"refreshInterval,omitempty"`
	Target          esv1beta1.ExternalSecretTarget              `json:"target,omitempty"`
	DataFrom        []esv1beta1.ExternalSecretDataFromRemoteRef `json:"dataFrom,omitempty"`
}

func marshalExternalSecret(es *esv1beta1.ExternalSecret) ([]byte, error) {
	return yaml.Marshal(externalSecretManifest{
		TypeMeta:   es.TypeMeta,
		ObjectMeta: es.ObjectMeta,
		Spec: externalSecretManifestSpec{
			RefreshInterval: es.Spec.RefreshInterval,
			Target:          es.Spec.Target,
			DataFrom:        es.Spec.DataFrom,
		},
	})
}

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
