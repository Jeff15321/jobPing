import { useEffect, useState } from 'react'
import './App.css'
import { api, Job } from './services/api'

function App() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  useEffect(() => {
    loadJobs()
  }, [])

  const loadJobs = async () => {
    try {
      setLoading(true)
      setError(null)
      const fetchedJobs = await api.getJobs(20)
      setJobs(fetchedJobs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load jobs')
    } finally {
      setLoading(false)
    }
  }

  const handleFetchJobs = async () => {
    try {
      setFetching(true)
      setError(null)
      setMessage(null)
      const result = await api.fetchJobs()
      setMessage(`${result.message}`)
      // Wait a moment for processing, then reload
      setTimeout(() => loadJobs(), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch jobs')
    } finally {
      setFetching(false)
    }
  }

  const getScoreColor = (score?: number) => {
    if (!score) return 'var(--text-muted)'
    if (score >= 80) return 'var(--success)'
    if (score >= 60) return 'var(--warning)'
    return 'var(--danger)'
  }

  return (
    <div className="app">
      <header className="header">
        <h1>🔍 JobPing</h1>
        <p>AI-Powered Job Scanner</p>
      </header>

      <main className="main">
        <div className="actions">
          <button 
            onClick={handleFetchJobs} 
            disabled={fetching}
            className="fetch-btn"
          >
            {fetching ? 'Fetching...' : '🚀 Fetch Latest Jobs'}
          </button>
          <button 
            onClick={loadJobs} 
            disabled={loading}
            className="refresh-btn"
          >
            {loading ? 'Loading...' : '🔄 Refresh'}
          </button>
        </div>

        {message && <div className="message success">{message}</div>}
        {error && <div className="message error">{error}</div>}

        {loading && jobs.length === 0 && (
          <div className="loading">Loading jobs...</div>
        )}

        {jobs.length === 0 && !loading && (
          <div className="empty">
            No jobs found. Click "Fetch Latest Jobs" to scrape new jobs!
          </div>
        )}

        <div className="jobs-list">
          {jobs.map((job) => (
            <div key={job.id} className="job-card">
              <div className="job-header">
                <h3 className="job-title">{job.title}</h3>
                {job.ai_score !== undefined && (
                  <span 
                    className="job-score"
                    style={{ color: getScoreColor(job.ai_score) }}
                  >
                    {job.ai_score}%
                  </span>
                )}
              </div>
              
              <div className="job-company">{job.company}</div>
              
              <div className="job-meta">
                <span className="job-location">📍 {job.location || 'Unknown'}</span>
                {job.is_remote && <span className="job-remote">🏠 Remote</span>}
                {job.job_type && <span className="job-type">{job.job_type}</span>}
              </div>

              {(job.min_salary || job.max_salary) && (
                <div className="job-salary">
                  💰 {job.min_salary && `$${job.min_salary.toLocaleString()}`}
                  {job.min_salary && job.max_salary && ' - '}
                  {job.max_salary && `$${job.max_salary.toLocaleString()}`}
                </div>
              )}

              {job.ai_analysis && (
                <div className="job-analysis">
                  <strong>AI Analysis:</strong> {job.ai_analysis}
                </div>
              )}

              <a 
                href={job.job_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="job-link"
              >
                View Job →
              </a>
            </div>
          ))}
        </div>
      </main>
    </div>
  )
}

export default App
