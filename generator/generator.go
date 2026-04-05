package generator

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/fergalhk-lab/deployer/config"
)

// Generate converts a Config into a single multi-document Kubernetes YAML manifest.
func Generate(cfg config.Config, opts Options) ([]byte, error) {
	var docs [][]byte

	ns, err := yaml.Marshal(GenerateNamespace(cfg))
	if err != nil {
		return nil, fmt.Errorf("marshal namespace: %w", err)
	}
	docs = append(docs, ns)

	for _, gs := range cfg.GeneratedSecrets {
		pg, err := yaml.Marshal(GeneratePasswordGenerator(gs, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal password generator %s: %w", gs.Name, err)
		}
		docs = append(docs, pg)

		es, err := marshalExternalSecret(GenerateExternalSecret(gs, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal external secret %s: %w", gs.Name, err)
		}
		docs = append(docs, es)
	}

	for _, svc := range cfg.Services {
		sa, err := yaml.Marshal(GenerateServiceAccount(svc.Name, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal service account %s: %w", svc.Name, err)
		}
		docs = append(docs, sa)

		d, err := yaml.Marshal(GenerateDeployment(svc, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal deployment %s: %w", svc.Name, err)
		}
		docs = append(docs, d)

		if svc.Ingress != nil {
			s, err := yaml.Marshal(GenerateService(svc, cfg))
			if err != nil {
				return nil, fmt.Errorf("marshal service %s: %w", svc.Name, err)
			}
			docs = append(docs, s)

			if svc.Ingress.Public != nil {
				ing, err := yaml.Marshal(GenerateIngress(svc, cfg, opts))
				if err != nil {
					return nil, fmt.Errorf("marshal ingress %s: %w", svc.Name, err)
				}
				docs = append(docs, ing)
			}
		}
	}

	for _, job := range cfg.InitJobs {
		sa, err := yaml.Marshal(GenerateServiceAccount(job.Name, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal service account %s: %w", job.Name, err)
		}
		docs = append(docs, sa)

		j, err := yaml.Marshal(GenerateJob(job, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal job %s: %w", job.Name, err)
		}
		docs = append(docs, j)
	}

	for _, cj := range cfg.CronJobs {
		sa, err := yaml.Marshal(GenerateServiceAccount(cj.Name, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal service account %s: %w", cj.Name, err)
		}
		docs = append(docs, sa)

		c, err := yaml.Marshal(GenerateCronJob(cj, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal cronjob %s: %w", cj.Name, err)
		}
		docs = append(docs, c)
	}

	return bytes.Join(docs, []byte("---\n")), nil
}
