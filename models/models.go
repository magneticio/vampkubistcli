package models

type VampConfig struct {
	RootPassword string `yaml:"rootPassword,omitempty" json:"rootPassword,omitempty"`
	DatabaseUrl  string `yaml:"databaseUrl,omitempty" json:"databaseUrl,omitempty"`
	DatabaseName string `yaml:"databaseName,omitempty" json:"databaseName,omitempty"`
	ImageName    string `yaml:"imageName,omitempty" json:"imageName,omitempty"`
	RepoUsername string `yaml:"repoUsername,omitempty" json:"repoUsername,omitempty"`
	RepoPassword string `yaml:"repoPassword,omitempty" json:"repoPassword,omitempty"`
	ImageTag     string `yaml:"imageTag,omitempty" json:"imageTag,omitempty"`
	Mode         string `yaml:"mode,omitempty" json:"mode,omitempty"`
}

type ErrorResponse struct {
	Message           string            `json:"message"`
	ValidationOutcome []ValidationError `json:"validationOutcome"`
}

type ValidationError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type Named struct {
	Name string `json:"name"`
}

type Versioned struct {
	Version string `json:"version"`
}

type Metadata struct {
	Metadata map[string]string `json:"metadata"`
}

type Permission struct {
	Read       bool `json:"read"`
	Write      bool `json:"write"`
	Delete     bool `json:"delete"`
	EditAccess bool `json:"editAccess"`
}

type VampService struct {
	Gateways         []string `json:"gateways"`
	Hosts            []string `json:"hosts"`
	Routes           []Route  `json:"routes"`
	ExposeInternally bool     `json:"exposeInternally"`
}

type Route struct {
	Protocol  string   `json:"protocol"`
	Condition string   `json:"condition,omitempty"`
	Rewrite   string   `json:"rewrite,omitempty"`
	Weights   []Weight `json:"weights"`
}

type Weight struct {
	Destination string `json:"destination"`
	Port        int64  `json:"port"`
	Version     string `json:"version"`
	Weight      int64  `json:"weight"`
}

type CanaryRelease struct {
	VampService  string            `json:"vampService"`
	Destination  string            `json:"destination,omitempty"`
	Port         string            `json:"port,omitempty"`
	Subset       string            `json:"subset,omitempty"`
	SubsetLabels map[string]string `json:"subsetLabels,omitempty"`
}
