package validate

import (
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Phase 7 (T7.4) coverage backfill for validate package.

func TestBinaryUnit_TableDriven(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"1", false}, {"1Ki", false}, {"512Mi", false}, {"2Gi", false}, {"1Ti", false}, {"1Pi", false}, {"1Ei", false},
		{"1.5Gi", false}, {".5Gi", false},
		{"", true}, {"1Xi", true}, {"abc", true}, {"1Mi extra", true}, {"-1Gi", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			err := BinaryUnit(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BinaryUnit(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestDecimalUnit_TableDriven(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"1", false}, {"100m", false}, {"2", false}, {"1k", false}, {"1M", false}, {"1G", false}, {"1T", false}, {"1P", false}, {"1E", false},
		{"1.5", false},
		{"", true}, {"1Xi", true}, {"abc", true}, {"-100m", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			err := DecimalUnit(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecimalUnit(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestConvertBinaryUnit_KnownValues(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"1", 1},
		{"1Ki", 1024},
		{"1Mi", 1024 * 1024},
		{"1Gi", 1024 * 1024 * 1024},
		{"1.5Gi", 1.5 * 1024 * 1024 * 1024},
	}
	for _, tt := range tests {
		got, err := ConvertBinaryUnit(tt.in)
		if err != nil {
			t.Errorf("ConvertBinaryUnit(%q) err: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ConvertBinaryUnit(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
	if _, err := ConvertBinaryUnit("garbage"); err == nil {
		t.Fatal("expected error on invalid input")
	}
}

func TestConvertDecimalUnit_KnownValues(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"1", 1},
		{"100m", 0.1},
		{"1k", 1000},
		{"1M", 1_000_000},
	}
	for _, tt := range tests {
		got, err := ConvertDecimalUnit(tt.in)
		if err != nil {
			t.Errorf("ConvertDecimalUnit(%q) err: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ConvertDecimalUnit(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
	if _, err := ConvertDecimalUnit(""); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestResourceRequirements_AcceptsValid(t *testing.T) {
	r := model.ResourceRequirements{
		Requests: model.ResourceList{CPU: "100m", Memory: "256Mi"},
		Limits:   model.ResourceList{CPU: "500m", Memory: "1Gi"},
	}
	if err := ResourceRequirements(r); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResourceRequirements_RejectsRequestExceedingLimit(t *testing.T) {
	r := model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "2Gi"},
		Limits:   model.ResourceList{Memory: "1Gi"},
	}
	err := ResourceRequirements(r)
	if err == nil {
		t.Fatal("expected error when request exceeds limit")
	}
	if !strings.Contains(err.Error(), "memory") {
		t.Errorf("error should mention memory: %v", err)
	}
}

func TestResourceRequirements_RejectsInvalidUnit(t *testing.T) {
	r := model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "garbage"},
	}
	if err := ResourceRequirements(r); err == nil {
		t.Fatal("expected error on invalid unit")
	}
}
