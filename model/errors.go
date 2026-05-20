package model

import "errors"

// Sentinel errors for the forge SDK.
//
// All public-package errors are prefixed with the sub-package name
// ("forge:", "serialize:", "client:", "expr:") for diagnostic clarity.
// Callers can use errors.Is to test category.
var (
	// ErrPathTraversal is returned when a file name escapes the intended output directory.
	ErrPathTraversal = errors.New("serialize: file name escapes output directory")

	// ErrEmptyImage is returned when a Container or Script BuildTemplate is called with Image=="".
	ErrEmptyImage = errors.New("forge: container image is required")

	// ErrEntrypointMissing is returned when a Workflow has templates but no Entrypoint.
	ErrEntrypointMissing = errors.New("forge: workflow has templates but no Entrypoint")

	// ErrTemplateAmbiguous is returned when a DAG Task or Step sets more than one of
	// Template, TemplateRef, Inline.
	ErrTemplateAmbiguous = errors.New("forge: exactly one of Template/TemplateRef/Inline must be set")

	// ErrTemplateMissing is returned when a DAG Task or Step sets none of
	// Template, TemplateRef, Inline.
	ErrTemplateMissing = errors.New("forge: one of Template/TemplateRef/Inline must be set")

	// ErrInvalidName is returned when a k8s resource name does not conform to RFC1123.
	ErrInvalidName = errors.New("client: invalid k8s resource name")

	// ErrInvalidNamespace is returned when a k8s namespace does not conform to RFC1123.
	ErrInvalidNamespace = errors.New("client: invalid k8s namespace")

	// ErrResponseTooLarge is returned when an HTTP response body exceeds the configured limit.
	ErrResponseTooLarge = errors.New("client: response body exceeds maximum allowed size")

	// ErrRedactedTokenLoaded is returned when json.Unmarshal observes the literal "***"
	// in a Token field — indicating the consumer is round-tripping a struct dump that
	// has the token redacted. Persisting such a struct silently loses auth.
	ErrRedactedTokenLoaded = errors.New("forge: refused to load redacted token placeholder \"***\" (see Stringer doc)")

	// ErrInvalidTimeout is returned when a duration string fails to parse.
	ErrInvalidTimeout = errors.New("forge: invalid timeout duration")
)
