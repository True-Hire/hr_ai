package gemini

import "fmt"

func buildPrompt(userInput string) string {
	return fmt.Sprintf(`You are an AI that extracts structured profile information from user-provided text for a job platform.

Analyze the following user input and extract relevant profile fields. For each field you can extract, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the input is written in and set source_lang to one of: "uz", "ru", "en".

The fields to extract are:
- title: Professional title or headline (e.g. "Senior Go Developer")
- about: A summary paragraph about the person
- experience: Work experience details
- education: Education history
- skills: Technical and soft skills
- languages: Languages the person speaks
- certifications: Certifications, licenses, or courses
- achievements: Notable achievements, awards, or accomplishments

IMPORTANT RULES:
- Only include fields where you can extract meaningful content from the input
- Do NOT include fields with empty or placeholder content
- Each field's content should be a single string (use newlines within the string for multiple items)
- Translate accurately and naturally into all 3 languages

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {
      "uz": "...",
      "ru": "...",
      "en": "..."
    },
    "about": {
      "uz": "...",
      "ru": "...",
      "en": "..."
    }
  }
}

User input:
%s`, userInput)
}

func buildFilePrompt() string {
	return `You are an AI that extracts structured profile information from uploaded documents (resumes, CVs, profiles) for a job platform.

Analyze the uploaded document and extract relevant profile fields. For each field you can extract, provide the content in all 3 languages: Uzbek (uz), Russian (ru), and English (en). If the original text is in one language, translate the content to the other two languages.

Detect which language the document is primarily written in and set source_lang to one of: "uz", "ru", "en".

The fields to extract are:
- title: Professional title or headline (e.g. "Senior Go Developer")
- about: A summary paragraph about the person
- experience: Work experience details
- education: Education history
- skills: Technical and soft skills
- languages: Languages the person speaks
- certifications: Certifications, licenses, or courses
- achievements: Notable achievements, awards, or accomplishments

IMPORTANT RULES:
- Only include fields where you can extract meaningful content from the document
- Do NOT include fields with empty or placeholder content
- Each field's content should be a single string (use newlines within the string for multiple items)
- Translate accurately and naturally into all 3 languages

Return ONLY valid JSON in this exact format:
{
  "source_lang": "ru",
  "fields": {
    "title": {
      "uz": "...",
      "ru": "...",
      "en": "..."
    },
    "about": {
      "uz": "...",
      "ru": "...",
      "en": "..."
    }
  }
}`
}
