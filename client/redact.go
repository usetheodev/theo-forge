package client

import (
	"encoding/json"
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// redactedTokenPlaceholder is the literal string emitted by Stringer/Marshal
// in place of a non-empty Token field. UnmarshalJSON rejects this literal
// to prevent silent auth loss when struct dumps round-trip through persistence
// (EC-6 from edge-case review).
const redactedTokenPlaceholder = "***"

// String implements fmt.Stringer with Token redacted to "***" (or "<empty>"
// when Token is unset). Prevents "%+v" / "%v" logging from leaking Bearer
// tokens (T2.6 / SEC-006).
func (s *WorkflowsService) String() string {
	if s == nil {
		return "<nil>"
	}
	token := redactedTokenPlaceholder
	if s.Token == "" {
		token = "<empty>"
	}
	return fmt.Sprintf("WorkflowsService{Host:%q Namespace:%q Token:%s VerifySSL:%v}",
		s.Host, s.Namespace, token, s.VerifySSL)
}

// workflowsServiceJSON is the on-wire shape used by Marshal/Unmarshal.
// The Token field is redacted on output and rejected on input when it
// matches the redacted placeholder.
type workflowsServiceJSON struct {
	Host      string `json:"host,omitempty"`
	Token     string `json:"token,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	VerifySSL bool   `json:"verifySSL"`
}

// MarshalJSON emits the struct with Token redacted (T2.6 / SEC-006).
// WorkflowsService is NOT safe for round-trip persistence — consumers
// who need to persist auth state must serialize their own struct.
func (s *WorkflowsService) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	token := s.Token
	if token != "" {
		token = redactedTokenPlaceholder
	}
	return json.Marshal(workflowsServiceJSON{
		Host:      s.Host,
		Token:     token,
		Namespace: s.Namespace,
		VerifySSL: s.VerifySSL,
	})
}

// UnmarshalJSON rejects the redacted Token placeholder to prevent silent
// auth loss when struct dumps are persisted and reloaded (EC-6).
func (s *WorkflowsService) UnmarshalJSON(data []byte) error {
	var raw workflowsServiceJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Token == redactedTokenPlaceholder {
		return model.ErrRedactedTokenLoaded
	}
	s.Host = raw.Host
	s.Token = raw.Token
	s.Namespace = raw.Namespace
	s.VerifySSL = raw.VerifySSL
	return nil
}
