package model

// Label represents a Prometheus label.
type Label struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

// Counter is a Prometheus counter metric.
type Counter struct {
	Value string `json:"value" yaml:"value"`
}

// Gauge is a Prometheus gauge metric.
type Gauge struct {
	Value    string `json:"value" yaml:"value"`
	Realtime *bool  `json:"realtime,omitempty" yaml:"realtime,omitempty"`
}

// Histogram is a Prometheus histogram metric.
type Histogram struct {
	Value   string    `json:"value" yaml:"value"`
	Buckets []float64 `json:"buckets" yaml:"buckets"`
}

// Metric represents a single Prometheus metric definition.
type Metric struct {
	Name      string     `json:"name" yaml:"name"`
	Help      string     `json:"help" yaml:"help"`
	Labels    []Label    `json:"labels,omitempty" yaml:"labels,omitempty"`
	When      string     `json:"when,omitempty" yaml:"when,omitempty"`
	Counter   *Counter   `json:"counter,omitempty" yaml:"counter,omitempty"`
	Gauge     *Gauge     `json:"gauge,omitempty" yaml:"gauge,omitempty"`
	Histogram *Histogram `json:"histogram,omitempty" yaml:"histogram,omitempty"`
}

// MetricsModel is the serializable Prometheus metrics collection.
type MetricsModel struct {
	Prometheus []Metric `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}
