package models

type VampConfig struct {
	RootPassword          string `yaml:"rootPassword,omitempty" json:"rootPassword,omitempty"`
	DatabaseUrl           string `yaml:"databaseUrl,omitempty" json:"databaseUrl,omitempty"`
	DatabaseName          string `yaml:"databaseName,omitempty" json:"databaseName,omitempty"`
	ImageName             string `yaml:"imageName,omitempty" json:"imageName,omitempty"`
	RepoUsername          string `yaml:"repoUsername,omitempty" json:"repoUsername,omitempty"`
	RepoPassword          string `yaml:"repoPassword,omitempty" json:"repoPassword,omitempty"`
	ImageTag              string `yaml:"imageTag,omitempty" json:"imageTag,omitempty"`
	Mode                  string `yaml:"mode,omitempty" json:"mode,omitempty"`
	AccessTokenExpiration string `yaml:"accessTokenExpiration,omitempty" json:"accessTokenExpiration,omitempty"`
	IstioInstallerImage   string `yaml:"istioInstallerImage,omitempty" json:"istioInstallerImage,omitempty"`
	IstioAdapterImage     string `yaml:"istioAdapterImage,omitempty" json:"istioAdapterImage,omitempty"`
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

type WithSpecification struct {
	Specification map[string]interface{} `json:"specification"`
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
	Port         *int              `json:"port,omitempty"`
	UpdatePeriod *int              `json:"updatePeriod,omitempty"`
	UpdateStep   *int              `json:"updateStep,omitempty"`
	Subset       string            `json:"subset,omitempty"`
	SubsetLabels map[string]string `json:"subsetLabels,omitempty"`
	Policies     []PolicyReference `json:"policies,omitempty"`
}

type PolicyReference struct {
	Name       string            `json:"name,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type Notification struct {
	Text string `json:"text,omitempty"`
}

type ExperimentMetric struct {
	Timestamp         int64   `json:"timestamp"`
	NumberOfElements  int64   `json:"numberOfElements"`
	StandardDeviation float64 `json:"standardDeviation"`
	Average           float64 `json:"average"`
}

type MetricValue struct {
	Timestamp         int64   `yaml:"timestamp" json:"timestamp"`
	NumberOfElements  int64   `yaml:"numberOfElements" json:"numberOfElements"`
	StandardDeviation float64 `yaml:"standardDeviation" json:"standardDeviation"`
	Average           float64 `yaml:"average" json:"average"`
	Sum               float64 `yaml:"sum" json:"sum"`
	Mediam            float64 `yaml:"median" json:"median"`
	Min               float64 `yaml:"min" json:"min"`
	Max               float64 `yaml:"max" json:"max"`
	Rate              float64 `yaml:"rate" json:"rate"`
	P999              float64 `yaml:"p999" json:"p999"`
	P99               float64 `yaml:"p99" json:"p99"`
	P95               float64 `yaml:"p95" json:"p95"`
	P75               float64 `yaml:"p75" json:"p75"`
}

type SubsetToPorts struct {
	Subset string `json:"subset"`
	Ports  []int  `json:"ports"`
}

type LabelsToPortMap struct {
	Map map[string]SubsetToPorts `json:"map"`
}

type DestinationsSubsetsMap struct {
	DestinationsMap map[string]LabelsToPortMap `json:"destinationsMap"`
	Labels          []string                   `json:"labels"`
}
