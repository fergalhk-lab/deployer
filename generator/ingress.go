package generator

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateIngress(svc config.Service, cfg config.Config, opts Options) *networkingv1.Ingress {
	path, pathType := ingressPath(svc.Ingress.Public.Path)
	port := int32(svc.Ingress.Port)
	ingressClass := opts.IngressClass

	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: cfg.Name,
			Labels:    BuildLabels(svc.Name, cfg.Name),
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClass,
			Rules: []networkingv1.IngressRule{
				{
					Host: svc.Ingress.Public.Domain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svc.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ingressPath returns the path string and PathType for the given IngressPath config.
// When p is nil it returns the default: "/" with PathTypePrefix.
func ingressPath(p *config.IngressPath) (string, networkingv1.PathType) {
	if p == nil {
		return "/", networkingv1.PathTypePrefix
	}
	if p.Literal != nil {
		return *p.Literal, networkingv1.PathTypeExact
	}
	if p.Prefix != nil {
		return *p.Prefix, networkingv1.PathTypePrefix
	}
	return "/", networkingv1.PathTypePrefix
}
