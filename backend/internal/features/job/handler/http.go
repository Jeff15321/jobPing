package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jobping/backend/internal/features/job/service"
)

type JobHandler struct {
	service *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{service: svc}
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

