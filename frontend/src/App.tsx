import { useEffect, useState } from 'react'
import './App.css'
import { api, Preference } from './services/api'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(api.isAuthenticated())
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [preferences, setPreferences] = useState<Preference[]>([])
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isLoginMode, setIsLoginMode] = useState(true)

  useEffect(() => {
    if (isAuthenticated) {
      loadPreferences()
    }
  }, [isAuthenticated])

  const loadPreferences = async () => {
    try {
      setLoading(true)
      const prefs = await api.getPreferences()
      setPreferences(prefs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load preferences')
    } finally {
      setLoading(false)
    }
  }

  const handleAuth = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      if (isLoginMode) {
        await api.login(username, password)
      } else {
        await api.register(username, password)
      }
      setIsAuthenticated(true)
      setUsername('')
      setPassword('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authentication failed')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    api.logout()
    setIsAuthenticated(false)
    setPreferences([])
  }

  const handleAddPreference = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newKey.trim() || !newValue.trim()) return

    try {
      setLoading(true)
      const pref = await api.createPreference(newKey.trim(), newValue.trim())
      setPreferences([pref, ...preferences])
      setNewKey('')
      setNewValue('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add preference')
    } finally {
      setLoading(false)
    }
  }

  const handleDeletePreference = async (id: string) => {
    try {
      await api.deletePreference(id)
      setPreferences(preferences.filter(p => p.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete preference')
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="app">
        <header className="header">
          <h1>🔐 JobPing</h1>
          <p>Sign in to manage your preferences</p>
        </header>

        <main className="main">
          <form onSubmit={handleAuth} className="auth-form">
            <h2>{isLoginMode ? 'Sign In' : 'Create Account'}</h2>
            
            <input
              type="text"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
            
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            
            {error && <div className="error">{error}</div>}
            
            <button type="submit" disabled={loading}>
              {loading ? 'Loading...' : (isLoginMode ? 'Sign In' : 'Create Account')}
            </button>
            
            <p className="toggle-auth">
              {isLoginMode ? "Don't have an account? " : "Already have an account? "}
              <button type="button" onClick={() => setIsLoginMode(!isLoginMode)}>
                {isLoginMode ? 'Sign Up' : 'Sign In'}
              </button>
            </p>
          </form>
        </main>
      </div>
    )
  }

  return (
    <div className="app">
      <header className="header">
        <h1>⚙️ Preferences</h1>
        <button onClick={handleLogout} className="logout-btn">Sign Out</button>
      </header>

      <main className="main">
        <form onSubmit={handleAddPreference} className="add-form">
          <input
            type="text"
            placeholder="Key (e.g., job_title)"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            required
          />
          <input
            type="text"
            placeholder="Value (e.g., Software Engineer)"
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            required
          />
          <button type="submit" disabled={loading}>Add</button>
        </form>

        {error && <div className="error">{error}</div>}

        {loading && preferences.length === 0 && (
          <div className="loading">Loading preferences...</div>
        )}

        {preferences.length === 0 && !loading && (
          <div className="empty">
            No preferences yet. Add your first one above!
          </div>
        )}

        <ul className="preferences-list">
          {preferences.map((pref) => (
            <li key={pref.id} className="preference-item">
              <div className="preference-content">
                <span className="preference-key">{pref.key}</span>
                <span className="preference-value">{pref.value}</span>
              </div>
              <button 
                onClick={() => handleDeletePreference(pref.id)}
                className="delete-btn"
              >
                ✕
              </button>
            </li>
          ))}
        </ul>
      </main>
    </div>
  )
}

export default App
