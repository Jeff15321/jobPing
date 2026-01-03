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
	ResearchCompany(ctx context.Context, company, title, description string) (map[string]interface{}, error)
	MatchJobToUser(ctx context.Context, job *JobMatchInput, userPrompt string) (*UserMatchResult, error)
}

type AIAnalysisResult struct {
	Score    int    `json:"score"`
	Analysis string `json:"analysis"`
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

func (c *openAIClient) ResearchCompany(ctx context.Context, company, title, description string) (map[string]interface{}, error) {
	if c.apiKey == "" {
		// Return mock result if no API key
		return map[string]interface{}{
			"company_size":  "Unknown",
			"industry":      "Technology",
			"culture":       "Unable to determine without API key",
			"funding":       "Unknown",
			"glassdoor":     "N/A",
			"research_note": "Mock data - no OpenAI API key configured",
		}, nil
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
		"score":             matchResult.Score,
		"explanation":       matchResult.Explanation,
		"pros":              matchResult.Pros,
		"cons":              matchResult.Cons,
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
