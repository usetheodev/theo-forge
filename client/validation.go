package client

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/usetheodev/theo-forge/model"
)

// RFC1123 DNS subdomain (k8s namespace + most resource names): lowercase
// alphanumerics, dashes, and dots; cannot start/end with non-alphanumeric.
// Max length 253. We enforce ≤253 separately.
var k8sNameRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

const k8sNameMaxLen = 253

// validateK8sName rejects names that do not conform to RFC1123 DNS subdomain rules.
// Returns model.ErrInvalidName (wrapped) on failure. (T2.3 / SEC-003).
func validateK8sName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name is empty", model.ErrInvalidName)
	}
	if len(name) > k8sNameMaxLen {
		return fmt.Errorf("%w: name %q exceeds %d chars", model.ErrInvalidName, name, k8sNameMaxLen)
	}
	if !k8sNameRe.MatchString(name) {
		return fmt.Errorf("%w: name %q must be lowercase alphanumerics, dashes, dots (RFC1123)", model.ErrInvalidName, name)
	}
	return nil
}

// validateK8sNamespace rejects namespaces that do not conform to RFC1123 DNS label rules.
// Same rules as name (k8s uses the same regex). Returns model.ErrInvalidNamespace (wrapped).
func validateK8sNamespace(ns string) error {
	if ns == "" {
		return fmt.Errorf("%w: namespace is empty", model.ErrInvalidNamespace)
	}
	if len(ns) > k8sNameMaxLen {
		return fmt.Errorf("%w: namespace %q exceeds %d chars", model.ErrInvalidNamespace, ns, k8sNameMaxLen)
	}
	if !k8sNameRe.MatchString(ns) {
		return fmt.Errorf("%w: namespace %q must be lowercase alphanumerics, dashes, dots (RFC1123)", model.ErrInvalidNamespace, ns)
	}
	return nil
}

// workflowsPath returns "/api/v1/workflows/{escaped-ns}" after validation.
func workflowsPath(ns string) (string, error) {
	if err := validateK8sNamespace(ns); err != nil {
		return "", err
	}
	return "/api/v1/workflows/" + url.PathEscape(ns), nil
}

// workflowPath returns "/api/v1/workflows/{escaped-ns}/{escaped-name}" after validation.
func workflowPath(ns, name string) (string, error) {
	if err := validateK8sNamespace(ns); err != nil {
		return "", err
	}
	if err := validateK8sName(name); err != nil {
		return "", err
	}
	return "/api/v1/workflows/" + url.PathEscape(ns) + "/" + url.PathEscape(name), nil
}

// workflowActionPath returns "/api/v1/workflows/{escaped-ns}/{escaped-name}/{action}".
// action is a fixed SDK-controlled string; not escaped.
func workflowActionPath(ns, name, action string) (string, error) {
	base, err := workflowPath(ns, name)
	if err != nil {
		return "", err
	}
	return base + "/" + action, nil
}
