package domain

// AIParsedVacancy represents the exact JSON structure we expect from Claude AI.
type AIParsedVacancy struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	MatchedMainCatID  string   `json:"matched_main_category_id"`
	MatchedSubCatID   string   `json:"matched_sub_category_id"`
	NewMainCategory   string   `json:"new_main_category"`
	NewSubCategory    string   `json:"new_sub_category"`
	MatchedTechIDs    []string `json:"matched_technology_ids"`
	MatchedSkillIDs   []string `json:"matched_skill_ids"`
	NewTechnologies   []string `json:"new_technologies"`
	NewSkills         []string `json:"new_skills"`
}
