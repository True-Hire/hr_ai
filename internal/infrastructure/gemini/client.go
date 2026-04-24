package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL      = "https://api.anthropic.com/v1/messages"
	embeddingURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"
	modelVersion = "claude-sonnet-4-6"
)

type Client struct {
	googleKey    string
	anthropicKey string
	httpClient   *http.Client
}

func NewClient(googleKey, anthropicKey string) *Client {
	return &Client{
		googleKey:    googleKey,
		anthropicKey: anthropicKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) ModelVersion() string {
	return modelVersion
}

const maxRetries = 5

var retryDelayRegex = regexp.MustCompile(`"retryDelay":\s*"(\d+)s?"`)

// doRequest sends a Gemini API request with retry on 429 rate limit errors.
// It parses the retryDelay from the response body and waits accordingly.
func (c *Client) doRequest(ctx context.Context, apiURL string, jsonBody []byte) ([]byte, error) {
	fullURL := apiURL
	isAnthropic := strings.Contains(apiURL, "anthropic.com")
	if !isAnthropic && !strings.Contains(apiURL, "key=") {
		connector := "?"
		if strings.Contains(apiURL, "?") {
			connector = "&"
		}
		fullURL = apiURL + connector + "key=" + c.googleKey
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("create API request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if isAnthropic {
			req.Header.Set("x-api-key", c.anthropicKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("API call: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			delay := parseRetryDelay(body)
			log.Printf("API rate limited, retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		return body, nil
	}
	return nil, fmt.Errorf("API: max retries exceeded")
}

// generateJSON sends parts to Gemini, retries on 429, and returns the text from the first candidate.
func (c *Client) generateJSON(ctx context.Context, parts []part) (string, error) {
	messages := []anthropicMessage{
		{
			Role:    "user",
			Content: make([]anthropicContent, 0, len(parts)),
		},
	}

	for _, p := range parts {
		if p.Text != "" {
			messages[0].Content = append(messages[0].Content, anthropicContent{
				Type: "text",
				Text: p.Text,
			})
		}
		if p.InlineData != nil {
			messages[0].Content = append(messages[0].Content, anthropicContent{
				Type: "image",
				Source: &anthropicSource{
					Type:      "base64",
					MediaType: p.InlineData.MimeType,
					Data:      p.InlineData.Data,
				},
			})
		}
	}

	reqBody := anthropicRequest{
		Model:     modelVersion,
		MaxTokens: 4096,
		Messages:  messages,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal Claude request: %w", err)
	}

	body, err := c.doRequest(ctx, baseURL, jsonBody)
	if err != nil {
		return "", err
	}

	var genResp anthropicResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", fmt.Errorf("unmarshal Claude response: %w", err)
	}

	if len(genResp.Content) == 0 {
		return "", fmt.Errorf("Claude returned no content")
	}

	return stripMarkdown(genResp.Content[0].Text), nil
}

func stripMarkdown(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		// Find the first newline
		if firstNewline := strings.Index(text, "\n"); firstNewline != -1 {
			text = text[firstNewline+1:]
		}
		// Find the last ```
		if lastBackticks := strings.LastIndex(text, "```"); lastBackticks != -1 {
			text = text[:lastBackticks]
		}
	}
	return strings.TrimSpace(text)
}

func parseRetryDelay(body []byte) time.Duration {
	matches := retryDelayRegex.FindSubmatch(body)
	if len(matches) >= 2 {
		if secs, err := strconv.Atoi(string(matches[1])); err == nil {
			return time.Duration(secs)*time.Second + time.Second // add 1s buffer
		}
	}
	return 5 * time.Second // default fallback
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
	SourceLang     string                 `json:"source_lang"`
	ProfileScore   int                    `json:"profile_score"`
	Fields         FlexibleFields         `json:"fields"`
	Skills         LangStringSlice        `json:"skills"`
	Certifications LangStringSlice        `json:"certifications"`
	Languages      []ParsedLanguageItem   `json:"languages"`
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

// ParsedCompanyFull is the result from Gemini when parsing free-form company description text.
type ParsedCompanyFull struct {
	SourceLang      string                       `json:"source_lang"`
	Fields          map[string]map[string]string `json:"fields"`
	EmployeeCount   int32                        `json:"employee_count"`
	Country         string                       `json:"country"`
	Address         string                       `json:"address"`
	Phone           string                       `json:"phone"`
	Telegram        string                       `json:"telegram"`
	TelegramChannel string                       `json:"telegram_channel"`
	Email           string                       `json:"email"`
	WebSite         string                       `json:"web_site"`
	Instagram       string                       `json:"instagram"`
}

func (c *Client) ParseCompanyFromText(ctx context.Context, userInput string) (*ParsedCompanyFull, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildCompanyParsePrompt(userInput)}})
	if err != nil {
		return nil, err
	}
	var parsed ParsedCompanyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini company full JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) ParseCompanyFromFile(ctx context.Context, fileData []byte, mimeType string) (*ParsedCompanyFull, error) {
	encoded := base64.StdEncoding.EncodeToString(fileData)
	prompt := buildCompanyParsePrompt("(see attached file)")
	text, err := c.generateJSON(ctx, []part{
		{Text: prompt},
		{InlineData: &inlineData{MimeType: mimeType, Data: encoded}},
	})
	if err != nil {
		return nil, err
	}
	var parsed ParsedCompanyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini company file JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) TranslateCompany(ctx context.Context, input string) (*ParsedCompany, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildCompanyPrompt(input)}})
	if err != nil {
		return nil, err
	}
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
	SourceLang     string                       `json:"source_lang"`
	Fields         map[string]map[string]string `json:"fields"`
	SalaryMin      int32                        `json:"salary_min"`
	SalaryMax      int32                        `json:"salary_max"`
	SalaryCurrency string                       `json:"salary_currency"`
	ExperienceMin  int32                        `json:"experience_min"`
	ExperienceMax  int32                        `json:"experience_max"`
	Format         string                       `json:"format"`
	Schedule       string                       `json:"schedule"`
	Phone          string                       `json:"phone"`
	Telegram       string                       `json:"telegram"`
	Email          string                       `json:"email"`
	Address        string                       `json:"address"`
	Skills         []string                     `json:"skills"`
}

func (c *Client) TranslateVacancy(ctx context.Context, input string) (*ParsedVacancy, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildVacancyPrompt(input)}})
	if err != nil {
		return nil, err
	}
	var parsed ParsedVacancy
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) ParseVacancyFromText(ctx context.Context, userInput string) (*ParsedVacancyFull, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildVacancyParsePrompt(userInput)}})
	if err != nil {
		return nil, err
	}
	var parsed ParsedVacancyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy full JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) MergeVacancy(ctx context.Context, existingJSON, additionalInfo string) (*ParsedVacancyFull, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildVacancyMergePrompt(existingJSON, additionalInfo)}})
	if err != nil {
		return nil, err
	}
	var parsed ParsedVacancyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy merge JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) ParseVacancyFromFile(ctx context.Context, fileData []byte, mimeType string) (*ParsedVacancyFull, error) {
	encoded := base64.StdEncoding.EncodeToString(fileData)
	prompt := buildVacancyParsePrompt("(see attached file)")
	text, err := c.generateJSON(ctx, []part{
		{Text: prompt},
		{InlineData: &inlineData{MimeType: mimeType, Data: encoded}},
	})
	if err != nil {
		return nil, err
	}
	var parsed ParsedVacancyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini vacancy file JSON: %w (raw: %s)", err, text)
	}
	return &parsed, nil
}

func (c *Client) EnhanceVacancyDescription(ctx context.Context, draftJSON string) (*ParsedVacancyFull, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildVacancyEnhancePrompt(draftJSON)}})
	if err != nil {
		return nil, err
	}
	var parsed ParsedVacancyFull
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini enhanced vacancy JSON: %w (raw: %s)", err, text)
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

	body, err := c.doRequest(ctx, embeddingURL, jsonBody)
	if err != nil {
		return nil, err
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
	resultText, err := c.generateJSON(ctx, []part{{Text: buildTranslateToEnglishPrompt(text)}})
	if err != nil {
		return "", err
	}
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
	resultText, err := c.generateJSON(ctx, []part{{Text: buildTranslateTextPrompt(text)}})
	if err != nil {
		return nil, err
	}
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

// SalaryEstimation is the result from Gemini salary estimation.
type SalaryEstimation struct {
	SalaryMin int32  `json:"salary_min"`
	SalaryMax int32  `json:"salary_max"`
	Currency  string `json:"currency"`
}

func (c *Client) EstimateSalary(ctx context.Context, profileSummary, country string) (*SalaryEstimation, error) {
	text, err := c.generateJSON(ctx, []part{{Text: buildSalaryEstimationPrompt(profileSummary, country)}})
	if err != nil {
		return nil, err
	}
	var result SalaryEstimation
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse gemini salary JSON: %w (raw: %s)", err, text)
	}
	return &result, nil
}

func (c *Client) callGemini(ctx context.Context, parts []part) (*ParsedProfile, error) {
	text, err := c.generateJSON(ctx, parts)
	if err != nil {
		return nil, err
	}
	var profile ParsedProfile
	if err := json.Unmarshal([]byte(text), &profile); err != nil {
		return nil, fmt.Errorf("parse gemini profile JSON: %w (raw: %s)", err, text)
	}
	return &profile, nil
}

// Claude (Anthropic) API types

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type   string           `json:"type"`
	Text   string           `json:"text,omitempty"`
	Source *anthropicSource `json:"source,omitempty"`
}

type anthropicSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
}

// Internal part structure (kept for compatibility with existing code)
type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}
