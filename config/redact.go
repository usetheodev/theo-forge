package config

import (
	"encoding/json"
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// redactedTokenPlaceholder is the literal string emitted in place of a
// non-empty Token. UnmarshalJSON rejects it (EC-6 from edge-case review).
const redactedTokenPlaceholder = "***"

// String implements fmt.Stringer with Token redacted. Prevents "%+v"
// logging of GlobalConfig from leaking Bearer tokens (T2.6 / SEC-006 / SEC-007).
func (g *GlobalConfig) String() string {
	if g == nil {
		return "<nil>"
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	token := redactedTokenPlaceholder
	if g.token == "" {
		token = "<empty>"
	}
	return fmt.Sprintf("GlobalConfig{Host:%q Token:%s Namespace:%q Image:%q ServiceAccountName:%q ImagePullPolicy:%q VerifySSL:%v}",
		g.host, token, g.namespace, g.image, g.serviceAccountName, g.imagePullPolicy, g.verifySSL)
}

// globalConfigJSON is the on-wire shape — Token redacted on output, rejected on input.
type globalConfigJSON struct {
	Host               string `json:"host,omitempty"`
	Token              string `json:"token,omitempty"`
	Namespace          string `json:"namespace,omitempty"`
	Image              string `json:"image,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	ImagePullPolicy    string `json:"imagePullPolicy,omitempty"`
	VerifySSL          bool   `json:"verifySSL"`
}

// MarshalJSON emits the struct with Token redacted (T2.6 / SEC-006).
func (g *GlobalConfig) MarshalJSON() ([]byte, error) {
	if g == nil {
		return []byte("null"), nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	token := g.token
	if token != "" {
		token = redactedTokenPlaceholder
	}
	return json.Marshal(globalConfigJSON{
		Host:               g.host,
		Token:              token,
		Namespace:          g.namespace,
		Image:              g.image,
		ServiceAccountName: g.serviceAccountName,
		ImagePullPolicy:    string(g.imagePullPolicy),
		VerifySSL:          g.verifySSL,
	})
}

// UnmarshalJSON rejects the redacted Token placeholder (EC-6).
func (g *GlobalConfig) UnmarshalJSON(data []byte) error {
	var raw globalConfigJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Token == redactedTokenPlaceholder {
		return model.ErrRedactedTokenLoaded
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.host = raw.Host
	g.token = raw.Token
	g.namespace = raw.Namespace
	g.image = raw.Image
	g.serviceAccountName = raw.ServiceAccountName
	g.imagePullPolicy = model.ImagePullPolicy(raw.ImagePullPolicy)
	g.verifySSL = raw.VerifySSL
	return nil
}
