package gemini

import "fmt"

func buildPrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that extracts structured profile information from user-provided text for a job platform.

Analyze the following user input and extract relevant profile fields. For each text field, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the input is written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Professional title or headline
- about: A summary paragraph about the person
- skills: Technical and soft skills
- languages: Languages the person speaks
- certifications: Certifications, licenses, or courses
- achievements: Notable achievements, awards, or accomplishments

STRUCTURED ARRAYS:
- experience: Array of work experiences. Each item has:
  - company (string, not translated)
  - position (translated: uz/ru/en)
  - start_date (string, e.g. "2020")
  - end_date (string, e.g. "2023" or "present")
  - projects (string, not translated)
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
- experience and education MUST be arrays (even if there's only one item)

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "experience": [
    {
      "company": "Company Name",
      "position": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2020",
      "end_date": "2023",
      "projects": "Project details",
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

func buildFilePrompt() string {
	return `You are an AI that extracts structured profile information from uploaded documents (resumes, CVs, profiles) for a job platform.

Analyze the uploaded document and extract relevant profile fields. For each text field, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the document is primarily written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Professional title or headline
- about: A summary paragraph about the person
- skills: Technical and soft skills
- languages: Languages the person speaks
- certifications: Certifications, licenses, or courses
- achievements: Notable achievements, awards, or accomplishments

STRUCTURED ARRAYS:
- experience: Array of work experiences. Each item has:
  - company (string, not translated)
  - position (translated: uz/ru/en)
  - start_date (string, e.g. "2020")
  - end_date (string, e.g. "2023" or "present")
  - projects (string, not translated)
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
- experience and education MUST be arrays (even if there's only one item)

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "about": {"uz": "...", "ru": "...", "en": "..."}
  },
  "experience": [
    {
      "company": "Company Name",
      "position": {"uz": "...", "ru": "...", "en": "..."},
      "start_date": "2020",
      "end_date": "2023",
      "projects": "Project details",
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
