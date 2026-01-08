package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jobping/backend/internal/features/job/service"
)

type JobHandler struct {
	service   *service.JobService
	jobspyURL string
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	jobspyURL := os.Getenv("JOBSPY_URL")
	if jobspyURL == "" {
		jobspyURL = "http://localhost:8081" // default for local dev
	}
	return &JobHandler{
		service:   svc,
		jobspyURL: jobspyURL,
	}
}

// GetJobs returns all processed jobs with AI analysis
func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	jobs, err := h.service.GetJobs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch jobs")
		return
	}

	writeJSON(w, http.StatusOK, ToJobsResponse(jobs))
}

// ProcessJob accepts a job via HTTP and runs AI analysis (for local pipeline testing)
func (h *JobHandler) ProcessJob(w http.ResponseWriter, r *http.Request) {
	var input service.JobInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job, err := h.service.ProcessJob(r.Context(), &input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process job")
		return
	}

	if job == nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "job already exists"})
		return
	}

	writeJSON(w, http.StatusCreated, ToJobResponse(*job))
}

// FetchJobs proxies the request to Python JobSpy service to fetch real jobs
func (h *JobHandler) FetchJobs(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	// Create proxy request to Python Flask
	proxyURL := h.jobspyURL + "/fetch"
	log.Printf("Proxying fetch request to: %s", proxyURL)

	client := &http.Client{Timeout: 120 * time.Second} // JobSpy can be slow
	req, err := http.NewRequestWithContext(r.Context(), "POST", proxyURL, bytes.NewReader(body))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create proxy request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		writeError(w, http.StatusBadGateway, "failed to reach job fetcher service")
		return
	}
	defer resp.Body.Close()

	// Read and forward the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read proxy response")
		return
	}

	// Forward headers and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// DeleteAllJobs removes all jobs from the database (for testing)
func (h *JobHandler) DeleteAllJobs(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllJobs(r.Context()); err != nil {
		log.Printf("Failed to delete jobs: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete jobs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "All jobs deleted successfully",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"code": status, "message": message})
}

