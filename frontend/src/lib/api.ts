import type { Player, Course, Match, Round, Score, HandicapRecord, StandingsEntry, Season } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class APIClient {
  private baseURL: string;
  private getToken?: () => Promise<string | null>;

  constructor(baseURL: string, getToken?: () => Promise<string | null>) {
    this.baseURL = baseURL;
    this.getToken = getToken;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options?.headers,
    };

    // Add Authorization header if token getter is available
    if (this.getToken) {
      try {
        const token = await this.getToken();
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }
      } catch (error) {
        console.warn('Failed to get auth token:', error);
      }
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.statusText}`);
    }

    return response.json();
  }

  // Set the token getter function (useful for updating after client initialization)
  setTokenGetter(getToken: () => Promise<string | null>) {
    this.getToken = getToken;
  }

  // Course methods
  async listCourses(): Promise<Course[]> {
    return this.request<Course[]>('/api/admin/courses');
  }

  async getCourse(id: string): Promise<Course> {
    return this.request<Course>(`/api/admin/courses/${id}`);
  }

  async createCourse(course: Omit<Course, 'id'>): Promise<Course> {
    return this.request<Course>('/api/admin/courses', {
      method: 'POST',
      body: JSON.stringify(course),
    });
  }

  async updateCourse(id: string, course: Partial<Course>): Promise<Course> {
    return this.request<Course>(`/api/admin/courses/${id}`, {
      method: 'PUT',
      body: JSON.stringify(course),
    });
  }

  // Player methods
  async listPlayers(activeOnly?: boolean): Promise<Player[]> {
    const query = activeOnly ? '?active=true' : '';
    return this.request<Player[]>(`/api/admin/players${query}`);
  }

  async getPlayer(id: string): Promise<Player> {
    return this.request<Player>(`/api/admin/players/${id}`);
  }

  async createPlayer(player: Omit<Player, 'id' | 'created_at'>): Promise<Player> {
    return this.request<Player>('/api/admin/players', {
      method: 'POST',
      body: JSON.stringify(player),
    });
  }

  async updatePlayer(id: string, player: Partial<Player>): Promise<Player> {
    return this.request<Player>(`/api/admin/players/${id}`, {
      method: 'PUT',
      body: JSON.stringify(player),
    });
  }

  // Season methods
  async listSeasons(): Promise<Season[]> {
    return this.request<Season[]>('/api/admin/seasons');
  }

  async getSeason(id: string): Promise<Season> {
    return this.request<Season>(`/api/admin/seasons/${id}`);
  }

  async createSeason(season: Omit<Season, 'id' | 'created_at'>): Promise<Season> {
    return this.request<Season>('/api/admin/seasons', {
      method: 'POST',
      body: JSON.stringify(season),
    });
  }

  async updateSeason(id: string, season: Partial<Season>): Promise<Season> {
    return this.request<Season>(`/api/admin/seasons/${id}`, {
      method: 'PUT',
      body: JSON.stringify(season),
    });
  }

  async getActiveSeason(): Promise<Season> {
    return this.request<Season>('/api/admin/seasons/active');
  }

  async getSeasonMatches(seasonId: string): Promise<Match[]> {
    return this.request<Match[]>(`/api/admin/seasons/${seasonId}/matches`);
  }

  // Match methods
  async listMatches(status?: string): Promise<Match[]> {
    const query = status ? `?status=${status}` : '';
    return this.request<Match[]>(`/api/admin/matches${query}`);
  }

  async getMatch(id: string): Promise<Match> {
    return this.request<Match>(`/api/admin/matches/${id}`);
  }

  async createMatch(match: Omit<Match, 'id' | 'status'>): Promise<Match> {
    return this.request<Match>('/api/admin/matches', {
      method: 'POST',
      body: JSON.stringify(match),
    });
  }

  async updateMatch(id: string, match: Partial<Match>): Promise<Match> {
    return this.request<Match>(`/api/admin/matches/${id}`, {
      method: 'PUT',
      body: JSON.stringify(match),
    });
  }

  // Score methods
  async enterScore(score: Omit<Score, 'id'>): Promise<Score> {
    return this.request<Score>('/api/admin/scores', {
      method: 'POST',
      body: JSON.stringify(score),
    });
  }

  async getMatchScores(matchId: string): Promise<Score[]> {
    return this.request<Score[]>(`/api/matches/${matchId}/scores`);
  }

  // Round methods
  async createRound(round: Omit<Round, 'id'>): Promise<Round> {
    return this.request<Round>('/api/admin/rounds', {
      method: 'POST',
      body: JSON.stringify(round),
    });
  }

  async getPlayerRounds(playerId: string): Promise<Round[]> {
    return this.request<Round[]>(`/api/players/${playerId}/rounds`);
  }

  // Handicap methods
  async getPlayerHandicap(playerId: string): Promise<HandicapRecord> {
    return this.request<HandicapRecord>(`/api/players/${playerId}/handicap`);
  }

  // Standings
  async getStandings(): Promise<StandingsEntry[]> {
    return this.request<StandingsEntry[]>('/api/standings');
  }

  // Job methods
  async recalculateHandicaps(): Promise<{ status: string }> {
    return this.request<{ status: string }>('/api/jobs/recalculate-handicaps', {
      method: 'POST',
    });
  }

  async processMatch(matchId: string): Promise<{ status: string }> {
    return this.request<{ status: string }>(`/api/jobs/process-match/${matchId}`, {
      method: 'POST',
    });
  }
}

export const api = new APIClient(API_URL);
