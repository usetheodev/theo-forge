package validate

import (
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Coverage backfill round 2: hit every branch of ResourceRequirements +
// error paths in ConvertBinaryUnit/ConvertDecimalUnit.

func TestConvertBinaryUnit_RejectsInvalid(t *testing.T) {
	if _, err := ConvertBinaryUnit("garbage"); err == nil {
		t.Fatal("expected error on invalid binary unit")
	}
	if _, err := ConvertBinaryUnit(""); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestConvertDecimalUnit_RejectsInvalid(t *testing.T) {
	if _, err := ConvertDecimalUnit("garbage"); err == nil {
		t.Fatal("expected error on invalid decimal unit")
	}
	if _, err := ConvertDecimalUnit(""); err == nil {
		t.Fatal("expected error on empty input")
	}
}

// ResourceRequirements has 3 resources × 4 code paths each:
//  1. request set + limit set, all valid
//  2. request set + limit set, request > limit (error)
//  3. request set, invalid format (error)
//  4. only limit set, valid
//  5. only limit set, invalid format (error)

func TestResourceRequirements_AllPaths_CPU(t *testing.T) {
	// 1. request + limit valid
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{CPU: "100m"},
		Limits:   model.ResourceList{CPU: "500m"},
	}); err != nil {
		t.Errorf("happy path: %v", err)
	}
	// 2. request > limit
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{CPU: "1"},
		Limits:   model.ResourceList{CPU: "500m"},
	}); err == nil {
		t.Errorf("expected error on request > limit")
	}
	// 3. invalid request format
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{CPU: "garbage"},
	}); err == nil {
		t.Errorf("expected error on invalid request format")
	}
	// 4. invalid limit format (with valid request)
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{CPU: "100m"},
		Limits:   model.ResourceList{CPU: "garbage"},
	}); err == nil {
		t.Errorf("expected error on invalid limit format")
	}
	// 5. only limit set, valid
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{CPU: "500m"},
	}); err != nil {
		t.Errorf("limit-only valid path: %v", err)
	}
	// 6. only limit set, invalid
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{CPU: "garbage"},
	}); err == nil {
		t.Errorf("expected error on invalid limit-only")
	}
}

func TestResourceRequirements_AllPaths_Memory(t *testing.T) {
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "256Mi"},
		Limits:   model.ResourceList{Memory: "1Gi"},
	}); err != nil {
		t.Errorf("happy: %v", err)
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "2Gi"},
		Limits:   model.ResourceList{Memory: "1Gi"},
	}); err == nil || !strings.Contains(err.Error(), "memory") {
		t.Errorf("expected memory request>limit error, got %v", err)
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid memory request format")
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{Memory: "256Mi"},
		Limits:   model.ResourceList{Memory: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid memory limit format")
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{Memory: "1Gi"},
	}); err != nil {
		t.Errorf("memory limit-only valid: %v", err)
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{Memory: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid memory limit-only")
	}
}

func TestResourceRequirements_AllPaths_EphemeralStorage(t *testing.T) {
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{EphemeralStorage: "1Gi"},
		Limits:   model.ResourceList{EphemeralStorage: "5Gi"},
	}); err != nil {
		t.Errorf("ephemeral happy: %v", err)
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{EphemeralStorage: "10Gi"},
		Limits:   model.ResourceList{EphemeralStorage: "1Gi"},
	}); err == nil {
		t.Errorf("expected ephemeral request>limit error")
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{EphemeralStorage: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid ephemeral request format")
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Requests: model.ResourceList{EphemeralStorage: "1Gi"},
		Limits:   model.ResourceList{EphemeralStorage: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid ephemeral limit format")
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{EphemeralStorage: "1Gi"},
	}); err != nil {
		t.Errorf("ephemeral limit-only valid: %v", err)
	}
	if err := ResourceRequirements(model.ResourceRequirements{
		Limits: model.ResourceList{EphemeralStorage: "garbage"},
	}); err == nil {
		t.Errorf("expected invalid ephemeral limit-only")
	}
}
