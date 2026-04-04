package generator

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateJob(job config.InitJob, cfg config.Config) *batchv1.Job {
	labels := BuildLabels(job.Name, cfg.Name)

	var backoffLimit int32
	if job.MaxRetries != nil {
		backoffLimit = int32(*job.MaxRetries)
	}

	podSpec := BuildPodSpec(job.Runnable)
	podSpec.RestartPolicy = corev1.RestartPolicyOnFailure

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: cfg.Name,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
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
