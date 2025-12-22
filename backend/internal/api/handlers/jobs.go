package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/ai-job-scanner/internal/database"
	"github.com/yourusername/ai-job-scanner/internal/integrations/jobspy"
)

type JobHandler struct {
	db      *database.DB
	fetcher *jobspy.Client
}

func NewJobHandler(db *database.DB, fetcher *jobspy.Client) *JobHandler {
	return &JobHandler{db: db, fetcher: fetcher}
}

func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	jobs, err := database.GetJobs(h.db, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	job, err := database.GetJobByID(h.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
func (h *JobHandler) ScanJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Fetch jobs from SpeedyApply API
	jobs, err := h.fetcher.FetchLatestJobs(ctx, 10)
	if err != nil {
		http.Error(w, "Failed to fetch jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store jobs in database
	storedCount := 0
	for _, job := range jobs {
		if err := database.UpsertJob(h.db, job); err != nil {
			// Log error but continue with other jobs
			continue
		}
		storedCount++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Scan completed successfully",
		"fetched":      len(jobs),
		"stored":       storedCount,
		"jobs":         jobs,
	})
}