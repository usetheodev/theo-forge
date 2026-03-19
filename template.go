package forge

import "github.com/usetheo/theo/forge/model"

// Templatable is implemented by types that can build an Argo Template.
type Templatable interface {
	BuildTemplate() (model.TemplateModel, error)
	GetName() string
}
