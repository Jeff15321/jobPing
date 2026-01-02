package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AIClient interface for AI analysis (OpenAI)
type AIClient interface {
	AnalyzeJob(ctx context.Context, title, company, description string) (*AIAnalysisResult, error)
}

type AIAnalysisResult struct {
	Score    int    `json:"score"`
	Analysis string `json:"analysis"`
}

type openAIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIClient() AIClient {
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

func (c *openAIClient) AnalyzeJob(ctx context.Context, title, company, description string) (*AIAnalysisResult, error) {
	if c.apiKey == "" {
		// Return mock result if no API key (for local testing)
		return &AIAnalysisResult{
			Score:    75,
			Analysis: "Mock analysis: This job appears to be a good fit based on the title and description.",
		}, nil
	}

	prompt := fmt.Sprintf(`Analyze this job posting and provide:
1. A match score from 0-100 (higher = better fit for a software engineer)
2. A brief analysis (2-3 sentences)

Job Title: %s
Company: %s
Description: %s

Respond in JSON format: {"score": number, "analysis": "string"}`, title, company, description)

	reqBody := openAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openAIMessage{
			{Role: "system", Content: "You are a job matching assistant. Analyze jobs for software engineers."},
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned status %d", resp.StatusCode)
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	var result AIAnalysisResult
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &result); err != nil {
		// If parsing fails, use the raw content as analysis
		return &AIAnalysisResult{
			Score:    50,
			Analysis: openAIResp.Choices[0].Message.Content,
		}, nil
	}

	return &result, nil
}

