package generator

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateService(svc config.Service, cfg config.Config) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: cfg.Name,
			Labels:    BuildLabels(svc.Name, cfg.Name),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app.kubernetes.io/name": svc.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:     int32(svc.Ingress.Port),
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}
