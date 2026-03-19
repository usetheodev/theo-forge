package forge

import "github.com/usetheo/theo/forge/model"

// RetryStrategy configures retry behavior for templates.
type RetryStrategy struct {
	Limit       *int
	RetryPolicy RetryPolicy
	Backoff     *Backoff
	Expression  string
}

// Build converts RetryStrategy to its serializable model.
func (r RetryStrategy) Build() model.RetryStrategyModel {
	return model.RetryStrategyModel{
		Limit:       r.Limit,
		RetryPolicy: string(r.RetryPolicy),
		Backoff:     r.Backoff,
		Expression:  r.Expression,
	}
}
