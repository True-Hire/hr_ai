package gemini

import "fmt"

const profileParseInstructions = `You are an AI that extracts structured profile information for a job platform. Your goal is to build the MOST COMPLETE resume possible from the provided data.

Analyze ALL provided content and extract every piece of relevant information. For each text field, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original content is in one language, translate it to the other two.

Detect which language the content is primarily in and set source_lang to one of: "uz", "ru", "en".

TEXT FIELDS (translated into 3 languages):
- title: Professional title or headline
- about: A comprehensive summary paragraph about the person
- achievements: Notable achievements, awards, or accomplishments

SKILLS — CRITICAL INSTRUCTIONS:
- Extract skills from EVERY section: title, about, experience, education, projects, certifications, achievements
- If the title says "UI/UX and Graphic Designer" → extract "UI/UX Design", "Graphic Design" as skills
- If the about text mentions tools like "WordPress, Tilda, Opencart" → extract EACH as a separate skill
- If someone says "project manager" or "руководитель проектов" → extract "Project Management"
- If someone mentions "writing technical specifications" → extract "Technical Writing", "Requirements Analysis"
- If someone mentions "conducting negotiations" → extract "Negotiation", "Communication"
- If someone mentions "mobile applications" → extract "Mobile App Design" or "Mobile Development"
- Extract ALL tools, platforms, technologies, frameworks, methodologies, and soft skills
- Be EXHAUSTIVE — extract every possible skill that is mentioned, implied, or can be inferred
- Each skill is a short tag/label (e.g. "Go", "Docker", "Web Design", "Project Management")
- Provide the full array translated for each language
- skills format: {"uz": ["...", "..."], "ru": ["...", "..."], "en": ["...", "..."]}

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

PROFILE SCORE — CRITICAL:
- profile_score: Integer 0-100. Evaluate this resume BRUTALLY and HONESTLY.
- How likely is this person to find a job with this resume?
- Consider: completeness of information, specificity of skills, quality of experience descriptions, presence of measurable achievements, education relevance, language proficiency.
- Do NOT be generous or user-friendly. Rate REALISTICALLY:
  - 0-20: Almost no useful information, cannot be used as a resume
  - 20-35: Very basic info, missing critical sections (no experience details, vague skills)
  - 35-50: Has some content but major gaps (e.g. experience mentioned in about but no structured entries)
  - 50-65: Decent resume with most sections filled but lacking specifics or measurable results
  - 65-80: Good resume with detailed experience, specific skills, and some achievements
  - 80-90: Excellent resume with quantified achievements, comprehensive skills, strong experience
  - 90-100: Almost impossible to achieve — reserved for perfectly structured resumes with extraordinary detail

IMPORTANT RULES:
- Build the MOST COMPLETE resume possible from whatever data is available
- If about text mentions working on specific types of projects but no experience entries exist, use that information to enrich the about section and extract skills
- Fill in gaps using context clues — infer everything you can
- Do NOT include fields with empty or placeholder content
- Translate accurately and naturally into all 3 languages
- skills, certifications, languages, experience and education MUST be arrays (even if there's only one item)

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "profile_score": 42,
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

func buildPrompt(userInput string) string {
	return fmt.Sprintf(`%s

User input:
%s`, profileParseInstructions, userInput)
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

func buildVacancyMergePrompt(existingJSON, additionalInfo string) string {
	return fmt.Sprintf(`You are an AI that merges additional information into an existing vacancy/job posting.

Below is the EXISTING vacancy data in JSON format, followed by NEW additional information from the HR manager.

Your task:
1. Take ALL existing data and keep it
2. Merge the new information — update fields that the new info clarifies, ADD new details to descriptions/requirements/responsibilities
3. Do NOT remove or lose any existing information
4. If the new info contradicts existing data, prefer the new info
5. For text fields (title, description, responsibilities, requirements, benefits), APPEND new details to existing content, translate into all 3 languages
6. For skills, merge both lists (no duplicates)

Return the MERGED result in the same JSON format.

EXISTING VACANCY DATA:
%s

NEW ADDITIONAL INFORMATION:
%s

Return ONLY valid JSON in the same format as the existing data.`, existingJSON, additionalInfo)
}

func buildVacancyEnhancePrompt(draftJSON string) string {
	return fmt.Sprintf(`You are a professional HR copywriter. You are given a raw vacancy draft (parsed from an HR manager's informal input). Your task is to rewrite it into a polished, professional job posting.

IMPORTANT: You must keep ALL the existing structured data (salary, experience, format, schedule, phone, telegram, email, address, skills) EXACTLY as-is. Only rewrite the TEXT FIELDS to be professional and compelling.

TEXT FIELDS to rewrite (in all 3 languages: en, ru, uz):
- title: Clean, professional job title (e.g. "Senior Go Backend Developer" not "we need a go dev")
- description: A compelling 2-3 sentence overview of the role and why someone should apply
- responsibilities: Clear, bullet-point-style list of duties (separated by "; ")
- requirements: Clear list of must-have qualifications (separated by "; ")
- benefits: Attractive list of what the company offers (separated by "; ")

WRITING GUIDELINES:
- Professional but engaging tone
- Be specific and detailed — expand on vague points
- For responsibilities: start each item with an action verb
- For requirements: be clear about what's mandatory vs nice-to-have
- For benefits: highlight both tangible (salary, insurance) and intangible (growth, culture) perks
- Translate naturally — don't do word-for-word translation. Each language should read as if written by a native speaker.
- Russian job postings often use "Мы ищем..." or "Обязанности:" style
- Uzbek job postings should use professional Uzbek, not transliterated Russian

Return ONLY valid JSON in the EXACT SAME format as the input (same structure, same non-text field values, rewritten text fields).

VACANCY DRAFT:
%s`, draftJSON)
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

func buildSalaryEstimationPrompt(profileSummary, country string) string {
	return fmt.Sprintf(`You are an expert salary analyst. Based on the professional profile below and the person's country of residence, estimate the average monthly salary range this person could realistically earn.

PROFILE:
%s

COUNTRY: %s

IMPORTANT RULES:
- Estimate the salary for the person's COUNTRY, not US/EU rates (unless they live there)
- Be REALISTIC — use actual market data for that country
- Consider: experience level, skills, specialization, industry
- Return salary in the most common currency for that country (e.g. UZS for Uzbekistan, RUB for Russia, USD for USA, KZT for Kazakhstan, etc.)
- If the country is empty or unknown, default to USD and international remote rates
- salary_min is the lower end of what this person could earn monthly
- salary_max is the upper end of what this person could earn monthly
- Be honest and realistic, not optimistic

Return ONLY valid JSON:
{
  "salary_min": 5000000,
  "salary_max": 8000000,
  "currency": "UZS"
}`, profileSummary, country)
}

func buildFilePrompt() string {
	return `The uploaded file is a resume, CV, profile document, or voice/audio recording where a person describes their experience, skills, and background. Extract all information from it.

` + profileParseInstructions
}
