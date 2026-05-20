package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// IntOrString is a Go-side mirror of k8s.io/apimachinery's intstr.IntOrString.
// We mirror locally to avoid pulling the k8s.io transitive tree (~50 MiB)
// per .claude/rules/dependencies.md.
//
// Used for Argo fields that accept EITHER an integer OR a string (commonly
// a port name or a percentage). Replaces interface{} on the previous
// PodDisruptionBudget.MinAvailable/MaxUnavailable, HTTPGetAction.Port,
// and TCPSocketAction.Port. (T4.3 / ADR-002)
//
// Construction:
//
//	model.IntOrStringFromInt(50)            // serializes as 50
//	model.IntOrStringFromString("25%")      // serializes as "25%"
//	model.IntOrStringFromString("metrics")  // serializes as "metrics"
type IntOrString struct {
	Type   IntOrStringType
	IntVal int32
	StrVal string
}

// IntOrStringType discriminates the IntOrString variant.
type IntOrStringType int

// IntOrStringType literals.
const (
	IntOrStringInt    IntOrStringType = 0
	IntOrStringString IntOrStringType = 1
)

// IntOrStringFromInt constructs an IntOrString from an int32.
func IntOrStringFromInt(n int32) IntOrString {
	return IntOrString{Type: IntOrStringInt, IntVal: n}
}

// IntOrStringFromString constructs an IntOrString from a string.
func IntOrStringFromString(s string) IntOrString {
	return IntOrString{Type: IntOrStringString, StrVal: s}
}

// String returns the wire representation: the int as decimal OR the string verbatim.
func (i IntOrString) String() string {
	if i.Type == IntOrStringString {
		return i.StrVal
	}
	return strconv.FormatInt(int64(i.IntVal), 10)
}

// MarshalJSON emits a bare integer or a JSON string depending on Type.
func (i IntOrString) MarshalJSON() ([]byte, error) {
	if i.Type == IntOrStringString {
		return json.Marshal(i.StrVal)
	}
	return json.Marshal(i.IntVal)
}

// UnmarshalJSON accepts either a JSON number (int) or a JSON string.
func (i *IntOrString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("IntOrString: empty input")
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("IntOrString: unmarshal string: %w", err)
		}
		i.Type = IntOrStringString
		i.StrVal = s
		i.IntVal = 0
		return nil
	}
	var n int32
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("IntOrString: unmarshal int32: %w", err)
	}
	i.Type = IntOrStringInt
	i.IntVal = n
	i.StrVal = ""
	return nil
}
