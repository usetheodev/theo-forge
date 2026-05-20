package forge

import (
	"encoding/json"
	"testing"
)

func TestParameterNoValueFailsToString(t *testing.T) {
	p := Parameter{Name: "my_name", Enum: []string{"1", "2", "3"}}
	_, err := p.String()
	if err == nil {
		t.Fatal("expected error when value is not set")
	}
	if err.Error() != "cannot represent Parameter as string: value is not set" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestParameterNoNameCanBeCreated(t *testing.T) {
	p := Parameter{Value: ptrStr("3"), Enum: []string{"1", "2", "3"}}
	if p.Value == nil {
		t.Fatal("expected value to be set")
	}
}

func TestParameterNoNameFailsAsInput(t *testing.T) {
	p := Parameter{Value: ptrStr("3")}
	_, err := p.AsInput()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
	if err.Error() != "name cannot be empty when used" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestParameterNoNameFailsAsArgument(t *testing.T) {
	p := Parameter{Value: ptrStr("3")}
	_, err := p.AsArgument()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestParameterNoNameFailsAsOutput(t *testing.T) {
	p := Parameter{Value: ptrStr("3")}
	_, err := p.AsOutput()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestParameterWithName(t *testing.T) {
	p := Parameter{Value: ptrStr("hello")}
	p2 := p.WithName("test")
	if p2.Name != "test" {
		t.Fatalf("expected name 'test', got '%s'", p2.Name)
	}
	if *p2.Value != "hello" {
		t.Fatalf("expected value 'hello', got '%s'", *p2.Value)
	}
	// Original unchanged
	if p.Name != "" {
		t.Fatal("original parameter name should remain empty")
	}
}

func TestParameterValuesSerialization(t *testing.T) {
	tests := []struct {
		name     string
		param    Parameter
		wantName string
		wantVal  string
		wantDef  string
		wantEnum []string
	}{
		{
			name:     "string value",
			param:    Parameter{Name: "test", Value: ptrStr("hello"), Default: ptrStr("world")},
			wantName: "test",
			wantVal:  "hello",
			wantDef:  "world",
		},
		{
			name:     "integer value",
			param:    Parameter{Name: "test", Value: ptrStr("1"), Default: ptrStr("2"), Enum: []string{"1", "2"}},
			wantName: "test",
			wantVal:  "1",
			wantDef:  "2",
			wantEnum: []string{"1", "2"},
		},
		{
			name:     "boolean value",
			param:    Parameter{Name: "test", Value: ptrStr("true")},
			wantName: "test",
			wantVal:  "true",
		},
		{
			name:     "null value",
			param:    Parameter{Name: "test", Value: ptrStr("null")},
			wantName: "test",
			wantVal:  "null",
		},
		{
			name:     "json object value",
			param:    Parameter{Name: "test", Value: ptrStr(`{"key":"val"}`)},
			wantName: "test",
			wantVal:  `{"key":"val"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := tt.param.AsInput()
			if err != nil {
				t.Fatalf("AsInput() error: %v", err)
			}
			if input.Name != tt.wantName {
				t.Errorf("name = %q, want %q", input.Name, tt.wantName)
			}
			if tt.wantVal != "" {
				if input.Value == nil || *input.Value != tt.wantVal {
					t.Errorf("value = %v, want %q", input.Value, tt.wantVal)
				}
			}
			if tt.wantDef != "" {
				if input.Default == nil || *input.Default != tt.wantDef {
					t.Errorf("default = %v, want %q", input.Default, tt.wantDef)
				}
			}
			if tt.wantEnum != nil {
				if len(input.Enum) != len(tt.wantEnum) {
					t.Errorf("enum length = %d, want %d", len(input.Enum), len(tt.wantEnum))
				}
			}
		})
	}
}

func TestParameterAsArgumentExcludesDefault(t *testing.T) {
	p := Parameter{Name: "test", Value: ptrStr("hello"), Default: ptrStr("world")}
	arg, err := p.AsArgument()
	if err != nil {
		t.Fatalf("AsArgument() error: %v", err)
	}
	if arg.Name != "test" {
		t.Errorf("name = %q, want 'test'", arg.Name)
	}
	if arg.Value == nil || *arg.Value != "hello" {
		t.Errorf("value = %v, want 'hello'", arg.Value)
	}
	if arg.Default != nil {
		t.Error("AsArgument should not include default")
	}
}

func TestParameterAsOutputOnlyValueAndValueFrom(t *testing.T) {
	p := Parameter{
		Name:        "test",
		Value:       ptrStr("hello"),
		Default:     ptrStr("world"),
		Description: "desc",
		GlobalName:  "global",
		ValueFrom:   &ValueFrom{Path: "/tmp/out"},
	}
	out, err := p.AsOutput()
	if err != nil {
		t.Fatalf("AsOutput() error: %v", err)
	}
	if out.Name != "test" {
		t.Errorf("name = %q, want 'test'", out.Name)
	}
	if out.ValueFrom == nil {
		t.Error("expected value_from to be set")
	}
	if out.GlobalName != "global" {
		t.Errorf("global_name = %q, want 'global'", out.GlobalName)
	}
}

func TestParameterModelJSON(t *testing.T) {
	p := Parameter{Name: "msg", Value: ptrStr("hello")}
	model, err := p.AsInput()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "msg" {
		t.Errorf("json name = %v, want 'msg'", m["name"])
	}
}

// --- NewRetryOnFailure (Phase 1 / T1.1) ---

func TestNewRetryOnFailure_PopulatesAllFields(t *testing.T) {
	rs := NewRetryOnFailure(2, "10s", "2", "120s")
	if rs == nil {
		t.Fatal("NewRetryOnFailure returned nil")
	}
	if rs.Limit == nil || *rs.Limit != 2 {
		t.Errorf("Limit = %v, want pointer to 2", rs.Limit)
	}
	if rs.RetryPolicy != RetryPolicy("OnFailure") {
		t.Errorf("RetryPolicy = %q, want OnFailure", rs.RetryPolicy)
	}
	if rs.Backoff == nil {
		t.Fatal("Backoff is nil")
	}
	if rs.Backoff.Duration != "10s" {
		t.Errorf("Backoff.Duration = %q, want 10s", rs.Backoff.Duration)
	}
	if rs.Backoff.Factor != "2" {
		t.Errorf("Backoff.Factor = %v, want \"2\"", rs.Backoff.Factor)
	}
	if rs.Backoff.MaxDuration != "120s" {
		t.Errorf("Backoff.MaxDuration = %q, want 120s", rs.Backoff.MaxDuration)
	}
}

func TestNewRetryOnFailure_SkipOOMExpression(t *testing.T) {
	rs := NewRetryOnFailure(3, "5s", "2", "60s")
	const want = `asInt(lastRetry.exitCode) != 137`
	if rs.Expression != want {
		t.Errorf("Expression = %q, want %q", rs.Expression, want)
	}
}

func TestNewRetryOnFailure_BuildProducesValidModel(t *testing.T) {
	rs := NewRetryOnFailure(2, "10s", "2", "120s")
	m := rs.Build()
	if m.Limit != "2" {
		t.Errorf("model.Limit = %v, want \"2\"", m.Limit)
	}
	if m.RetryPolicy != "OnFailure" {
		t.Errorf("model.RetryPolicy = %q, want OnFailure", m.RetryPolicy)
	}
	if m.Backoff == nil {
		t.Fatal("model.Backoff is nil")
	}
	if m.Backoff.Factor != "2" {
		t.Errorf("model.Backoff.Factor = %v, want \"2\"", m.Backoff.Factor)
	}
	if m.Expression != `asInt(lastRetry.exitCode) != 137` {
		t.Errorf("model.Expression mismatch: %q", m.Expression)
	}
}

func TestNewRetryOnFailure_TableDriven(t *testing.T) {
	cases := []struct {
		name  string
		limit int
	}{
		{"zero-limit", 0},
		{"two-limit", 2},
		{"ten-limit", 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rs := NewRetryOnFailure(tc.limit, "1s", "2", "30s")
			if rs.Limit == nil || *rs.Limit != tc.limit {
				t.Errorf("Limit = %v, want %d", rs.Limit, tc.limit)
			}
		})
	}
}

// --- Backoff.Factor coverage backfill (Phase 1 / T1.1 EC-4) ---
// review.db code-p4-backoff-factor-partial-normalization

func TestBackoffFactor_Float64Normalization(t *testing.T) {
	rs := RetryStrategy{
		Limit:   Ptr(2),
		Backoff: &Backoff{Duration: "5s", Factor: 1.5, MaxDuration: "30s"},
	}
	m := rs.Build()
	got, ok := m.Backoff.Factor.(string)
	if !ok {
		t.Fatalf("Factor not normalized to string: %T %v", m.Backoff.Factor, m.Backoff.Factor)
	}
	if got != "1.5" {
		t.Errorf("Factor = %q, want \"1.5\"", got)
	}
}

func TestBackoffFactor_PointerFloat64(t *testing.T) {
	f := 2.5
	rs := RetryStrategy{
		Limit:   Ptr(2),
		Backoff: &Backoff{Duration: "5s", Factor: &f, MaxDuration: "30s"},
	}
	m := rs.Build()
	got, ok := m.Backoff.Factor.(string)
	if !ok {
		t.Fatalf("Factor not normalized to string: %T %v", m.Backoff.Factor, m.Backoff.Factor)
	}
	if got != "2.5" {
		t.Errorf("Factor = %q, want \"2.5\"", got)
	}
}
