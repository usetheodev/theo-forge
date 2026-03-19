package forge

import "github.com/usetheo/theo/forge/validate"

// ValidateBinaryUnit validates a binary resource unit (memory: Ki, Mi, Gi, Ti, Pi, Ei).
func ValidateBinaryUnit(s string) error {
	return validate.BinaryUnit(s)
}

// ValidateDecimalUnit validates a decimal resource unit (CPU: m, k, M, G, T, P, E).
func ValidateDecimalUnit(s string) error {
	return validate.DecimalUnit(s)
}

// ConvertBinaryUnit converts a binary unit string to its numeric value in base units (bytes).
func ConvertBinaryUnit(s string) (float64, error) {
	return validate.ConvertBinaryUnit(s)
}

// ConvertDecimalUnit converts a decimal unit string to its numeric value in base units.
func ConvertDecimalUnit(s string) (float64, error) {
	return validate.ConvertDecimalUnit(s)
}

// ValidateResourceRequirements checks that requests don't exceed limits and values are positive.
func ValidateResourceRequirements(r ResourceRequirements) error {
	return validate.ResourceRequirements(r)
}
