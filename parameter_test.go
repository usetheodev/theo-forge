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
<<<<<<< HEAD
=======

>>>>>>> c091b749b1cf2a810840f9aa1db0505b8a592dcc
