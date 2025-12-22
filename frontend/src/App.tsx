import { useEffect, useState } from 'react'
import { JobList } from './components/JobList'
import { Job } from './types/job'
import { jobService } from './services/jobService'
import './App.css'

function App() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [scanning, setScanning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [scanMessage, setScanMessage] = useState<string | null>(null)

  useEffect(() => {
    loadJobs()
  }, [])

  const loadJobs = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await jobService.getJobs(10)
      setJobs(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load jobs')
    } finally {
      setLoading(false)
    }
  }

  const scanJobs = async () => {
    try {
      setScanning(true)
      setError(null)
      setScanMessage(null)
      
      const result = await jobService.scanJobs()
      setScanMessage(`✅ Scanned successfully! Fetched ${result.fetched} jobs, stored ${result.stored} new jobs.`)
      
      // Reload jobs after scanning
      await loadJobs()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to scan jobs')
    } finally {
      setScanning(false)
    }
  }

  return (
    <div className="app">
      <header className="header">
        <h1>🤖 AI Job Scanner</h1>
        <p>Smart job matching powered by AI</p>
      </header>

      <main className="main">
        {loading && <div className="loading">Loading jobs...</div>}
        
        {error && (
          <div className="error">
            <p>Error: {error}</p>
            <button onClick={loadJobs}>Retry</button>
          </div>
        )}

        {!loading && !error && (
          <>
            <div className="stats">
              <span>{jobs.length} jobs found</span>
              <div className="actions">
                <button 
                  onClick={scanJobs} 
                  className="scan-btn"
                  disabled={scanning}
                >
                  {scanning ? '🔄 Scanning...' : '🔍 Scan for Jobs'}
                </button>
                <button onClick={loadJobs} className="refresh-btn">
                  Refresh
                </button>
              </div>
            </div>
            
            {scanMessage && (
              <div className="scan-message">
                {scanMessage}
              </div>
            )}
            
            <JobList jobs={jobs} />
          </>
        )}
      </main>
    </div>
  )
}

export default App
