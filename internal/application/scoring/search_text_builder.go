package scoring

import (
	"fmt"
	"strings"
)

// BuildSearchText assembles clean searchable text from normalized profile data
func BuildSearchText(
	role string,
	roleFamily string,
	seniority string,
	skills []string,
	domains []string,
	companies []string,
	universities []string,
	educationFields []string,
	locationCity string,
	experienceMonths int,
) string {
	var parts []string

	if role != "" {
		parts = append(parts, role)
	}
	if seniority != "" {
		parts = append(parts, seniority)
	}
	if experienceMonths > 0 {
		years := experienceMonths / 12
		parts = append(parts, fmt.Sprintf("%d years experience", years))
	}
	if len(skills) > 0 {
		parts = append(parts, "Skills: "+strings.Join(skills, " "))
	}
	if len(domains) > 0 {
		parts = append(parts, "Domains: "+strings.Join(domains, " "))
	}
	if len(companies) > 0 {
		parts = append(parts, "Companies: "+strings.Join(companies, " "))
	}
	if len(universities) > 0 {
		parts = append(parts, "Education: "+strings.Join(universities, " "))
	}
	if len(educationFields) > 0 {
		parts = append(parts, "Fields: "+strings.Join(educationFields, " "))
	}
	if locationCity != "" {
		parts = append(parts, "Location: "+locationCity)
	}
	if roleFamily != "" && roleFamily != role {
		parts = append(parts, roleFamily)
	}

	return strings.Join(parts, ". ")
}
