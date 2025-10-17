export interface Player {
  id: string;
  name: string;
  email: string;
  active: boolean;
  established: boolean;
  created_at: string;
}

export interface Course {
  id: string;
  name: string;
  par: number;
  course_rating: number;
  slope_rating: number;
  hole_handicaps: number[];
  hole_pars: number[];
}

export interface Round {
  id: string;
  player_id: string;
  date: string;
  course_id: string;
  gross_scores: number[];
  adjusted_gross_scores: number[];
  total_gross: number;
  total_adjusted: number;
}

export interface HandicapRecord {
  id: string;
  player_id: string;
  league_handicap: number;
  course_handicap: number;
  playing_handicap: number;
  updated_at: string;
}

export interface Match {
  id: string;
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
