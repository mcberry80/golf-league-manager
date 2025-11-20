import type {
    Player,
    Course,
    Season,
    Match,
    Score,
    Round,
    HandicapRecord,
    StandingsEntry,
    UserInfo,
    CreateCourseRequest,
    CreatePlayerRequest,
    CreateSeasonRequest,
    CreateMatchRequest,
    CreateScoreRequest,
    CreateRoundRequest,
    LinkPlayerRequest,
} from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class APIClient {
    private baseURL: string;
    private getAuthToken: (() => Promise<string | null>) | null = null;

    constructor(baseURL: string) {
        this.baseURL = baseURL;
    }

    setAuthTokenProvider(provider: () => Promise<string | null>) {
        this.getAuthToken = provider;
    }

    private async request<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<T> {
        const token = this.getAuthToken ? await this.getAuthToken() : null;

        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
        };

        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${this.baseURL}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`API Error: ${response.status} - ${errorText}`);
        }

        return response.json();
    }

    // Course endpoints
    async createCourse(data: CreateCourseRequest): Promise<Course> {
        return this.request<Course>('/api/admin/courses', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listCourses(): Promise<Course[]> {
        return this.request<Course[]>('/api/admin/courses');
    }

    async getCourse(id: string): Promise<Course> {
        return this.request<Course>(`/api/admin/courses/${id}`);
    }

    async updateCourse(id: string, data: Partial<Course>): Promise<Course> {
        return this.request<Course>(`/api/admin/courses/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    // Player endpoints
    async createPlayer(data: CreatePlayerRequest): Promise<Player> {
        return this.request<Player>('/api/admin/players', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listPlayers(activeOnly: boolean = false): Promise<Player[]> {
        const query = activeOnly ? '?active=true' : '';
        return this.request<Player[]>(`/api/admin/players${query}`);
    }

    async getPlayer(id: string): Promise<Player> {
        return this.request<Player>(`/api/admin/players/${id}`);
    }

    async updatePlayer(id: string, data: Partial<Player>): Promise<Player> {
        return this.request<Player>(`/api/admin/players/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    // Season endpoints
    async createSeason(data: CreateSeasonRequest): Promise<Season> {
        return this.request<Season>('/api/admin/seasons', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listSeasons(): Promise<Season[]> {
        return this.request<Season[]>('/api/admin/seasons');
    }

    async getSeason(id: string): Promise<Season> {
        return this.request<Season>(`/api/admin/seasons/${id}`);
    }

    async updateSeason(id: string, data: Partial<Season>): Promise<Season> {
        return this.request<Season>(`/api/admin/seasons/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async getActiveSeason(): Promise<Season> {
        return this.request<Season>('/api/admin/seasons/active');
    }

    async getSeasonMatches(seasonId: string): Promise<Match[]> {
        return this.request<Match[]>(`/api/admin/seasons/${seasonId}/matches`);
    }

    // Match endpoints
    async createMatch(data: CreateMatchRequest): Promise<Match> {
        return this.request<Match>('/api/admin/matches', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listMatches(status?: string): Promise<Match[]> {
        const query = status ? `?status=${status}` : '';
        return this.request<Match[]>(`/api/admin/matches${query}`);
    }

    async getMatch(id: string): Promise<Match> {
        return this.request<Match>(`/api/admin/matches/${id}`);
    }

    async updateMatch(id: string, data: Partial<Match>): Promise<Match> {
        return this.request<Match>(`/api/admin/matches/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    // Score endpoints
    async enterScore(data: CreateScoreRequest): Promise<Score> {
        return this.request<Score>('/api/admin/scores', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getMatchScores(matchId: string): Promise<Score[]> {
        return this.request<Score[]>(`/api/matches/${matchId}/scores`);
    }

    // Round endpoints
    async createRound(data: CreateRoundRequest): Promise<Round> {
        return this.request<Round>('/api/admin/rounds', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getPlayerRounds(playerId: string): Promise<Round[]> {
        return this.request<Round[]>(`/api/players/${playerId}/rounds`);
    }

    // Handicap endpoints
    async getPlayerHandicap(playerId: string): Promise<HandicapRecord> {
        return this.request<HandicapRecord>(`/api/players/${playerId}/handicap`);
    }

    // User endpoints
    async linkPlayerAccount(data: LinkPlayerRequest): Promise<Player> {
        return this.request<Player>('/api/user/link-player', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getCurrentUser(): Promise<UserInfo> {
        return this.request<UserInfo>('/api/user/me');
    }

    // Standings endpoints
    async getStandings(): Promise<StandingsEntry[]> {
        return this.request<StandingsEntry[]>('/api/standings');
    }

    // Job endpoints
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
export default api;
