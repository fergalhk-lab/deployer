package generator

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fergalhk-lab/deployer/config"
)

const mainContainerName = "main"

func BuildContainer(r config.Runnable) corev1.Container {
	var env []corev1.EnvVar
	for _, e := range r.Env {
		if e.RawValue == nil {
			continue
		}
		env = append(env, corev1.EnvVar{
			Name:  e.Name,
			Value: *e.RawValue,
		})
	}

	if r.IAMRoleARN != "" {
		env = append(env,
			corev1.EnvVar{Name: "AWS_ROLE_ARN", Value: r.IAMRoleARN},
			corev1.EnvVar{Name: "AWS_WEB_IDENTITY_TOKEN_FILE", Value: "/var/run/secrets/eks.amazonaws.com/serviceaccount/token"},
		)
	}

	return corev1.Container{
		Name:    mainContainerName,
		Image:   fmt.Sprintf("%s:%s", r.Image.Repository, r.Image.Tag),
		Command: r.Command,
		Args:    r.Args,
		Env:     env,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(r.Resources.CPU),
				corev1.ResourceMemory: resource.MustParse(r.Resources.Memory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(r.Resources.Memory),
			},
		},
	}
}

func BuildPodSpec(r config.Runnable, name string) corev1.PodSpec {
	return corev1.PodSpec{
		ServiceAccountName: name,
		Containers:         []corev1.Container{BuildContainer(r)},
	}
}
