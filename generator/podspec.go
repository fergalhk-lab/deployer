package generator

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fergalhk-lab/deployer/config"
)

const mainContainerName = "main"

const (
	iamTokenDir  = "/var/run/secrets/eks.amazonaws.com/serviceaccount"
	iamTokenFile = iamTokenDir + "/token"
)

func BuildContainer(r config.Runnable) corev1.Container {
	var env []corev1.EnvVar
	for _, e := range r.Env {
		switch {
		case e.RawValue != nil:
			env = append(env, corev1.EnvVar{
				Name:  e.Name,
				Value: *e.RawValue,
			})
		case e.FromSecret != nil:
			env = append(env, corev1.EnvVar{
				Name: e.Name,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: e.FromSecret.Name},
						Key:                  e.FromSecret.Key,
					},
				},
			})
		}
	}

	if r.IAMRoleARN != "" {
		env = append(env,
			corev1.EnvVar{Name: "AWS_ROLE_ARN", Value: r.IAMRoleARN},
			corev1.EnvVar{Name: "AWS_WEB_IDENTITY_TOKEN_FILE", Value: iamTokenFile},
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
	container := BuildContainer(r)

	spec := corev1.PodSpec{
		ServiceAccountName: name,
		Containers:         []corev1.Container{container},
	}

	if r.IAMRoleARN != "" {
		expiry := int64(86400)
		spec.Volumes = []corev1.Volume{
			{
				Name: "aws-iam-token",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
									Audience:          "sts.amazonaws.com",
									ExpirationSeconds: &expiry,
									Path:              "token",
								},
							},
						},
					},
				},
			},
		}
		spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "aws-iam-token",
				MountPath: iamTokenDir,
				ReadOnly:  true,
			},
		}
	}

	return spec
}
