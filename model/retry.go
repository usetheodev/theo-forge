package model

// RetryPolicy defines when to retry a step/task.
type RetryPolicy string

const (
	RetryAlways           RetryPolicy = "Always"
	RetryOnFailure        RetryPolicy = "OnFailure"
	RetryOnError          RetryPolicy = "OnError"
	RetryOnTransientError RetryPolicy = "OnTransientError"
)

// Backoff defines the backoff strategy for retries.
type Backoff struct {
	Duration    string `json:"duration,omitempty" yaml:"duration,omitempty"`
	Factor      *int   `json:"factor,omitempty" yaml:"factor,omitempty"`
	MaxDuration string `json:"maxDuration,omitempty" yaml:"maxDuration,omitempty"`
}

// RetryStrategyModel is the serializable representation of a retry strategy.
type RetryStrategyModel struct {
	Limit       *int   `json:"limit,omitempty" yaml:"limit,omitempty"`
	RetryPolicy string `json:"retryPolicy,omitempty" yaml:"retryPolicy,omitempty"`
	Backoff     *Backoff `json:"backoff,omitempty" yaml:"backoff,omitempty"`
	Expression  string `json:"expression,omitempty" yaml:"expression,omitempty"`
}
