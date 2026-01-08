package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AIClient interface for AI matching
type AIClient interface {
	MatchJobToUser(ctx context.Context, job *JobMatchInput, userPrompt string) (*UserMatchResult, error)
}

type JobMatchInput struct {
	Title       string
	Company     string
	Description string
	CompanyInfo map[string]interface{}
}

type UserMatchResult struct {
	Score       int                    `json:"score"`
	Explanation string                 `json:"explanation"`
	Pros        []string               `json:"pros"`
	Cons        []string               `json:"cons"`
	Analysis    map[string]interface{} `json:"analysis"`
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

func (c *openAIClient) MatchJobToUser(ctx context.Context, job *JobMatchInput, userPrompt string) (*UserMatchResult, error) {
	if c.apiKey == "" {
		// Return mock result if no API key
		return &UserMatchResult{
			Score:       65,
			Explanation: "Mock match: Unable to analyze without API key. This is a placeholder score.",
			Pros:        []string{"Job exists", "Company exists"},
			Cons:        []string{"No real analysis performed"},
			Analysis: map[string]interface{}{
				"mock": true,
			},
		}, nil
	}

	companyInfoStr := ""
	if job.CompanyInfo != nil {
		infoBytes, _ := json.Marshal(job.CompanyInfo)
		companyInfoStr = string(infoBytes)
	}

	prompt := fmt.Sprintf(`Match this job to a user's preferences and provide a compatibility score.

USER'S PREFERENCES/IDEAL JOB:
%s

JOB DETAILS:
Title: %s
Company: %s
Description: %s

COMPANY RESEARCH:
%s

Analyze how well this job matches the user's preferences. Provide a JSON response:
{
  "score": 0-100 (how well this job matches their preferences),
  "explanation": "2-3 sentence explanation of the match",
  "pros": ["reasons this job is a good fit"],
  "cons": ["reasons this job might not be ideal"],
  "key_match_factors": ["specific factors from user preferences that match"]
}`, userPrompt, job.Title, job.Company, truncate(job.Description, 800), companyInfoStr)

	result, err := c.callOpenAI(ctx, prompt, "You are a job matching assistant. Be honest and balanced in your analysis.")
	if err != nil {
		return nil, err
	}

	var matchResult UserMatchResult
	if err := json.Unmarshal([]byte(result), &matchResult); err != nil {
		return &UserMatchResult{
			Score:       50,
			Explanation: result,
			Analysis:    map[string]interface{}{"parse_error": true, "raw": result},
		}, nil
	}

	// Convert to map for storage
	matchResult.Analysis = map[string]interface{}{
		"score":       matchResult.Score,
		"explanation": matchResult.Explanation,
		"pros":        matchResult.Pros,
		"cons":        matchResult.Cons,
	}

	return &matchResult, nil
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


