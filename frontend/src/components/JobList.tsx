import { Job } from '../types/job'
import { JobCard } from './JobCard'
import './JobList.css'

interface JobListProps {
  jobs: Job[]
}

export function JobList({ jobs }: JobListProps) {
  if (jobs.length === 0) {
    return (
      <div className="empty-state">
        <p>No jobs found. The scanner will fetch new jobs every 10 minutes.</p>
      </div>
    )
  }

  return (
    <div className="job-list">
      {jobs.map((job) => (
        <JobCard key={job.id} job={job} />
      ))}
    </div>
  )
}
