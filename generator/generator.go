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

	for _, svc := range cfg.Services {
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
		j, err := yaml.Marshal(GenerateJob(job, cfg))
		if err != nil {
			return nil, fmt.Errorf("marshal job %s: %w", job.Name, err)
		}
		docs = append(docs, j)
	}

	return bytes.Join(docs, []byte("---\n")), nil
}
