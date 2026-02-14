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

func buildVacancyPrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that translates vacancy/job posting information into 3 languages: Uzbek (uz), Russian (ru), and English (en).

Analyze the following vacancy data and translate the text fields. If the original text is in one language, translate the content to the other two languages.

Detect which language the input is primarily written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Job title/position name
- description: General description of the vacancy
- responsibilities: Job responsibilities and duties
- requirements: Requirements for the candidate
- benefits: What the company offers (benefits, perks)

IMPORTANT RULES:
- Only include fields where you can extract meaningful content
- Do NOT include fields with empty or placeholder content
- Translate accurately and naturally into all 3 languages
- Keep professional tone appropriate for job postings

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "description": {"uz": "...", "ru": "...", "en": "..."},
    "responsibilities": {"uz": "...", "ru": "...", "en": "..."},
    "requirements": {"uz": "...", "ru": "...", "en": "..."},
    "benefits": {"uz": "...", "ru": "...", "en": "..."}
  }
}

Vacancy data:
%s`, userInput)
}

func buildVacancyParsePrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that extracts structured vacancy/job posting information from free-form text provided by an HR manager.

Analyze the following text and extract ALL relevant information. For text fields, provide translations in all 3 languages: Uzbek (uz), Russian (ru), and English (en).

Detect which language the input is primarily written in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Job title/position name
- description: General description of the vacancy
- responsibilities: Job responsibilities and duties
- requirements: Requirements for the candidate (qualifications, experience needed, etc.)
- benefits: What the company offers (salary details mentioned in text, perks, benefits)

NON-TEXT FIELDS (extract as-is, do not translate):
- salary_min: Minimum salary as integer (0 if not mentioned)
- salary_max: Maximum salary as integer (0 if not mentioned)
- salary_currency: Currency code like "USD", "UZS", "RUB" (default "USD" if not mentioned)
- experience_min: Minimum years of experience as integer (0 if not mentioned)
- experience_max: Maximum years of experience as integer (0 if not mentioned)
- format: Work format - one of "remote", "hybrid", "office" (default "office" if not mentioned)
- schedule: Work schedule - one of "full-time", "part-time", "contract", "internship" (default "full-time" if not mentioned)
- phone: Phone number if mentioned (empty string if not)
- telegram: Telegram contact if mentioned (empty string if not)
- email: Email contact if mentioned (empty string if not)
- address: Office address if mentioned (empty string if not)

SKILLS (extract as array of strings):
- skills: Array of key skills, technologies, hashtags, tools mentioned in the text. Extract everything that looks like a skill, technology, framework, tool, or hashtag. Examples: "Go", "Python", "Docker", "PostgreSQL", "REST API", "Git", "Agile", "Leadership", etc.

IMPORTANT RULES:
- Extract as much information as possible from the text
- For text fields, translate accurately and naturally into all 3 languages
- For skills, normalize to standard names (e.g. "#golang" → "Go", "#python" → "Python")
- Keep skills in English where possible (technology names are universal)
- If salary is given as a range like "1000-2000$", extract min and max separately
- Do NOT include fields with empty or placeholder content for text fields

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {"uz": "...", "ru": "...", "en": "..."},
    "description": {"uz": "...", "ru": "...", "en": "..."},
    "responsibilities": {"uz": "...", "ru": "...", "en": "..."},
    "requirements": {"uz": "...", "ru": "...", "en": "..."},
    "benefits": {"uz": "...", "ru": "...", "en": "..."}
  },
  "salary_min": 1000,
  "salary_max": 2000,
  "salary_currency": "USD",
  "experience_min": 2,
  "experience_max": 5,
  "format": "office",
  "schedule": "full-time",
  "phone": "+998901234567",
  "telegram": "@company",
  "email": "hr@company.com",
  "address": "Tashkent, Amir Temur 1",
  "skills": ["Go", "PostgreSQL", "Docker", "REST API", "Git"]
}

HR's job posting text:
%s`, userInput)
}

func buildTranslateToEnglishPrompt(text string) string {
	return fmt.Sprintf(`You are a translator. Translate the following text to English. If the text is already in English, return it as-is.

Return ONLY valid JSON in this exact format:
{"text": "translated english text here"}

Text to translate:
%s`, text)
}

func buildTranslateTextPrompt(text string) string {
	return fmt.Sprintf(`You are a translator. Translate the following text into 3 languages: Uzbek (uz), Russian (ru), and English (en).

Detect which language the input is written in and set source_lang accordingly.

Translate accurately and naturally into all 3 languages. Preserve the original meaning and tone.

Return ONLY valid JSON in this exact format:
{
  "source_lang": "en",
  "translations": {
    "uz": "translated text in Uzbek",
    "ru": "translated text in Russian",
    "en": "translated text in English"
  }
}

Text to translate:
%s`, text)
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
