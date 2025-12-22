import { Job, JobsResponse } from '../types/job'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

class JobService {
  async getJobs(limit: number = 10): Promise<Job[]> {
    const response = await fetch(`${API_BASE_URL}/jobs?limit=${limit}`)
    
    if (!response.ok) {
      throw new Error(`Failed to fetch jobs: ${response.statusText}`)
    }
    
    const data: JobsResponse = await response.json()
    return data.jobs || []
  }

  async getJob(id: string): Promise<Job> {
    const response = await fetch(`${API_BASE_URL}/jobs/${id}`)
    
    if (!response.ok) {
      throw new Error(`Failed to fetch job: ${response.statusText}`)
    }
    
    return response.json()
  }

  async scanJobs(): Promise<{ message: string; fetched: number; stored: number; jobs: Job[] }> {
    const response = await fetch(`${API_BASE_URL}/jobs/scan`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    })
    
    if (!response.ok) {
      throw new Error(`Failed to scan jobs: ${response.statusText}`)
    }
    
    return response.json()
  }
}

export const jobService = new JobService()
