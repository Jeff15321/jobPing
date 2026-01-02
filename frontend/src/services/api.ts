const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface AuthResponse {
  token: string;
  user: {
    id: string;
    username: string;
  };
}

interface Preference {
  id: string;
  key: string;
  value: string;
}

interface PreferencesResponse {
  preferences: Preference[];
}

interface Job {
  id: string;
  title: string;
  company: string;
  location: string;
  job_url: string;
  job_type: string;
  is_remote: boolean;
  min_salary?: number;
  max_salary?: number;
  date_posted?: string;
  ai_score?: number;
  ai_analysis?: string;
}

interface JobsResponse {
  jobs: Job[];
}

interface FetchJobsResponse {
  message: string;
  jobs_found: number;
  jobs_queued: number;
}

class ApiService {
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem('token');
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_URL}${path}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    if (response.status === 204) {
      return {} as T;
    }

    return response.json();
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken() {
    return this.token;
  }

  isAuthenticated() {
    return !!this.token;
  }

  // Auth
  async register(username: string, password: string): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/api/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setToken(response.token);
    return response;
  }

  async login(username: string, password: string): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/api/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setToken(response.token);
    return response;
  }

  logout() {
    this.setToken(null);
  }

  // Preferences
  async getPreferences(): Promise<Preference[]> {
    const response = await this.request<PreferencesResponse>('/api/preferences');
    return response.preferences || [];
  }

  async createPreference(key: string, value: string): Promise<Preference> {
    return this.request<Preference>('/api/preferences', {
      method: 'POST',
      body: JSON.stringify({ key, value }),
    });
  }

  async updatePreference(id: string, value: string): Promise<Preference> {
    return this.request<Preference>(`/api/preferences/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ value }),
    });
  }

  async deletePreference(id: string): Promise<void> {
    await this.request<void>(`/api/preferences/${id}`, {
      method: 'DELETE',
    });
  }

  // Jobs
  async fetchJobs(searchTerm?: string, location?: string): Promise<FetchJobsResponse> {
    return this.request<FetchJobsResponse>('/api/jobs/fetch', {
      method: 'POST',
      body: JSON.stringify({
        search_term: searchTerm || 'software engineer',
        location: location || 'San Francisco, CA',
        results_wanted: 5,
      }),
    });
  }

  async getJobs(limit?: number): Promise<Job[]> {
    const response = await this.request<JobsResponse>(`/api/jobs${limit ? `?limit=${limit}` : ''}`);
    return response.jobs || [];
  }
}

export const api = new ApiService();
export type { Preference, AuthResponse, Job, FetchJobsResponse };

