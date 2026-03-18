package scoring

import (
	"math"
	"strings"
)

// ParsedQuery represents a parsed HR search query
type ParsedQuery struct {
	RoleFamily               string   `json:"role_family"`
	PrimaryRole              string   `json:"primary_role"`
	MustDomains              []string `json:"must_domains"`
	PreferredDomains         []string `json:"preferred_domains"`
	Skills                   []string `json:"skills"`
	Seniority                string   `json:"seniority"`
	PreferredEducationFields []string `json:"preferred_education_fields"`
	LocationCity             string   `json:"location_city"`
}

// CalcRoleMatchScore computes role match between query and candidate
func CalcRoleMatchScore(queryRole, queryFamily, candidateRole, candidateFamily string) float64 {
	if queryRole == "" && queryFamily == "" {
		return 0.5
	}
	if candidateRole == queryRole && queryRole != "" {
		return 1.0
	}
	if candidateFamily == queryFamily && queryFamily != "" {
		return 0.75
	}
	// Adjacent families
	adjacent := map[string][]string{
		"backend":  {"fullstack", "devops"},
		"frontend": {"fullstack", "design"},
		"mobile":   {"frontend", "fullstack"},
		"data":     {"backend"},
		"devops":   {"backend"},
	}
	if adj, ok := adjacent[queryFamily]; ok {
		for _, a := range adj {
			if candidateFamily == a {
				return 0.45
			}
		}
	}
	return 0.0
}

// CalcDomainMatchScore computes domain overlap
func CalcDomainMatchScore(mustDomains, preferredDomains, candidateDomains []string) float64 {
	if len(mustDomains) == 0 && len(preferredDomains) == 0 {
		return 0.5
	}
	candidateSet := make(map[string]bool)
	for _, d := range candidateDomains {
		candidateSet[d] = true
	}

	mustCoverage := 0.0
	if len(mustDomains) > 0 {
		matched := 0
		for _, d := range mustDomains {
			if candidateSet[d] {
				matched++
			}
		}
		mustCoverage = float64(matched) / float64(len(mustDomains))
	}

	prefCoverage := 0.0
	if len(preferredDomains) > 0 {
		matched := 0
		for _, d := range preferredDomains {
			if candidateSet[d] {
				matched++
			}
		}
		prefCoverage = float64(matched) / float64(len(preferredDomains))
	}

	if len(mustDomains) > 0 {
		return clamp01(0.7*mustCoverage + 0.3*prefCoverage)
	}
	return clamp01(prefCoverage)
}

// CalcSkillMatchScore computes skill overlap between query and candidate
func CalcSkillMatchScore(querySkills, candidateSkills []string) float64 {
	if len(querySkills) == 0 {
		return 0.5
	}
	candidateSet := make(map[string]bool)
	for _, s := range candidateSkills {
		candidateSet[s] = true
	}
	matched := 0
	for _, s := range querySkills {
		normalized := NormalizeSkill(s)
		if candidateSet[normalized] || candidateSet[s] || candidateSet[strings.ToLower(s)] {
			matched++
		}
	}
	return float64(matched) / float64(len(querySkills))
}

// CalcEducationMatchScore computes education field match
func CalcEducationMatchScore(preferredFields, candidateFields []string) float64 {
	if len(preferredFields) == 0 {
		return 0.5
	}
	for _, pf := range preferredFields {
		pfLower := strings.ToLower(pf)
		for _, cf := range candidateFields {
			cfLower := strings.ToLower(cf)
			if strings.Contains(cfLower, pfLower) || strings.Contains(pfLower, cfLower) {
				return 1.0
			}
		}
	}
	// Check adjacent
	adjacentFields := map[string][]string{
		"economics":        {"finance", "business", "accounting"},
		"computer science": {"software", "engineering", "информатик", "программ"},
		"mathematics":      {"statistics", "математик"},
	}
	for _, pf := range preferredFields {
		pfLower := strings.ToLower(pf)
		if adjacents, ok := adjacentFields[pfLower]; ok {
			for _, adj := range adjacents {
				for _, cf := range candidateFields {
					if strings.Contains(strings.ToLower(cf), adj) {
						return 0.6
					}
				}
			}
		}
	}
	return 0.0
}

// CalcSeniorityMatchScore computes seniority fit
func CalcSeniorityMatchScore(querySeniority, candidateSeniority string) float64 {
	if querySeniority == "" {
		return 0.5
	}
	if candidateSeniority == querySeniority {
		return 1.0
	}
	levels := map[string]int{
		"intern": 0, "junior": 1, "middle": 2, "senior": 3, "lead": 4,
	}
	ql, qOk := levels[querySeniority]
	cl, cOk := levels[candidateSeniority]
	if !qOk || !cOk {
		return 0.3
	}
	diff := math.Abs(float64(ql - cl))
	if diff <= 1 {
		return 0.6
	}
	return 0.2
}

// CalcQueryRelevanceScore combines all relevance components
func CalcQueryRelevanceScore(roleMatch, domainMatch, skillMatch, educationMatch, seniorityMatch, textRank float64) float64 {
	return clamp01(
		0.35*roleMatch +
			0.20*domainMatch +
			0.18*skillMatch +
			0.10*educationMatch +
			0.07*seniorityMatch +
			0.10*textRank,
	)
}

// RoleBonusWeights defines role-family-specific bonus weights
type RoleBonusWeights struct {
	DevOpsSupport       float64
	Ownership           float64
	EngineeringEnv      float64
	ProjectManagement   float64
	ClientCommunication float64
	Leadership          float64
}

// GetRoleBonusWeights returns weights for a given role family
func GetRoleBonusWeights(roleFamily string) RoleBonusWeights {
	switch roleFamily {
	case "backend":
		return RoleBonusWeights{0.28, 0.24, 0.16, 0.12, 0.10, 0.10}
	case "frontend":
		return RoleBonusWeights{0.10, 0.20, 0.10, 0.15, 0.25, 0.20}
	case "mobile":
		return RoleBonusWeights{0.15, 0.25, 0.10, 0.15, 0.20, 0.15}
	case "data":
		return RoleBonusWeights{0.10, 0.20, 0.25, 0.15, 0.15, 0.15}
	case "devops":
		return RoleBonusWeights{0.35, 0.20, 0.20, 0.10, 0.05, 0.10}
	case "qa":
		return RoleBonusWeights{0.15, 0.20, 0.15, 0.20, 0.15, 0.15}
	default:
		return RoleBonusWeights{0.20, 0.20, 0.15, 0.15, 0.15, 0.15}
	}
}

// CalcRoleSpecificBonusScore computes role-specific bonus
func CalcRoleSpecificBonusScore(roleFamily string, devopsSupport, ownership, engEnv, projectMgmt, clientComm, leadership float64) float64 {
	w := GetRoleBonusWeights(roleFamily)
	return clamp01(
		w.DevOpsSupport*devopsSupport +
			w.Ownership*ownership +
			w.EngineeringEnv*engEnv +
			w.ProjectManagement*projectMgmt +
			w.ClientCommunication*clientComm +
			w.Leadership*leadership,
	)
}

// CalcMarketStrengthScore computes market strength
func CalcMarketStrengthScore(companyPrestige, projectComplexity, internshipQuality, educationQuality, growthTrajectory, competition float64) float64 {
	return clamp01(
		0.35*companyPrestige +
			0.20*projectComplexity +
			0.15*internshipQuality +
			0.10*educationQuality +
			0.10*growthTrajectory +
			0.10*competition,
	)
}

// CalcFinalScore computes the final ranking score
func CalcFinalScore(queryRelevance, roleSpecificBonus, marketStrength float64) float64 {
	return 0.55*queryRelevance + 0.25*roleSpecificBonus + 0.20*marketStrength
}
