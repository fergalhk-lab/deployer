package config

type Config struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
	InitJobs []InitJob `json:"initJobs"`
}

type Service struct {
	Runnable `json:",inline"`

	Name    string   `json:"name"`
	Ingress *Ingress `json:"ingress"`
}

type InitJob struct {
	Runnable `json:",inline"`

	Name       string `json:"name"`
	MaxRetries *uint  `json:"maxRetries"`
}

type Runnable struct {
	Image     Image     `json:"image"`
	Resources Resources `json:"resources"`
	Command   []string  `json:"command"`
	Args      []string  `json:"args"`
	Env       []Env     `json:"env"`
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
