package generator

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fergalhk-lab/deployer/config"
)

func GenerateCronJob(job config.CronJob, cfg config.Config) *batchv1.CronJob {
	labels := BuildLabels(job.Name, cfg.Name)

	return &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: cfg.Name,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: job.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: buildJobSpec(job.BaseJob, labels),
			},
		},
	}
}
