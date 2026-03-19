package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
)

// RetryStrategy configures retry behavior for templates.
type RetryStrategy struct {
	Limit       *int
	RetryPolicy RetryPolicy
	Backoff     *Backoff
	Expression  string
}

// Build converts RetryStrategy to its serializable model.
func (r RetryStrategy) Build() model.RetryStrategyModel {
	var limit interface{}
	if r.Limit != nil {
		limit = fmt.Sprintf("%d", *r.Limit)
	}
	var backoff *model.Backoff
	if r.Backoff != nil {
		b := *r.Backoff
		// Normalize factor to string for Argo compatibility
		if factor, ok := b.Factor.(*int); ok && factor != nil {
			b.Factor = fmt.Sprintf("%d", *factor)
		} else if factor, ok := b.Factor.(int); ok {
			b.Factor = fmt.Sprintf("%d", factor)
		}
		backoff = &b
	}
	return model.RetryStrategyModel{
		Limit:       limit,
		RetryPolicy: string(r.RetryPolicy),
		Backoff:     backoff,
		Expression:  r.Expression,
	}
}
