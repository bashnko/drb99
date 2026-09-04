package service

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate      *validator.Validate
	spdxPattern   = regexp.MustCompile(`^[A-Za-z0-9\.\-\+]+$`)
	semverPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	platformSet   = map[string]struct{}{
		"linux-amd64":   {},
		"linux-arm64":   {},
		"darwin-amd64":  {},
		"darwin-arm64":  {},
		"windows-amd64": {},
	}
)

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("spdx", validateSPDX)
	_ = validate.RegisterValidation("semver", validateSemver)
	_ = validate.RegisterValidation("platform", validatePlatform)
}

func validateSPDX(fl validator.FieldLevel) bool {
	return spdxPattern.MatchString(fl.Field().String())
}

func validateSemver(fl validator.FieldLevel) bool {
	return semverPattern.MatchString(fl.Field().String())
}

func validatePlatform(fl validator.FieldLevel) bool {
	_, ok := platformSet[fl.Field().String()]
	return ok
}

func ValidateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		var errs validator.ValidationErrors
		if !errors.As(err, &errs) {
			return err
		}
		var messages []string
		for _, e := range errs {
			messages = append(messages, formatValidationError(e))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}
	return nil
}

func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "spdx":
		return fmt.Sprintf("%s must be a valid SPDX license identifier", field)
	case "semver":
		return fmt.Sprintf("%s must be a valid semantic version (e.g., v1.2.3)", field)
	case "platform":
		return fmt.Sprintf("%s must be a valid platform (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)", field)
	case "dive":
		return fmt.Sprintf("%s contains invalid entries", field)
	case "keys":
		return fmt.Sprintf("%s keys must be valid URLs", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

func (r *GenerateRequest) Validate() error {
	return ValidateStruct(r)
}

func (r *PrefillRequest) Validate() error {
	return ValidateStruct(r)
}

func (f *Features) IsEmpty() bool {
	if f == nil {
		return true
	}
	v := reflect.ValueOf(f).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Bool() {
			return false
		}
	}
	return true
}
