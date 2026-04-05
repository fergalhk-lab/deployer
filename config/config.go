package config

import (
	"os"

	"sigs.k8s.io/yaml"
)

func FromYAML(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type Config struct {
	Name             string            `json:"name"`
	GeneratedSecrets []GeneratedSecret `json:"generatedSecrets"`
	Services         []Service         `json:"services"`
	InitJobs         []InitJob         `json:"initJobs"`
	CronJobs         []CronJob         `json:"cronJobs"`
}

type GeneratedSecret struct {
	Name    string `json:"name"`
	Length  int    `json:"length"`
	Symbols bool   `json:"symbols"`
}

type Service struct {
	Runnable    `json:",inline"`
	Name        string       `json:"name"`
	Ingress     *Ingress     `json:"ingress"`
	HealthCheck *HealthCheck `json:"healthcheck"`
}

type BaseJob struct {
	Runnable   `json:",inline"`
	Name       string `json:"name"`
	MaxRetries *uint  `json:"maxRetries"`
}

type InitJob struct {
	BaseJob `json:",inline"`
}

type CronJob struct {
	BaseJob  `json:",inline"`
	Schedule string `json:"schedule"`
}

type Runnable struct {
	Image      Image     `json:"image"`
	Resources  Resources `json:"resources"`
	Command    []string  `json:"command"`
	Args       []string  `json:"args"`
	Env        []Env     `json:"env"`
	IAMRoleARN string    `json:"iamRoleARN,omitempty"`
}

type Ingress struct {
	Port   uint16         `json:"port"`
	Public *PublicIngress `json:"public"`
}

type HealthCheck struct {
	Port *uint16 `json:"port,omitempty"` // nil = use ingress port; &0 = explicitly disabled
	Path string  `json:"path,omitempty"` // defaults to /readyz if empty
}

type PublicIngress struct {
	Domain string       `json:"domain"`
	Path   *IngressPath `json:"path,omitempty"`
}

// IngressPath specifies the path match rule for a public ingress.
// Exactly one of Prefix or Literal must be set. If neither is set the
// ingress generator defaults to "/" with PathTypePrefix.
type IngressPath struct {
	Prefix  *string `json:"prefix,omitempty"`
	Literal *string `json:"literal,omitempty"`
}

type Image struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type Resources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Env struct {
	Name                string  `json:"name"`
	RawValue            *string `json:"rawValue"`
	FromGeneratedSecret *string `json:"fromGeneratedSecret"`
}
