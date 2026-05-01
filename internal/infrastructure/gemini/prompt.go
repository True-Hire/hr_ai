package gemini

import (
	"fmt"
)

func buildPrompt(userInput string, taxonomyContext string) string {
	return fmt.Sprintf(`You are an expert HR assistant. Parse the following candidate profile information and return a strictly valid JSON object.
STRICT TAXONOMY RULES:
1. **MAIN CATEGORY**: Identify the primary professional field (e.g., IT, Construction, Sales, Medicine).
2. **SUB-CATEGORY**: Identify the specific specialization (e.g., Backend, Mobile, UI/UX, Sales, Management).
3. **TECHNOLOGIES**: Extract tools, software, frameworks, and hardware used (e.g., Python, Docker, Figma, Perforator, Stethoscope).
4. **SKILLS**: Extract conceptual methodologies and professional knowledge (e.g., Agile, Team Management, Tile Laying, Surgery).

TAXONOMY CONTEXT (Use IDs if matched, otherwise suggest new names):
%s

The output must match this structure:
{
  "source_lang": "uz/ru/en",
  "profile_score": 0-100,
  "fields": {
    "full_name": {"uz": "...", "ru": "...", "en": "..."},
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "matched_main_category_id": "UUID or empty",
  "matched_sub_category_id": "UUID or empty",
  "new_main_category": "Name if no match",
  "new_sub_category": "Name if no match",
  "matched_technology_ids": ["UUID", "UUID"],
  "matched_skill_ids": ["UUID", "UUID"],
  "new_technologies": ["name1", "name2"],
  "new_skills": ["name1", "name2"],
  "certifications": {
    "uz": ["cert1"], "ru": ["cert1"], "en": ["cert1"]
  },
  "languages": [
    {"name": {"uz": "Ingliz tili", "en": "English"}, "level": "B2"}
  ],
  "experience": [
    {
      "company": "...",
      "position": {"uz": "...", "en": "..."},
      "start_date": "YYYY-MM",
      "end_date": "YYYY-MM or present",
      "description": {"uz": "...", "en": "..."},
      "projects": [
         {"project": "Project Name", "items": {"uz": ["task1"], "en": ["task1"]}}
      ]
    }
  ],
  "education": [
    {
      "institution": "...",
      "degree": {"uz": "...", "en": "..."},
      "field_of_study": {"uz": "...", "en": "..."},
      "start_date": "YYYY",
      "end_date": "YYYY"
    }
  ]
}

If information is missing, use empty strings or arrays.
User Input: %s`, taxonomyContext, userInput)
}

func buildFilePrompt(taxonomyContext string) string {
	return fmt.Sprintf(`You are an expert HR assistant. Analyze the attached file (CV/Resume) and extract the candidate profile information.
STRICT TAXONOMY RULES:
1. **MAIN CATEGORY**: Identify the primary professional field (e.g., IT, Construction, Sales, Medicine).
2. **SUB-CATEGORY**: Identify the specific specialization (e.g., Backend, Mobile, UI/UX, Sales, Management).
3. **TECHNOLOGIES**: Extract tools, software, frameworks, and hardware used (e.g., Python, Docker, Figma, Perforator, Stethoscope).
4. **SKILLS**: Extract conceptual methodologies and professional knowledge (e.g., Agile, Team Management, Tile Laying, Surgery).

TAXONOMY CONTEXT (Use IDs if matched, otherwise suggest new names):
%s

Return a strictly valid JSON object.
The output must match this structure:
{
  "source_lang": "uz/ru/en",
  "profile_score": 0-100,
  "fields": {
    "full_name": {"uz": "...", "ru": "...", "en": "..."},
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "matched_main_category_id": "UUID or empty",
  "matched_sub_category_id": "UUID or empty",
  "new_main_category": "Name if no match",
  "new_sub_category": "Name if no match",
  "matched_technology_ids": ["UUID", "UUID"],
  "matched_skill_ids": ["UUID", "UUID"],
  "new_technologies": ["name1", "name2"],
  "new_skills": ["name1", "name2"],
  "certifications": {
    "uz": ["cert1"], "ru": ["cert1"], "en": ["cert1"]
  },
  "languages": [
    {"name": {"uz": "Ingliz tili", "en": "English"}, "level": "B2"}
  ],
  "experience": [
    {
      "company": "...",
      "position": {"uz": "...", "en": "..."},
      "start_date": "YYYY-MM",
      "end_date": "YYYY-MM or present",
      "description": {"uz": "...", "en": "..."},
      "projects": [
         {"project": "Project Name", "items": {"uz": ["task1"], "en": ["task1"]}}
      ]
    }
  ],
  "education": [
    {
      "institution": "...",
      "degree": {"uz": "...", "en": "..."},
      "field_of_study": {"uz": "...", "en": "..."},
      "start_date": "YYYY",
      "end_date": "YYYY"
    }
  ]
}

If information is missing, use empty strings or arrays. Do not include any explanation outside the JSON.`, taxonomyContext)
}

func buildCompanyParsePrompt(userInput string) string {
	return fmt.Sprintf(`Analyze the following company information and return a strictly valid JSON object.
Structure:
{
  "source_lang": "uz/ru/en",
  "fields": {
    "name": {"uz": "...", "en": "..."},
    "description": {"uz": "...", "en": "..."}
  },
  "employee_count": 0,
  "country": "...",
  "address": "...",
  "phone": "...",
  "telegram": "...",
  "email": "...",
  "web_site": "..."
}
Input: %s`, userInput)
}

func buildCompanyPrompt(input string) string {
	return fmt.Sprintf(`Translate and structure this company name/short info: %s. Return JSON with "fields" map and "source_lang".`, input)
}

func buildVacancyPrompt(input string) string {
	return fmt.Sprintf(`Translate and structure this job title/short info: %s. Return JSON with "fields" map and "source_lang".`, input)
}

func buildVacancyParsePrompt(userInput string) string {
	return fmt.Sprintf(`Extract job vacancy details from the text below and return a strictly valid JSON object.
Structure:
{
  "source_lang": "uz/ru/en",
  "fields": {
    "title": {"uz": "...", "en": "..."},
    "description": {"uz": "...", "en": "..."}
  },
  "salary_min": 0,
  "salary_max": 0,
  "salary_currency": "USD/UZS",
  "experience_min": 0,
  "experience_max": 0,
  "format": "remote/office/hybrid",
  "schedule": "full-time/part-time",
  "skills": {"en": ["skill1", "skill2"]}
}
Input: %s`, userInput)
}

func buildVacancyMergePrompt(existingJSON, additionalInfo string) string {
	return fmt.Sprintf(`Update the existing vacancy JSON with the additional information provided.
Existing JSON: %s
Additional Info: %s`, existingJSON, additionalInfo)
}

func buildVacancyEnhancePrompt(draftJSON string) string {
	return fmt.Sprintf(`Enhance the professional tone and clarity of this vacancy description.
Draft JSON: %s`, draftJSON)
}

func buildTranslateToEnglishPrompt(text string) string {
	return fmt.Sprintf(`Translate the following text to English. Return JSON: {"text": "..."}.
Text: %s`, text)
}

func buildTranslateTextPrompt(text string) string {
	return fmt.Sprintf(`Translate the following text into Uzbek, Russian, and English.
Return JSON:
{
  "source_lang": "...",
  "translations": {
    "uz": "...",
    "ru": "...",
    "en": "..."
  }
}
Text: %s`, text)
}

func buildSalaryEstimationPrompt(profileSummary, country string) string {
	return fmt.Sprintf(`Estimate the salary range for a candidate with the following profile in %s.
Profile Summary: %s
Return JSON: {"salary_min": 0, "salary_max": 0, "currency": "USD"}`, country, profileSummary)
}
