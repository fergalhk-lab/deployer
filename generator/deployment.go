package generator

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateDeployment(svc config.Service, cfg config.Config) *appsv1.Deployment {
	labels := BuildLabels(svc.Name, cfg.Name)
	selector := map[string]string{
		"app.kubernetes.io/name": svc.Name,
	}
	replicas := int32(1)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: cfg.Name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": mainContainerName,
					},
				},
				Spec: BuildPodSpec(svc.Runnable),
			},
		},
	}
}
