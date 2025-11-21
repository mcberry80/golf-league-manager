export interface League {
    id: string;
    name: string;
    description: string;
    created_by: string;
    created_at: string;
}

export interface LeagueMember {
    id: string;
    league_id: string;
    player_id: string;
    role: 'admin' | 'player';
    joined_at: string;
}

export interface LeagueMemberWithPlayer extends LeagueMember {
    player: Player;
}

export interface Player {
    id: string;
    name: string;
    email: string;
    clerk_user_id?: string;
    active: boolean;
    established: boolean;
    created_at: string;
}

export interface Course {
    id: string;
    league_id: string;
    name: string;
    par: number;
    course_rating: number;
    slope_rating: number;
    hole_handicaps: number[]; // 9 holes, difficulty rankings 1-9
    hole_pars: number[]; // 9 holes
}

export interface Round {
    id: string;
    player_id: string;
    league_id: string;
    date: string;
    course_id: string;
    gross_scores: number[]; // 9 holes
    adjusted_gross_scores: number[]; // 9 holes
    total_gross: number;
    total_adjusted: number;
}

export interface HandicapRecord {
    id: string;
    player_id: string;
    league_id: string;
    league_handicap: number;
    course_handicap: number;
    playing_handicap: number;
    updated_at: string;
}

export interface Season {
    id: string;
    league_id: string;
    name: string;
    start_date: string;
    end_date: string;
    active: boolean;
    description: string;
    created_at: string;
}

export interface Match {
    id: string;
    league_id: string;
    season_id: string;
    week_number: number;
    player_a_id: string;
    player_b_id: string;
    course_id: string;
    match_date: string;
    status: 'scheduled' | 'completed';
}

export interface Score {
    id: string;
    match_id: string;
    player_id: string;
    hole_number: number;
    gross_score: number;
    net_score: number;
    strokes_received: number;
    player_absent?: boolean;
}

export interface StandingsEntry {
    player_id: string;
    player_name: string;
    matches_played: number;
    matches_won: number;
    matches_lost: number;
    matches_tied: number;
    total_points: number;
    league_handicap: number;
}

export interface UserInfo {
    linked: boolean;
    clerk_user_id?: string;
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
    course_rating: number;
    slope_rating: number;
    hole_handicaps: number[];
    hole_pars: number[];
}

export interface CreatePlayerRequest {
    name: string;
    email: string;
}

export interface CreateSeasonRequest {
    name: string;
    start_date: string;
    end_date: string;
    active: boolean;
    description: string;
}

export interface CreateMatchRequest {
    season_id: string;
    week_number: number;
    player_a_id: string;
    player_b_id: string;
    course_id: string;
    match_date: string;
}

export interface CreateScoreRequest {
    match_id: string;
    player_id: string;
    hole_number: number;
    gross_score: number;
    net_score: number;
    strokes_received: number;
    player_absent?: boolean;
}

export interface CreateRoundRequest {
    player_id: string;
    course_id: string;
    date: string;
    gross_scores: number[];
}

export interface LinkPlayerRequest {
    email: string;
}
