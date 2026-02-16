package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL      = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
	embeddingURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"
	modelVersion = "gemini-2.5-flash"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) ModelVersion() string {
	return modelVersion
}

// LangStringSlice can unmarshal from both a JSON object (map) and an empty array.
// Gemini sometimes returns [] instead of {} when there are no entries.
type LangStringSlice map[string][]string

func (l *LangStringSlice) UnmarshalJSON(data []byte) error {
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err == nil {
		*l = m
		return nil
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		*l = make(map[string][]string)
		return nil
	}
	return fmt.Errorf("expected object or array, got %s", string(data))
}

// FlexibleFields handles Gemini returning field values as either strings or arrays.
// e.g. "title": {"en": "Dev"} vs "achievements": {"en": ["Award 1", "Award 2"]}
// Array values are joined with "\n" so downstream code always gets map[string]string.
type FlexibleFields map[string]map[string]string

func (f *FlexibleFields) UnmarshalJSON(data []byte) error {
	var raw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	result := make(map[string]map[string]string, len(raw))
	for fieldName, translations := range raw {
		result[fieldName] = make(map[string]string, len(translations))
		for lang, val := range translations {
			// Try string first
			var s string
			if err := json.Unmarshal(val, &s); err == nil {
				result[fieldName][lang] = s
				continue
			}
			// Try array of strings, join with newline
			var arr []string
			if err := json.Unmarshal(val, &arr); err == nil {
				result[fieldName][lang] = strings.Join(arr, "\n")
				continue
			}
			result[fieldName][lang] = string(val)
		}
	}
	*f = result
	return nil
}

// ParsedProfile is the structured result from Gemini profile parsing.
type ParsedProfile struct {
	SourceLang     string               `json:"source_lang"`
	ProfileScore   int                  `json:"profile_score"`
	Fields         FlexibleFields       `json:"fields"`
	Skills         LangStringSlice      `json:"skills"`
	Certifications LangStringSlice      `json:"certifications"`
	Languages      []ParsedLanguageItem `json:"languages"`
	Experience     []ParsedExperienceItem `json:"experience"`
	Education      []ParsedEducationItem  `json:"education"`
}

type ParsedLanguageItem struct {
	Name  map[string]string `json:"name"`
	Level string            `json:"level"`
}

type ParsedProjectItem struct {
	Project string              `json:"project"`
	Items   map[string][]string `json:"items"`
}

type ParsedExperienceItem struct {
	Company     string              `json:"company"`
	Position    map[string]string   `json:"position"`
	StartDate   string              `json:"start_date"`
	EndDate     string              `json:"end_date"`
	Projects    []ParsedProjectItem `json:"projects"`
	WebSite     string              `json:"web_site"`
	Description map[string]string   `json:"description"`
}

type ParsedEducationItem struct {
	Institution  string            `json:"institution"`
	Degree       map[string]string `json:"degree"`
	FieldOfStudy map[string]string `json:"field_of_study"`
	StartDate    string            `json:"start_date"`
	EndDate      string            `json:"end_date"`
	Location     string            `json:"location"`
	Description  map[string]string `json:"description"`
}

// ParsedCompany is the structured result from Gemini company translation.
type ParsedCompany struct {
	SourceLang string                       `json:"source_lang"`
	Fields     map[string]map[string]string `json:"fields"`
}

func (c *Client) TranslateCompany(ctx context.Context, input string) (*ParsedCompany, error) {
	parts := []part{
		{Text: buildCompanyPrompt(input)},
	}

	reqBody := generateRequest{
		Contents:         []content{{Parts: parts}},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text

	var parsed ParsedCompany
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini company JSON: %w (raw: %s)", err, text)
	}

	return &parsed, nil
}

// ParsedVacancy is the structured result from Gemini vacancy text translation.
type ParsedVacancy struct {
	SourceLang string                       `json:"source_lang"`
	Fields     map[string]map[string]string `json:"fields"`
}

// ParsedVacancyFull is the result from Gemini when parsing a full job posting text.
type ParsedVacancyFull struct {
	SourceLang    string                       `json:"source_lang"`
	Fields        map[string]map[string]string `json:"fields"`
	SalaryMin     int32                        `json:"salary_min"`
	SalaryMax     int32                        `json:"salary_max"`
	SalaryCurrency string                      `json:"salary_currency"`
	ExperienceMin int32                        `json:"experience_min"`
	ExperienceMax int32                        `json:"experience_max"`
	Format        string                       `json:"format"`
	Schedule      string                       `json:"schedule"`
	Phone         string                       `json:"phone"`
	Telegram      string                       `json:"telegram"`
	Email         string                       `json:"email"`
	Address       string                       `json:"address"`
	Skills        []string                     `json:"skills"`
}

func (c *Client) TranslateVacancy(ctx context.Context, input string) (*ParsedVacancy, error) {
	parts := []part{
		{Text: buildVacancyPrompt(input)},
	}

	reqBody := generateRequest{
		Contents:         []content{{Parts: parts}},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text

	var parsed ParsedVacancy
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy JSON: %w (raw: %s)", err, text)
	}

	return &parsed, nil
}

func (c *Client) ParseVacancyFromText(ctx context.Context, userInput string) (*ParsedVacancyFull, error) {
	parts := []part{
		{Text: buildVacancyParsePrompt(userInput)},
	}

	reqBody := generateRequest{
		Contents:         []content{{Parts: parts}},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text

	var parsed ParsedVacancyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy full JSON: %w (raw: %s)", err, text)
	}

	return &parsed, nil
}

func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]any{
		"model": "models/gemini-embedding-001",
		"content": map[string]any{
			"parts": []map[string]string{
				{"text": text},
			},
		},
		"outputDimensionality": 768,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	url := embeddingURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal embedding response: %w", err)
	}

	return result.Embedding.Values, nil
}

func (c *Client) TranslateToEnglish(ctx context.Context, text string) (string, error) {
	parts := []part{
		{Text: buildTranslateToEnglishPrompt(text)},
	}

	reqBody := generateRequest{
		Contents:         []content{{Parts: parts}},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal translate request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create translate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("translate API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read translate response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("translate API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", fmt.Errorf("unmarshal translate response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("translate returned no content")
	}

	resultText := genResp.Candidates[0].Content.Parts[0].Text

	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return "", fmt.Errorf("parse translate JSON: %w (raw: %s)", err, resultText)
	}

	return parsed.Text, nil
}

// TranslatedText is the result from Gemini when translating a single text into 3 languages.
type TranslatedText struct {
	SourceLang   string            `json:"source_lang"`
	Translations map[string]string `json:"translations"`
}

func (c *Client) TranslateText(ctx context.Context, text string) (*TranslatedText, error) {
	parts := []part{
		{Text: buildTranslateTextPrompt(text)},
	}

	reqBody := generateRequest{
		Contents:         []content{{Parts: parts}},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	resultText := genResp.Candidates[0].Content.Parts[0].Text

	var parsed TranslatedText
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini translate JSON: %w (raw: %s)", err, resultText)
	}

	return &parsed, nil
}

func (c *Client) ParseProfileFromText(ctx context.Context, userInput string) (*ParsedProfile, error) {
	parts := []part{
		{Text: buildPrompt(userInput)},
	}
	return c.callGemini(ctx, parts)
}

func (c *Client) ParseProfileFromFile(ctx context.Context, fileData []byte, mimeType string) (*ParsedProfile, error) {
	encoded := base64.StdEncoding.EncodeToString(fileData)
	parts := []part{
		{Text: buildFilePrompt()},
		{InlineData: &inlineData{MimeType: mimeType, Data: encoded}},
	}
	return c.callGemini(ctx, parts)
}

func (c *Client) callGemini(ctx context.Context, parts []part) (*ParsedProfile, error) {
	reqBody := generateRequest{
		Contents: []content{{Parts: parts}},
		GenerationConfig: generationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := baseURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	text := genResp.Candidates[0].Content.Parts[0].Text

	var profile ParsedProfile
	if err := json.Unmarshal([]byte(text), &profile); err != nil {
		return nil, fmt.Errorf("parse gemini profile JSON: %w (raw: %s)", err, text)
	}

	return &profile, nil
}

// Gemini API request types

type generateRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generation_config"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type generationConfig struct {
	ResponseMimeType string `json:"response_mime_type"`
}

// Gemini API response types

type generateResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content contentResponse `json:"content"`
}

type contentResponse struct {
	Parts []partResponse `json:"parts"`
}

type partResponse struct {
	Text string `json:"text"`
}
