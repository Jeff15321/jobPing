import { Job } from '../types/job'
import './JobCard.css'

interface JobCardProps {
  job: Job
}

export function JobCard({ job }: JobCardProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  return (
    <div className="job-card">
      <div className="job-header">
        <h3 className="job-title">{job.title}</h3>
        <span className="job-company">{job.company}</span>
      </div>

      <div className="job-meta">
        <span className="job-location">📍 {job.location}</span>
        <span className="job-date">🕒 {formatDate(job.posted_at)}</span>
      </div>

      <p className="job-description">
        {job.description.length > 200
          ? `${job.description.substring(0, 200)}...`
          : job.description}
      </p>

      {job.ai_analysis && (
        <div className="job-ai-analysis">
          <span className="ai-badge">🤖 AI Analysis Available</span>
        </div>
      )}

      <div className="job-actions">
        <a
          href={job.url}
          target="_blank"
          rel="noopener noreferrer"
          className="apply-btn"
        >
          View Job →
        </a>
      </div>
    </div>
  )
}
