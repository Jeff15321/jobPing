export interface Job {
  id: string
  title: string
  company: string
  location: string
  description: string
  url: string
  posted_at: string
  fetched_at: string
  ai_analysis?: {
    company_reputation?: string
    benefits?: string[]
    work_culture?: string
    perks?: string[]
    summary?: string
  }
  created_at: string
}

export interface JobsResponse {
  jobs: Job[]
  count: number
}
