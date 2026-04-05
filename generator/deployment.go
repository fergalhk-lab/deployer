package generator

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateDeployment(svc config.Service, cfg config.Config) *appsv1.Deployment {
	labels := BuildLabels(svc.Name, cfg.Name)
	selector := map[string]string{
		"app.kubernetes.io/name": svc.Name,
	}
	replicas := int32(1)
	podSpec := BuildPodSpec(svc.Runnable, svc.Name)
	podSpec.Containers[0].ReadinessProbe = buildReadinessProbe(svc)

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
				Spec: podSpec,
			},
		},
	}
}

// buildReadinessProbe derives the readiness probe for a service according to
// the following rules:
//   - If healthcheck.port is explicitly set to 0: no probe.
//   - If healthcheck.port is explicitly set to N: probe on port N.
//   - If healthcheck.port is nil and the service has an ingress: probe on the ingress port.
//   - Otherwise: no probe.
//
// The path defaults to /readyz unless overridden by healthcheck.path.
func buildReadinessProbe(svc config.Service) *corev1.Probe {
	var port uint16
	if svc.HealthCheck != nil && svc.HealthCheck.Port != nil {
		port = *svc.HealthCheck.Port
	} else if svc.Ingress != nil {
		port = svc.Ingress.Port
	}
	if port == 0 {
		return nil
	}
	path := "/readyz"
	if svc.HealthCheck != nil && svc.HealthCheck.Path != "" {
		path = svc.HealthCheck.Path
	}
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt32(int32(port)),
			},
		},
	}
}
