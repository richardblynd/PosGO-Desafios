package services

import (
	"regexp"
)

type ValidationService struct{}

func NewValidationService() *ValidationService {
	return &ValidationService{}
}

func (vs *ValidationService) ValidateZipcode(zipcode string) bool {
	zipCodeRegex := regexp.MustCompile(`^[0-9]{8}$`)
	return zipCodeRegex.MatchString(zipcode)
}
