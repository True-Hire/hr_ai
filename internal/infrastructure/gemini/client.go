package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL      = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
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

// ParsedProfile is the structured result from Gemini profile parsing.
type ParsedProfile struct {
	SourceLang string                       `json:"source_lang"`
	Fields     map[string]map[string]string `json:"fields"`
	Experience []ParsedExperienceItem       `json:"experience"`
	Education  []ParsedEducationItem        `json:"education"`
}

type ParsedExperienceItem struct {
	Company     string            `json:"company"`
	Position    map[string]string `json:"position"`
	StartDate   string            `json:"start_date"`
	EndDate     string            `json:"end_date"`
	Projects    string            `json:"projects"`
	WebSite     string            `json:"web_site"`
	Description map[string]string `json:"description"`
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
