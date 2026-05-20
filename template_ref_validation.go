package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// validateTemplateReference enforces Argo's "exactly one of Template, TemplateRef, Inline"
// rule for DAG tasks and workflow steps. (T3.3 / code-p4-template-ref-mutual-exclusion)
//
// who is a short prefix included in error messages, e.g. "task fan-out" or "step deploy".
func validateTemplateReference(template string, templateRef *model.TemplateRef, inline Templatable, who string) error {
	set := 0
	if template != "" {
		set++
	}
	if templateRef != nil {
		set++
	}
	if inline != nil {
		set++
	}
	switch set {
	case 0:
		return fmt.Errorf("%s: %w", who, model.ErrTemplateMissing)
	case 1:
		return nil
	default:
		return fmt.Errorf("%s: %w (got %d set)", who, model.ErrTemplateAmbiguous, set)
	}
}
