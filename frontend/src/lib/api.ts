import type {
    League,
    LeagueMember,
    LeagueMemberWithPlayer,
    Player,
    Course,
    Season,
    Match,
    Score,
    Round,
    HandicapRecord,
    StandingsEntry,
    BulletinMessage,
    UserInfo,
    CreateLeagueRequest,
    CreateCourseRequest,
    CreatePlayerRequest,
    CreateSeasonRequest,
    CreateMatchRequest,
    CreateScoreRequest,
    CreateRoundRequest,
    LinkPlayerRequest,
    MatchDay,
    CreateMatchDayRequest,
    ScoreSubmission,
    MatchDayScoresResponse,
    ScoreEntryResponse,
    SeasonPlayer,
    SeasonPlayerWithPlayer,
    LeagueInvite,
    InviteDetails,
    AcceptInviteResponse,
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

    // League endpoints
    async createLeague(data: CreateLeagueRequest): Promise<League> {
        return this.request<League>('/api/leagues', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listLeagues(): Promise<League[]> {
        return this.request<League[]>('/api/leagues');
    }

    async getLeague(id: string): Promise<League> {
        return this.request<League>(`/api/leagues/${id}`);
    }

    async updateLeague(id: string, data: Partial<League>): Promise<League> {
        return this.request<League>(`/api/leagues/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    // League Member endpoints
    async addLeagueMember(leagueId: string, email: string, name?: string, provisionalHandicap?: number): Promise<LeagueMember> {
        return this.request<LeagueMember>(`/api/leagues/${leagueId}/members`, {
            method: 'POST',
            body: JSON.stringify({ email, name, provisionalHandicap: provisionalHandicap || 0 }),
        });
    }

    async listLeagueMembers(leagueId: string): Promise<LeagueMemberWithPlayer[]> {
        return this.request<LeagueMemberWithPlayer[]>(`/api/leagues/${leagueId}/members`);
    }

    async updateLeagueMemberRole(leagueId: string, playerId: string, role: 'admin' | 'player'): Promise<LeagueMember> {
        return this.request<LeagueMember>(`/api/leagues/${leagueId}/members/${playerId}`, {
            method: 'PUT',
            body: JSON.stringify({ role }),
        });
    }

    async updateLeagueMember(leagueId: string, playerId: string, data: { role?: 'admin' | 'player'; provisionalHandicap?: number }): Promise<LeagueMember> {
        return this.request<LeagueMember>(`/api/leagues/${leagueId}/members/${playerId}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async removeLeagueMember(leagueId: string, playerId: string): Promise<void> {
        return this.request<void>(`/api/leagues/${leagueId}/members/${playerId}`, {
            method: 'DELETE',
        });
    }

    // Season Player endpoints
    async addSeasonPlayer(leagueId: string, seasonId: string, playerId: string, provisionalHandicap?: number): Promise<SeasonPlayer> {
        return this.request<SeasonPlayer>(`/api/leagues/${leagueId}/seasons/${seasonId}/players`, {
            method: 'POST',
            body: JSON.stringify({ playerId, provisionalHandicap: provisionalHandicap || 0 }),
        });
    }

    async listSeasonPlayers(leagueId: string, seasonId: string): Promise<SeasonPlayerWithPlayer[]> {
        return this.request<SeasonPlayerWithPlayer[]>(`/api/leagues/${leagueId}/seasons/${seasonId}/players`);
    }

    async updateSeasonPlayer(leagueId: string, seasonId: string, playerId: string, data: { provisionalHandicap?: number }): Promise<SeasonPlayer> {
        return this.request<SeasonPlayer>(`/api/leagues/${leagueId}/seasons/${seasonId}/players/${playerId}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async removeSeasonPlayer(leagueId: string, seasonId: string, playerId: string): Promise<void> {
        return this.request<void>(`/api/leagues/${leagueId}/seasons/${seasonId}/players/${playerId}`, {
            method: 'DELETE',
        });
    }

    // Course endpoints
    async createCourse(leagueId: string, data: CreateCourseRequest): Promise<Course> {
        return this.request<Course>(`/api/leagues/${leagueId}/courses`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listCourses(leagueId: string): Promise<Course[]> {
        return this.request<Course[]>(`/api/leagues/${leagueId}/courses`);
    }

    async getCourse(leagueId: string, id: string): Promise<Course> {
        return this.request<Course>(`/api/leagues/${leagueId}/courses/${id}`);
    }

    async updateCourse(leagueId: string, id: string, data: Partial<Course>): Promise<Course> {
        return this.request<Course>(`/api/leagues/${leagueId}/courses/${id}`, {
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
    async createSeason(leagueId: string, data: CreateSeasonRequest): Promise<Season> {
        return this.request<Season>(`/api/leagues/${leagueId}/seasons`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listSeasons(leagueId: string): Promise<Season[]> {
        return this.request<Season[]>(`/api/leagues/${leagueId}/seasons`);
    }

    async getSeason(leagueId: string, id: string): Promise<Season> {
        return this.request<Season>(`/api/leagues/${leagueId}/seasons/${id}`);
    }

    async updateSeason(leagueId: string, id: string, data: Partial<Season>): Promise<Season> {
        return this.request<Season>(`/api/leagues/${leagueId}/seasons/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async getActiveSeason(leagueId: string): Promise<Season> {
        return this.request<Season>(`/api/leagues/${leagueId}/seasons/active`);
    }

    async getSeasonMatches(leagueId: string, seasonId: string): Promise<Match[]> {
        return this.request<Match[]>(`/api/leagues/${leagueId}/seasons/${seasonId}/matches`);
    }

    // Match endpoints
    async createMatch(leagueId: string, data: CreateMatchRequest): Promise<Match> {
        return this.request<Match>(`/api/leagues/${leagueId}/matches`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listMatches(leagueId: string, status?: string): Promise<Match[]> {
        const query = status ? `?status=${status}` : '';
        return this.request<Match[]>(`/api/leagues/${leagueId}/matches${query}`);
    }

    async getMatch(leagueId: string, id: string): Promise<Match> {
        return this.request<Match>(`/api/leagues/${leagueId}/matches/${id}`);
    }

    async updateMatch(leagueId: string, id: string, data: Partial<Match>): Promise<Match> {
        return this.request<Match>(`/api/leagues/${leagueId}/matches/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    // Match Day endpoints
    async createMatchDay(leagueId: string, data: CreateMatchDayRequest): Promise<MatchDay> {
        return this.request<MatchDay>(`/api/leagues/${leagueId}/match-days`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async listMatchDays(leagueId: string): Promise<MatchDay[]> {
        return this.request<MatchDay[]>(`/api/leagues/${leagueId}/match-days`);
    }

    async getMatchDay(leagueId: string, id: string): Promise<MatchDay> {
        return this.request<MatchDay>(`/api/leagues/${leagueId}/match-days/${id}`);
    }

    async updateMatchDay(leagueId: string, id: string, data: { date?: string; courseId?: string }): Promise<MatchDay> {
        return this.request<MatchDay>(`/api/leagues/${leagueId}/match-days/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async deleteMatchDay(leagueId: string, id: string): Promise<{ status: string }> {
        return this.request<{ status: string }>(`/api/leagues/${leagueId}/match-days/${id}`, {
            method: 'DELETE',
        });
    }

    async updateMatchDayMatchups(leagueId: string, matchDayId: string, matches: { id?: string; playerAId: string; playerBId: string }[]): Promise<{ matchDay: MatchDay; matches: Match[] }> {
        return this.request<{ matchDay: MatchDay; matches: Match[] }>(`/api/leagues/${leagueId}/match-days/${matchDayId}/matchups`, {
            method: 'PUT',
            body: JSON.stringify({ matches }),
        });
    }

    async getMatchDayScores(leagueId: string, matchDayId: string): Promise<MatchDayScoresResponse> {
        return this.request<MatchDayScoresResponse>(`/api/leagues/${leagueId}/match-days/${matchDayId}/scores`);
    }

    async enterMatchDayScores(leagueId: string, matchDayId: string, scores: ScoreSubmission[]): Promise<ScoreEntryResponse> {
        return this.request<ScoreEntryResponse>(`/api/leagues/${leagueId}/match-days/scores`, {
            method: 'POST',
            body: JSON.stringify({ matchDayId, scores }),
        });
    }

    // Score endpoints
    async enterScore(leagueId: string, data: CreateScoreRequest): Promise<Score> {
        return this.request<Score>(`/api/leagues/${leagueId}/scores`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getMatchScores(leagueId: string, matchId: string): Promise<Score[]> {
        return this.request<Score[]>(`/api/leagues/${leagueId}/matches/${matchId}/scores`);
    }

    async enterScoreBatch(leagueId: string, scores: CreateScoreRequest[]): Promise<void> {
        return this.request<void>(`/api/leagues/${leagueId}/scores/batch`, {
            method: 'POST',
            body: JSON.stringify({ scores }),
        });
    }

    // Round endpoints
    async createRound(leagueId: string, data: CreateRoundRequest): Promise<Round> {
        return this.request<Round>(`/api/leagues/${leagueId}/rounds`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getPlayerRounds(leagueId: string, playerId: string): Promise<Round[]> {
        return this.request<Round[]>(`/api/leagues/${leagueId}/players/${playerId}/rounds`);
    }

    async getPlayerScores(leagueId: string, playerId: string): Promise<Score[]> {
        return this.request<Score[]>(`/api/leagues/${leagueId}/players/${playerId}/scores`);
    }

    // Handicap endpoints
    async getPlayerHandicap(leagueId: string, playerId: string): Promise<HandicapRecord> {
        return this.request<HandicapRecord>(`/api/leagues/${leagueId}/players/${playerId}/handicap`);
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
    async getStandings(leagueId: string): Promise<StandingsEntry[]> {
        return this.request<StandingsEntry[]>(`/api/leagues/${leagueId}/standings`);
    }

    // Bulletin board endpoints
    async listBulletinMessages(leagueId: string, seasonId: string, limit?: number): Promise<BulletinMessage[]> {
        const query = limit ? `?limit=${limit}` : '';
        return this.request<BulletinMessage[]>(`/api/leagues/${leagueId}/seasons/${seasonId}/bulletin${query}`);
    }

    async createBulletinMessage(leagueId: string, seasonId: string, content: string): Promise<BulletinMessage> {
        return this.request<BulletinMessage>(`/api/leagues/${leagueId}/seasons/${seasonId}/bulletin`, {
            method: 'POST',
            body: JSON.stringify({ content }),
        });
    }

    async deleteBulletinMessage(leagueId: string, messageId: string): Promise<{ status: string }> {
        return this.request<{ status: string }>(`/api/leagues/${leagueId}/bulletin/${messageId}`, {
            method: 'DELETE',
        });
    }

    // Job endpoints
    async recalculateHandicaps(leagueId: string): Promise<{ status: string }> {
        return this.request<{ status: string }>(`/api/leagues/${leagueId}/jobs/recalculate-handicaps`, {
            method: 'POST',
        });
    }

    async processMatch(leagueId: string, matchId: string): Promise<{ status: string }> {
        return this.request<{ status: string }>(`/api/leagues/${leagueId}/jobs/process-match/${matchId}`, {
            method: 'POST',
        });
    }

    // League invite endpoints
    async createLeagueInvite(leagueId: string, expiresInDays?: number, maxUses?: number): Promise<LeagueInvite> {
        return this.request<LeagueInvite>(`/api/leagues/${leagueId}/invites`, {
            method: 'POST',
            body: JSON.stringify({ expiresInDays, maxUses }),
        });
    }

    async listLeagueInvites(leagueId: string): Promise<LeagueInvite[]> {
        return this.request<LeagueInvite[]>(`/api/leagues/${leagueId}/invites`);
    }

    async revokeLeagueInvite(leagueId: string, inviteId: string): Promise<void> {
        return this.request<void>(`/api/leagues/${leagueId}/invites/${inviteId}`, {
            method: 'DELETE',
        });
    }

    async getInviteByToken(token: string): Promise<InviteDetails> {
        return this.request<InviteDetails>(`/api/invites/${token}`);
    }

    async acceptInvite(token: string): Promise<AcceptInviteResponse> {
        return this.request<AcceptInviteResponse>(`/api/invites/${token}/accept`, {
            method: 'POST',
        });
    }
}

export const api = new APIClient(API_URL);
export default api;
