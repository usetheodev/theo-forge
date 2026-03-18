package forge

import (
	"math"
	"testing"
)

func TestValidateBinaryUnit(t *testing.T) {
	valid := []string{"500Ki", "1Mi", "2Gi", "1Ti", "1.5Pi", "1.5Ei", "42", "0.5"}
	for _, v := range valid {
		if err := ValidateBinaryUnit(v); err != nil {
			t.Errorf("expected valid: %q, got error: %v", v, err)
		}
	}

	invalid := []string{"Mi", "5K", "Ti", "abc", "1.5Z", "500m", "2k"}
	for _, v := range invalid {
		if err := ValidateBinaryUnit(v); err == nil {
			t.Errorf("expected invalid: %q", v)
		}
	}
}

func TestValidateDecimalUnit(t *testing.T) {
	valid := []string{"0.5", "1", "500m", "2k", "1.5M", "42"}
	for _, v := range valid {
		if err := ValidateDecimalUnit(v); err != nil {
			t.Errorf("expected valid: %q, got error: %v", v, err)
		}
	}

	invalid := []string{"abc", "K", "2e", "1.5Z", "1.5Ki", "1.5Mi"}
	for _, v := range invalid {
		if err := ValidateDecimalUnit(v); err == nil {
			t.Errorf("expected invalid: %q", v)
		}
	}
}

func TestConvertDecimalUnit(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"500m", 0.5},
		{"2k", 2000.0},
		{"1.5M", 1500000.0},
		{"42", 42.0},
		{"1", 1.0},
		{"0.5", 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ConvertDecimalUnit(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestConvertDecimalUnitInvalid(t *testing.T) {
	invalid := []string{"1.5Z", "abc", "1.5Ki", "1.5Mi"}
	for _, v := range invalid {
		_, err := ConvertDecimalUnit(v)
		if err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestConvertBinaryUnit(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"500Ki", 512000.0},
		{"1Mi", 1048576.0},
		{"2Gi", 2147483648.0},
		{"42", 42.0},
		{"0.5", 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ConvertBinaryUnit(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestConvertBinaryUnitInvalid(t *testing.T) {
	invalid := []string{"1.5Z", "abc", "500m", "2k"}
	for _, v := range invalid {
		_, err := ConvertBinaryUnit(v)
		if err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestValidateResourceRequirementsValid(t *testing.T) {
	tests := []struct {
		name string
		res  ResourceRequirements
	}{
		{"cpu only", ResourceRequirements{
			Requests: ResourceList{CPU: "500m"},
			Limits:   ResourceList{CPU: "1"},
		}},
		{"memory only", ResourceRequirements{
			Requests: ResourceList{Memory: "256Mi"},
			Limits:   ResourceList{Memory: "1Gi"},
		}},
		{"cpu and memory", ResourceRequirements{
			Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
			Limits:   ResourceList{CPU: "500m", Memory: "512Mi"},
		}},
		{"equal request and limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1"},
			Limits:   ResourceList{CPU: "1"},
		}},
		{"request only", ResourceRequirements{
			Requests: ResourceList{CPU: "500m"},
		}},
		{"limit only", ResourceRequirements{
			Limits: ResourceList{CPU: "1"},
		}},
		{"ephemeral storage", ResourceRequirements{
			Requests: ResourceList{EphemeralStorage: "1Gi"},
			Limits:   ResourceList{EphemeralStorage: "50Gi"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateResourceRequirements(tt.res); err != nil {
				t.Errorf("expected valid, got: %v", err)
			}
		})
	}
}

func TestValidateResourceRequirementsInvalid(t *testing.T) {
	tests := []struct {
		name string
		res  ResourceRequirements
		msg  string
	}{
		{"cpu request > limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1"},
			Limits:   ResourceList{CPU: "500m"},
		}, "request must be smaller or equal to limit"},
		{"cpu millicores request > limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1000m"},
			Limits:   ResourceList{CPU: "800m"},
		}, "request must be smaller or equal to limit"},
		{"memory request > limit", ResourceRequirements{
			Requests: ResourceList{Memory: "1Gi"},
			Limits:   ResourceList{Memory: "512Mi"},
		}, "request must be smaller or equal to limit"},
		{"ephemeral request > limit", ResourceRequirements{
			Requests: ResourceList{EphemeralStorage: "100Gi"},
			Limits:   ResourceList{EphemeralStorage: "50Gi"},
		}, "request must be smaller or equal to limit"},
		{"invalid cpu format", ResourceRequirements{
			Requests: ResourceList{CPU: "500a"},
		}, "invalid"},
		{"invalid memory format", ResourceRequirements{
			Requests: ResourceList{Memory: "500m"},
		}, "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceRequirements(tt.res)
			if err == nil {
				t.Fatal("expected error")
			}
			if !contains(err.Error(), tt.msg) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.msg)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
