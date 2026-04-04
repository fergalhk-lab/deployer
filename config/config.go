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
	Name     string    `json:"name"`
	Services []Service `json:"services"`
	InitJobs []InitJob `json:"initJobs"`
	CronJobs []CronJob `json:"cronJobs"`
}

type Service struct {
	Runnable `json:",inline"`

	Name    string   `json:"name"`
	Ingress *Ingress `json:"ingress"`
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

type PublicIngress struct {
	Domain string `json:"domain"`
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
	Name     string  `json:"name"`
	RawValue *string `json:"rawValue"`
	// TODO - implement persistentRandom
}
