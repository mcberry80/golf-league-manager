export interface League {
    id: string;
    name: string;
    description: string;
    createdBy: string;
    createdAt: string;
}

export interface LeagueMember {
    id: string;
    leagueId: string;
    playerId: string;
    role: 'admin' | 'player';
    provisionalHandicap: number; // Starting handicap for the season (Golf League Rules 3.2)
    joinedAt: string;
}

export interface LeagueMemberWithPlayer extends LeagueMember {
    player: Player;
}

export interface Player {
    id: string;
    name: string;
    email: string;
    clerkUserId?: string;
    active: boolean;
    established: boolean;
    createdAt: string;
}

export interface Course {
    id: string;
    leagueId: string;
    name: string;
    par: number;
    courseRating: number;
    slopeRating: number;
    holeHandicaps: number[]; // 9 holes, difficulty rankings 1-9
    holePars: number[]; // 9 holes
}

export interface Round {
    id: string;
    playerId: string;
    leagueId: string;
    date: string;
    courseId: string;
    grossScores: number[]; // 9 holes
    adjustedGrossScores: number[]; // 9 holes
    totalGross: number;
    totalAdjusted: number;
    handicapDifferential: number; // Calculated differential for this round
}

export interface HandicapRecord {
    id: string;
    playerId: string;
    leagueId: string;
    leagueHandicap: number;
    courseHandicap: number;
    playingHandicap: number;
    updatedAt: string;
}

export interface Season {
    id: string;
    leagueId: string;
    name: string;
    startDate: string;
    endDate: string;
    active: boolean;
    description: string;
    createdAt: string;
}

export interface MatchDay {
    id: string;
    leagueId: string;
    seasonId: string;
    date: string;
    courseId: string;
    createdAt: string;
}

export interface Match {
    id: string;
    leagueId: string;
    seasonId: string;
    matchDayId?: string;
    weekNumber: number;
    playerAId: string;
    playerBId: string;
    courseId: string;
    matchDate: string;
    status: 'scheduled' | 'completed';
}

export interface Score {
    id: string;
    matchId: string;
    playerId: string;
    holeScores: number[];
    grossScore: number;
    netScore: number;
    adjustedGross: number;
    handicapDifferential: number;
    strokesReceived: number;
    playerAbsent?: boolean;
}

export interface StandingsEntry {
    playerId: string;
    playerName: string;
    matchesPlayed: number;
    matchesWon: number;
    matchesLost: number;
    matchesTied: number;
    totalPoints: number;
    leagueHandicap: number;
}

export interface UserInfo {
    linked: boolean;
    clerkUserId?: string;
    player?: Player;
    leagues?: LeagueMember[]; // List of leagues the user is a member of
}

// Request/Response types
export interface CreateLeagueRequest {
    name: string;
    description: string;
}

export interface CreateCourseRequest {
    name: string;
    par: number;
    courseRating: number;
    slopeRating: number;
    holeHandicaps: number[];
    holePars: number[];
}

export interface CreatePlayerRequest {
    name: string;
    email: string;
}

export interface CreateSeasonRequest {
    name: string;
    startDate: string;
    endDate: string;
    active: boolean;
    description: string;
}

export interface CreateMatchRequest {
    seasonId: string;
    weekNumber: number;
    playerAId: string;
    playerBId: string;
    courseId: string;
    matchDate: string;
    matchDayId?: string;
}

export interface CreateMatchDayRequest {
    date: string;
    courseId: string;
    seasonId: string;
    matches: Partial<Match>[];
}

export interface ScoreSubmission {
    matchId: string;
    playerId: string;
    holeScores: number[];
}

export interface CreateScoreRequest {
    matchId: string;
    playerId: string;
    holeNumber: number;
    grossScore: number;
    netScore: number;
    strokesReceived: number;
    playerAbsent?: boolean;
}

export interface CreateRoundRequest {
    playerId: string;
    courseId: string;
    date: string;
    grossScores: number[];
}

export interface LinkPlayerRequest {
    email: string;
}
