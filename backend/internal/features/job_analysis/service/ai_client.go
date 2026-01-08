package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AIClient interface for AI analysis (ChatGPT)
type AIClient interface {
	ResearchCompany(ctx context.Context, company, title, description string) (map[string]interface{}, error)
}

type openAIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAIClient() AIClient {
	return &openAIClient{
		apiKey:     os.Getenv("OPENAI_API_KEY"),
		httpClient: &http.Client{},
	}
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *openAIClient) ResearchCompany(ctx context.Context, company, title, description string) (map[string]interface{}, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	prompt := fmt.Sprintf(`Research and analyze this company for a job seeker. Provide factual information if you know it, otherwise indicate uncertainty.

Company: %s
Job Title: %s
Job Description (partial): %s

Provide a JSON response with these fields:
{
  "company_size": "estimated employee count or range",
  "industry": "primary industry",
  "culture": "brief description of work culture if known",
  "funding": "funding status/stage if known",
  "notable_info": "any notable facts (acquisitions, layoffs, growth, etc.)",
  "tech_stack": "common technologies used if known",
  "work_life_balance": "reputation for work-life balance if known",
  "red_flags": ["any concerning patterns"],
  "green_flags": ["positive indicators"]
}`, company, title, truncate(description, 500))

	result, err := c.callOpenAI(ctx, prompt, "You are a company research assistant. Provide factual, balanced information about companies.")
	if err != nil {
		return nil, err
	}

	var companyInfo map[string]interface{}
	if err := json.Unmarshal([]byte(result), &companyInfo); err != nil {
		// Return raw content if JSON parsing fails
		return map[string]interface{}{
			"raw_analysis": result,
		}, nil
	}

	return companyInfo, nil
}

func (c *openAIClient) callOpenAI(ctx context.Context, prompt, systemPrompt string) (string, error) {
	reqBody := openAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api returned status %d", resp.StatusCode)
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
