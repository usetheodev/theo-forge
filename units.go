package forge

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	binaryUnitRe  = regexp.MustCompile(`^([0-9]*\.?[0-9]+)(Ki|Mi|Gi|Ti|Pi|Ei)?$`)
	decimalUnitRe = regexp.MustCompile(`^([0-9]*\.?[0-9]+)(m|k|M|G|T|P|E)?$`)
)

// ValidateBinaryUnit validates a binary resource unit (memory: Ki, Mi, Gi, Ti, Pi, Ei).
func ValidateBinaryUnit(s string) error {
	if !binaryUnitRe.MatchString(s) {
		return fmt.Errorf("invalid binary unit %q: must match <number>[Ki|Mi|Gi|Ti|Pi|Ei]", s)
	}
	return nil
}

// ValidateDecimalUnit validates a decimal resource unit (CPU: m, k, M, G, T, P, E).
func ValidateDecimalUnit(s string) error {
	if !decimalUnitRe.MatchString(s) {
		return fmt.Errorf("invalid decimal unit %q: must match <number>[m|k|M|G|T|P|E]", s)
	}
	return nil
}

// ConvertBinaryUnit converts a binary unit string to its numeric value in base units (bytes).
func ConvertBinaryUnit(s string) (float64, error) {
	if err := ValidateBinaryUnit(s); err != nil {
		return 0, err
	}

	matches := binaryUnitRe.FindStringSubmatch(s)
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse number: %w", err)
	}

	suffix := matches[2]
	multipliers := map[string]float64{
		"":   1,
		"Ki": math.Pow(2, 10),
		"Mi": math.Pow(2, 20),
		"Gi": math.Pow(2, 30),
		"Ti": math.Pow(2, 40),
		"Pi": math.Pow(2, 50),
		"Ei": math.Pow(2, 60),
	}

	mult, ok := multipliers[suffix]
	if !ok {
		return 0, fmt.Errorf("unknown binary suffix %q", suffix)
	}

	return num * mult, nil
}

// ConvertDecimalUnit converts a decimal unit string to its numeric value in base units.
func ConvertDecimalUnit(s string) (float64, error) {
	if err := ValidateDecimalUnit(s); err != nil {
		return 0, err
	}

	matches := decimalUnitRe.FindStringSubmatch(s)
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse number: %w", err)
	}

	suffix := matches[2]
	multipliers := map[string]float64{
		"":  1,
		"m": 0.001,
		"k": 1000,
		"M": 1e6,
		"G": 1e9,
		"T": 1e12,
		"P": 1e15,
		"E": 1e18,
	}

	mult, ok := multipliers[suffix]
	if !ok {
		return 0, fmt.Errorf("unknown decimal suffix %q", suffix)
	}

	return num * mult, nil
}

// ValidateResourceRequirements checks that requests don't exceed limits and values are positive.
func ValidateResourceRequirements(r ResourceRequirements) error {
	// Validate CPU
	if r.Requests.CPU != "" {
		if err := ValidateDecimalUnit(r.Requests.CPU); err != nil {
			return fmt.Errorf("invalid cpu request: %w", err)
		}
		reqVal, _ := ConvertDecimalUnit(r.Requests.CPU)
		if reqVal < 0 {
			return fmt.Errorf("cpu request must be positive")
		}
		if r.Limits.CPU != "" {
			if err := ValidateDecimalUnit(r.Limits.CPU); err != nil {
				return fmt.Errorf("invalid cpu limit: %w", err)
			}
			limVal, _ := ConvertDecimalUnit(r.Limits.CPU)
			if reqVal > limVal {
				return fmt.Errorf("cpu request must be smaller or equal to limit (%s > %s)", r.Requests.CPU, r.Limits.CPU)
			}
		}
	}
	if r.Limits.CPU != "" && r.Requests.CPU == "" {
		if err := ValidateDecimalUnit(r.Limits.CPU); err != nil {
			return fmt.Errorf("invalid cpu limit: %w", err)
		}
	}

	// Validate Memory
	if r.Requests.Memory != "" {
		if err := ValidateBinaryUnit(r.Requests.Memory); err != nil {
			return fmt.Errorf("invalid memory request: %w", err)
		}
		reqVal, _ := ConvertBinaryUnit(r.Requests.Memory)
		if reqVal < 0 {
			return fmt.Errorf("memory request must be positive")
		}
		if r.Limits.Memory != "" {
			if err := ValidateBinaryUnit(r.Limits.Memory); err != nil {
				return fmt.Errorf("invalid memory limit: %w", err)
			}
			limVal, _ := ConvertBinaryUnit(r.Limits.Memory)
			if reqVal > limVal {
				return fmt.Errorf("memory request must be smaller or equal to limit (%s > %s)", r.Requests.Memory, r.Limits.Memory)
			}
		}
	}
	if r.Limits.Memory != "" && r.Requests.Memory == "" {
		if err := ValidateBinaryUnit(r.Limits.Memory); err != nil {
			return fmt.Errorf("invalid memory limit: %w", err)
		}
	}

	// Validate Ephemeral Storage
	if r.Requests.EphemeralStorage != "" {
		if err := ValidateBinaryUnit(r.Requests.EphemeralStorage); err != nil {
			return fmt.Errorf("invalid ephemeral storage request: %w", err)
		}
		if r.Limits.EphemeralStorage != "" {
			if err := ValidateBinaryUnit(r.Limits.EphemeralStorage); err != nil {
				return fmt.Errorf("invalid ephemeral storage limit: %w", err)
			}
			reqVal, _ := ConvertBinaryUnit(r.Requests.EphemeralStorage)
			limVal, _ := ConvertBinaryUnit(r.Limits.EphemeralStorage)
			if reqVal > limVal {
				return fmt.Errorf("ephemeral storage request must be smaller or equal to limit (%s > %s)",
					r.Requests.EphemeralStorage, r.Limits.EphemeralStorage)
			}
		}
	}
	if r.Limits.EphemeralStorage != "" && r.Requests.EphemeralStorage == "" {
		if err := ValidateBinaryUnit(r.Limits.EphemeralStorage); err != nil {
			return fmt.Errorf("invalid ephemeral storage limit: %w", err)
		}
	}

	return nil
}

// parsePositiveDecimal checks that a decimal CPU value is positive.
func parsePositiveDecimal(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	val, err := ConvertDecimalUnit(s)
	if err != nil {
		return err
	}
	if val < 0 {
		return fmt.Errorf("value must be positive, got %s", s)
	}
	return nil
}
