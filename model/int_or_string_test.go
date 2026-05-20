package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// T4.3 / ADR-002 regression tests for IntOrString.

func TestIntOrString_MarshalAsInt(t *testing.T) {
	b, err := json.Marshal(IntOrStringFromInt(50))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "50" {
		t.Errorf("got %q, want \"50\"", string(b))
	}
}

func TestIntOrString_MarshalAsString(t *testing.T) {
	b, err := json.Marshal(IntOrStringFromString("25%"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"25%"` {
		t.Errorf("got %q, want \"25%%\"", string(b))
	}
}

func TestIntOrString_UnmarshalInt(t *testing.T) {
	var got IntOrString
	if err := json.Unmarshal([]byte("42"), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != IntOrStringInt || got.IntVal != 42 {
		t.Errorf("got %+v, want {Int,42}", got)
	}
}

func TestIntOrString_UnmarshalString(t *testing.T) {
	var got IntOrString
	if err := json.Unmarshal([]byte(`"metrics"`), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != IntOrStringString || got.StrVal != "metrics" {
		t.Errorf("got %+v, want {String,\"metrics\"}", got)
	}
}

func TestIntOrString_UnmarshalRoundTrip(t *testing.T) {
	for _, in := range []IntOrString{
		IntOrStringFromInt(8080),
		IntOrStringFromString("admin"),
		IntOrStringFromInt(0),
		IntOrStringFromString(""),
	} {
		b, _ := json.Marshal(in)
		var back IntOrString
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatalf("unmarshal %q: %v", b, err)
		}
		if back.Type != in.Type || back.IntVal != in.IntVal || back.StrVal != in.StrVal {
			t.Errorf("round-trip lost: %+v → %+v", in, back)
		}
	}
}

func TestIntOrString_UnmarshalRejectsInvalid(t *testing.T) {
	var got IntOrString
	if err := json.Unmarshal([]byte("true"), &got); err == nil {
		t.Fatal("expected error on boolean input")
	}
	if err := json.Unmarshal([]byte(""), &got); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestIntOrString_String(t *testing.T) {
	if s := IntOrStringFromInt(123).String(); s != "123" {
		t.Errorf("got %q", s)
	}
	if s := IntOrStringFromString("foo").String(); s != "foo" {
		t.Errorf("got %q", s)
	}
}

// PodDisruptionBudget serialization with IntOrString.
func TestPodDisruptionBudget_Marshal_IntOrString(t *testing.T) {
	pdb := PodDisruptionBudget{
		MinAvailable: &IntOrString{Type: IntOrStringString, StrVal: "50%"},
	}
	b, err := json.Marshal(pdb)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"minAvailable":"50%"`) {
		t.Errorf("got %q, want \"minAvailable\":\"50%%\"", string(b))
	}
}

func TestHTTPGetAction_PortMarshalsCorrectly(t *testing.T) {
	a := HTTPGetAction{Port: IntOrStringFromInt(8080)}
	b, _ := json.Marshal(a)
	if !strings.Contains(string(b), `"port":8080`) {
		t.Errorf("got %q", string(b))
	}
	a = HTTPGetAction{Port: IntOrStringFromString("metrics")}
	b, _ = json.Marshal(a)
	if !strings.Contains(string(b), `"port":"metrics"`) {
		t.Errorf("got %q", string(b))
	}
}
