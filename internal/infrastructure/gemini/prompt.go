package gemini

import "fmt"

func buildPrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that extracts structured profile information from user-provided text for a job platform.

Analyze the following user input and extract relevant profile fields. For each text field, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the input is written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Professional title or headline
- about: A summary paragraph about the person
- achievements: Notable achievements, awards, or accomplishments

SKILLS (array of skill tag strings, translated per language):
- skills: Each skill is a short tag/label (e.g. "Go", "Docker", "Web Design"). Provide the full array translated for each language.

CERTIFICATIONS (array of certification strings, translated per language):
- certifications: Each certification is a short label. Provide the full array translated for each language.

LANGUAGES (array of objects):
- languages: Array of languages the person speaks. Each item has:
  - name (translated: uz/ru/en)
  - level (string, e.g. "B2", "C1", "Native", not translated)

STRUCTURED ARRAYS:
- experience: Array of work experiences. Each item has:
  - company (string, not translated)
  - position (translated: uz/ru/en)
  - start_date (string, e.g. "2020")
  - end_date (string, e.g. "2023" or "present")
  - projects: Array of project objects. Each project has:
    - project (string, not translated - project name)
    - items (translated: uz/ru/en - array of strings describing what was done)
  - web_site (string URL, not translated)
  - description (translated: uz/ru/en)

- education: Array of education entries. Each item has:
  - institution (string, not translated)
  - degree (translated: uz/ru/en)
  - field_of_study (translated: uz/ru/en)
  - start_date (string)
  - end_date (string)
  - location (string, not translated)
  - description (translated: uz/ru/en)

IMPORTANT RULES:
- Only include fields/items where you can extract meaningful content
- Do NOT include fields with empty or placeholder content
- Translate accurately and naturally into all 3 languages
- skills, certifications, languages, experience and education MUST be arrays (even if there's only one item)

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "skills": {
    "uz": ["Навык 1", "Навык 2"],
    "ru": ["Навык 1", "Навык 2"],
    "en": ["Skill 1", "Skill 2"]
  },
  "certifications": {
    "uz": ["Sertifikat 1"],
    "ru": ["Сертификат 1"],
    "en": ["Certificate 1"]
  },
  "languages": [
    {
      "name": {"uz": "Ingliz tili", "ru": "Английский", "en": "English"},
      "level": "B2"
    }
  ],
  "experience": [
    {
      "company": "Company Name",
      "position": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2020",
      "end_date": "2023",
      "projects": [
        {
          "project": "Project Name",
          "items": {"uz": ["...", "..."], "ru": ["...", "..."], "en": ["...", "..."]}
        }
      ],
      "web_site": "https://example.com",
      "description": {"uz": "...", "ru": "...", "en": "..."}
    }
  ],
  "education": [
    {
      "institution": "University Name",
      "degree": {"uz": "...", "ru": "...", "en": "..."},
      "field_of_study": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2014",
      "end_date": "2018",
      "location": "City",
      "description": {"uz": "...", "ru": "...", "en": "..."}
    }
  ]
}

User input:
%s`, userInput)
}

func buildCompanyPrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that translates company information into 3 languages: Uzbek (uz), Russian (ru), and English (en).

Analyze the following company data and translate the text fields. If the original text is in one language, translate the content to the other two languages.

Detect which language the input is primarily written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- name: Company name
- activity_type: Type of business activity
- company_type: Type of company (e.g. LLC, Inc, etc.)
- about: Description of the company
- market: Market or industry the company operates in

IMPORTANT RULES:
- Only include fields where you can extract meaningful content
- Do NOT include fields with empty or placeholder content
- Translate accurately and naturally into all 3 languages
- If a field is a proper noun (like a company name), transliterate rather than translate

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "name": {"uz": "...", "ru": "...", "en": "..."},
    "activity_type": {"uz": "...", "ru": "...", "en": "..."},
    "company_type": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."},
    "market": {"uz": "...", "ru": "...", "en": "..."}
  }
}

Company data:
%s`, userInput)
}

func buildFilePrompt() string {
	return `You are an AI that extracts structured profile information from uploaded documents (resumes, CVs, profiles) for a job platform.

Analyze the uploaded document and extract relevant profile fields. For each text field, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the document is primarily written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Professional title or headline
- about: A summary paragraph about the person
- achievements: Notable achievements, awards, or accomplishments

SKILLS (array of skill tag strings, translated per language):
- skills: Each skill is a short tag/label (e.g. "Go", "Docker", "Web Design"). Provide the full array translated for each language.

CERTIFICATIONS (array of certification strings, translated per language):
- certifications: Each certification is a short label. Provide the full array translated for each language.

LANGUAGES (array of objects):
- languages: Array of languages the person speaks. Each item has:
  - name (translated: uz/ru/en)
  - level (string, e.g. "B2", "C1", "Native", not translated)

STRUCTURED ARRAYS:
- experience: Array of work experiences. Each item has:
  - company (string, not translated)
  - position (translated: uz/ru/en)
  - start_date (string, e.g. "2020")
  - end_date (string, e.g. "2023" or "present")
  - projects: Array of project objects. Each project has:
    - project (string, not translated - project name)
    - items (translated: uz/ru/en - array of strings describing what was done)
  - web_site (string URL, not translated)
  - description (translated: uz/ru/en)

- education: Array of education entries. Each item has:
  - institution (string, not translated)
  - degree (translated: uz/ru/en)
  - field_of_study (translated: uz/ru/en)
  - start_date (string)
  - end_date (string)
  - location (string, not translated)
  - description (translated: uz/ru/en)

IMPORTANT RULES:
- Only include fields/items where you can extract meaningful content
- Do NOT include fields with empty or placeholder content
- Translate accurately and naturally into all 3 languages
- skills, certifications, languages, experience and education MUST be arrays (even if there's only one item)

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "skills": {
    "uz": ["Навык 1", "Навык 2"],
    "ru": ["Навык 1", "Навык 2"],
    "en": ["Skill 1", "Skill 2"]
  },
  "certifications": {
    "uz": ["Sertifikat 1"],
    "ru": ["Сертификат 1"],
    "en": ["Certificate 1"]
  },
  "languages": [
    {
      "name": {"uz": "Ingliz tili", "ru": "Английский", "en": "English"},
      "level": "B2"
    }
  ],
  "experience": [
    {
      "company": "Company Name",
      "position": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2020",
      "end_date": "2023",
      "projects": [
        {
          "project": "Project Name",
          "items": {"uz": ["...", "..."], "ru": ["...", "..."], "en": ["...", "..."]}
        }
      ],
      "web_site": "https://example.com",
      "description": {"uz": "...", "ru": "...", "en": "..."}
    }
  ],
  "education": [
    {
      "institution": "University Name",
      "degree": {"uz": "...", "ru": "...", "en": "..."},
      "field_of_study": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2014",
      "end_date": "2018",
      "location": "City",
      "description": {"uz": "...", "ru": "...", "en": "..."}
    }
  ]
}`
}
