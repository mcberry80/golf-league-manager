import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type { MatchDay, Match, Score, LeagueMemberWithPlayer, Course, Season } from '../types'
import { Calendar, ChevronDown, ChevronUp } from 'lucide-react'
import { 
    ExpandableCard, 
    ScorecardTable,
    EMPTY_STROKES_ARRAY 
} from '../components/Scorecard'

// Match day status constants
const MATCH_DAY_STATUS = {
    COMPLETED: 'completed',
    LOCKED: 'locked',
    SCHEDULED: 'scheduled'
} as const

interface MatchWithScores {
    match: Match
    playerAName: string
    playerBName: string
    playerAScore?: Score
    playerBScore?: Score
}

interface MatchDayWithMatches {
    matchDay: MatchDay
    matches: MatchWithScores[]
    courseName: string
}

export default function Results() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()
    
    const [matchDays, setMatchDays] = useState<MatchDay[]>([])
    const [matches, setMatches] = useState<Match[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [seasons, setSeasons] = useState<Season[]>([])
    const [allScores, setAllScores] = useState<Map<string, Score[]>>(new Map())
    
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')
    
    // UI state
    const [selectedSeasonId, setSelectedSeasonId] = useState<string>('')
    const [expandedMatchDayId, setExpandedMatchDayId] = useState<string | null>(null)
    const [expandedMatchId, setExpandedMatchId] = useState<string | null>(null)

    // Load all data
    useEffect(() => {
        async function loadData() {
            const effectiveLeagueId = leagueId || currentLeague?.id
            if (!effectiveLeagueId) return

            try {
                setLoading(true)
                const [matchDaysData, matchesData, membersData, coursesData, seasonsData] = await Promise.all([
                    api.listMatchDays(effectiveLeagueId),
                    api.listMatches(effectiveLeagueId),
                    api.listLeagueMembers(effectiveLeagueId),
                    api.listCourses(effectiveLeagueId),
                    api.listSeasons(effectiveLeagueId)
                ])

                setMatchDays(matchDaysData)
                setMatches(matchesData)
                setMembers(membersData)
                setCourses(coursesData)
                setSeasons(seasonsData)

                // Set default season to active season
                const activeSeason = seasonsData.find(s => s.active)
                if (activeSeason) {
                    setSelectedSeasonId(activeSeason.id)
                } else if (seasonsData.length > 0) {
                    setSelectedSeasonId(seasonsData[0].id)
                }

                // Fetch scores for completed matches in batches to avoid overwhelming the server
                const completedMatches = matchesData.filter(m => m.status === 'completed')
                const scoresMap = new Map<string, Score[]>()
                const BATCH_SIZE = 5
                
                for (let i = 0; i < completedMatches.length; i += BATCH_SIZE) {
                    const batch = completedMatches.slice(i, i + BATCH_SIZE)
                    await Promise.all(
                        batch.map(async (match) => {
                            try {
                                const matchScores = await api.getMatchScores(effectiveLeagueId, match.id)
                                scoresMap.set(match.id, matchScores)
                            } catch (err) {
                                // Log but don't fail the entire page for individual match score errors
                                console.warn(`Failed to load scores for match ${match.id}:`, err)
                            }
                        })
                    )
                }
                
                setAllScores(scoresMap)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load results')
            } finally {
                setLoading(false)
            }
        }

        if (!leagueLoading) {
            loadData()
        }
    }, [leagueId, currentLeague, leagueLoading])

    // Helper to get player name by ID
    const getPlayerName = (playerId: string): string => {
        const member = members.find(m => m.playerId === playerId)
        return member?.player?.name || 'Unknown'
    }

    // Helper to get course by ID
    const getCourse = (courseId: string): Course | undefined => {
        return courses.find(c => c.id === courseId)
    }

    // Format date for display
    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString('en-US', { 
            weekday: 'short',
            year: 'numeric', 
            month: 'short', 
            day: 'numeric' 
        })
    }

    // Get match days grouped by season and sorted by date
    const getMatchDaysByWeek = (): MatchDayWithMatches[] => {
        const filteredMatchDays = selectedSeasonId
            ? matchDays.filter(md => md.seasonId === selectedSeasonId)
            : matchDays

        return filteredMatchDays
            .filter(md => md.status === MATCH_DAY_STATUS.COMPLETED || md.status === MATCH_DAY_STATUS.LOCKED || md.hasScores)
            .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
            .map(matchDay => {
                const matchDayMatches = matches.filter(m => m.matchDayId === matchDay.id)
                const course = getCourse(matchDay.courseId)
                
                const matchesWithScores: MatchWithScores[] = matchDayMatches.map(match => {
                    const matchScores = allScores.get(match.id) || []
                    const playerAScore = matchScores.find(s => s.playerId === match.playerAId)
                    const playerBScore = matchScores.find(s => s.playerId === match.playerBId)
                    
                    return {
                        match,
                        playerAName: getPlayerName(match.playerAId),
                        playerBName: getPlayerName(match.playerBId),
                        playerAScore,
                        playerBScore
                    }
                })

                return {
                    matchDay,
                    matches: matchesWithScores,
                    courseName: course?.name || 'Unknown Course'
                }
            })
    }

    // Calculate points for a match
    const calculateMatchPoints = (
        playerAScore: Score | undefined, 
        playerBScore: Score | undefined,
        match: Match
    ): { playerAHolePoints: number[], playerBHolePoints: number[], playerATotal: number, playerBTotal: number } => {
        const playerAHolePoints: number[] = []
        const playerBHolePoints: number[] = []
        let playerATotal = 0
        let playerBTotal = 0
        
        // Use stored match points when available
        if (match.playerAPoints !== undefined && match.playerBPoints !== undefined) {
            playerATotal = match.playerAPoints
            playerBTotal = match.playerBPoints
        }
        
        if (playerAScore && playerBScore) {
            const playerANetScores = playerAScore.matchNetHoleScores || playerAScore.holeScores
            const playerBNetScores = playerBScore.matchNetHoleScores || playerBScore.holeScores
            
            for (let i = 0; i < 9; i++) {
                const aNet = playerANetScores[i]
                const bNet = playerBNetScores[i]
                
                if (aNet < bNet) {
                    playerAHolePoints.push(2)
                    playerBHolePoints.push(0)
                } else if (aNet > bNet) {
                    playerAHolePoints.push(0)
                    playerBHolePoints.push(2)
                } else {
                    playerAHolePoints.push(1)
                    playerBHolePoints.push(1)
                }
            }
            
            // If no stored totals, calculate them
            if (match.playerAPoints === undefined || match.playerBPoints === undefined) {
                playerATotal = playerAHolePoints.reduce((a, b) => a + b, 0)
                playerBTotal = playerBHolePoints.reduce((a, b) => a + b, 0)
                
                // Add overall match points (4 points for lower total)
                const aMatchNet = playerAScore.matchNetScore ?? playerAScore.netScore
                const bMatchNet = playerBScore.matchNetScore ?? playerBScore.netScore
                
                if (aMatchNet < bMatchNet) {
                    playerATotal += 4
                } else if (aMatchNet > bMatchNet) {
                    playerBTotal += 4
                } else {
                    playerATotal += 2
                    playerBTotal += 2
                }
            }
        }
        
        return { playerAHolePoints, playerBHolePoints, playerATotal, playerBTotal }
    }

    // Build scorecard rows for a match
    const buildScorecardRows = (
        matchData: MatchWithScores,
        course: Course | undefined
    ) => {
        const { match, playerAName, playerBName, playerAScore, playerBScore } = matchData
        
        if (!playerAScore || !playerBScore) return []

        const { playerAHolePoints, playerBHolePoints, playerATotal, playerBTotal } = 
            calculateMatchPoints(playerAScore, playerBScore, match)

        return [
            ...(course?.holePars ? [{ label: 'Par', scores: course.holePars, total: course.par, withBorder: false }] : []),
            ...(course?.holeHandicaps ? [{ label: 'Hole Hdcp', scores: course.holeHandicaps, total: '', withBorder: true }] : []),
            { label: `${playerAName} Gross`, scores: playerAScore.holeScores, total: playerAScore.grossScore, withBorder: false },
            { label: `${playerAName} Strokes`, scores: playerAScore.matchStrokes || EMPTY_STROKES_ARRAY, total: playerAScore.strokesReceived, withBorder: false, color: 'var(--color-accent)' },
            { label: `${playerAName} Net`, scores: playerAScore.matchNetHoleScores || playerAScore.holeScores, total: playerAScore.matchNetScore ?? playerAScore.netScore, withBorder: false, color: 'var(--color-primary)', bgColor: 'rgba(16, 185, 129, 0.1)' },
            { label: `${playerAName} Pts`, scores: playerAHolePoints, total: playerATotal, withBorder: true, color: 'var(--color-primary)', bgColor: 'rgba(16, 185, 129, 0.15)' },
            { label: `${playerBName} Gross`, scores: playerBScore.holeScores, total: playerBScore.grossScore, withBorder: false },
            { label: `${playerBName} Strokes`, scores: playerBScore.matchStrokes || EMPTY_STROKES_ARRAY, total: playerBScore.strokesReceived, withBorder: false, color: 'var(--color-warning)' },
            { label: `${playerBName} Net`, scores: playerBScore.matchNetHoleScores || playerBScore.holeScores, total: playerBScore.matchNetScore ?? playerBScore.netScore, withBorder: false, color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.1)' },
            { label: `${playerBName} Pts`, scores: playerBHolePoints, total: playerBTotal, withBorder: false, color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.15)' }
        ]
    }

    // Determine winner of match
    const getMatchResult = (match: Match): { winner: string, loser: string, isTie: boolean } | null => {
        const aPoints = match.playerAPoints
        const bPoints = match.playerBPoints
        
        if (aPoints === undefined || bPoints === undefined) return null
        
        const playerAName = getPlayerName(match.playerAId)
        const playerBName = getPlayerName(match.playerBId)
        
        if (aPoints > bPoints) {
            return { winner: playerAName, loser: playerBName, isTie: false }
        } else if (bPoints > aPoints) {
            return { winner: playerBName, loser: playerAName, isTie: false }
        } else {
            return { winner: playerAName, loser: playerBName, isTie: true }
        }
    }

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    const effectiveLeague = currentLeague
    const matchDayData = getMatchDaysByWeek()

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Home
                </Link>

                <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h1>Match Results</h1>
                    {effectiveLeague && <p className="text-gray-400 mt-1">{effectiveLeague.name}</p>}
                </div>

                {error && (
                    <div className="alert alert-error">
                        {error}
                    </div>
                )}

                {!effectiveLeague ? (
                    <div className="alert alert-info">
                        Select a league to view results.
                    </div>
                ) : (
                    <>
                        {/* Season Selector */}
                        {seasons.length > 0 && (
                            <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                                <label style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)', display: 'block' }}>
                                    Filter by Season
                                </label>
                                <select
                                    className="form-input"
                                    value={selectedSeasonId}
                                    onChange={(e) => setSelectedSeasonId(e.target.value)}
                                    style={{ maxWidth: '300px' }}
                                >
                                    <option value="">All Seasons</option>
                                    {seasons.map(season => (
                                        <option key={season.id} value={season.id}>
                                            {season.name} {season.active ? '(Current)' : ''}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        )}

                        <div className="card-glass">
                            <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                <Calendar className="w-5 h-5" /> Match Weeks
                            </h3>
                            
                            {matchDayData.length === 0 ? (
                                <p style={{ color: 'var(--color-text-muted)' }}>No completed match results yet.</p>
                            ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                    {matchDayData.map((mdData) => {
                                        const isMatchDayExpanded = expandedMatchDayId === mdData.matchDay.id
                                        const completedMatchesCount = mdData.matches.filter(m => m.match.status === 'completed').length
                                        
                                        return (
                                            <div 
                                                key={mdData.matchDay.id}
                                                style={{ 
                                                    border: '1px solid var(--color-border)',
                                                    borderRadius: 'var(--radius-md)',
                                                    overflow: 'hidden'
                                                }}
                                            >
                                                {/* Match Day Header */}
                                                <button
                                                    onClick={() => setExpandedMatchDayId(isMatchDayExpanded ? null : mdData.matchDay.id)}
                                                    style={{
                                                        width: '100%',
                                                        padding: 'var(--spacing-md)',
                                                        background: 'rgba(255, 255, 255, 0.02)',
                                                        border: 'none',
                                                        cursor: 'pointer',
                                                        display: 'flex',
                                                        justifyContent: 'space-between',
                                                        alignItems: 'center',
                                                        color: 'var(--color-text)'
                                                    }}
                                                >
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
                                                        <span className="badge badge-primary">
                                                            Week {mdData.matchDay.weekNumber || '?'}
                                                        </span>
                                                        <div style={{ textAlign: 'left' }}>
                                                            <p style={{ fontWeight: '600', marginBottom: '0.25rem' }}>
                                                                {formatDate(mdData.matchDay.date)}
                                                            </p>
                                                            <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                                {mdData.courseName} • {completedMatchesCount} match{completedMatchesCount !== 1 ? 'es' : ''}
                                                            </p>
                                                    </div>
                                                    </div>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
                                                        <span className={`badge ${mdData.matchDay.status === MATCH_DAY_STATUS.LOCKED ? 'badge-secondary' : 'badge-success'}`}>
                                                            {mdData.matchDay.status}
                                                        </span>
                                                        {isMatchDayExpanded ? <ChevronUp className="w-5 h-5" /> : <ChevronDown className="w-5 h-5" />}
                                                    </div>
                                                </button>
                                                
                                                {/* Match Day Content */}
                                                {isMatchDayExpanded && (
                                                    <div style={{ 
                                                        padding: 'var(--spacing-md)',
                                                        background: 'rgba(0, 0, 0, 0.2)',
                                                        borderTop: '1px solid var(--color-border)'
                                                    }}>
                                                        {mdData.matches.length === 0 ? (
                                                            <p style={{ color: 'var(--color-text-muted)' }}>No matches for this week.</p>
                                                        ) : (
                                                            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                                                {mdData.matches.map((matchData) => {
                                                                    const { match, playerAName, playerBName, playerAScore, playerBScore } = matchData
                                                                    const isMatchExpanded = expandedMatchId === match.id
                                                                    const matchResult = getMatchResult(match)
                                                                    const course = getCourse(match.courseId)
                                                                    const { playerATotal, playerBTotal } = calculateMatchPoints(playerAScore, playerBScore, match)

                                                                    const scorecardRows = buildScorecardRows(matchData, course)

                                                                    return (
                                                                        <ExpandableCard
                                                                            key={match.id}
                                                                            isExpanded={isMatchExpanded}
                                                                            onToggle={() => setExpandedMatchId(isMatchExpanded ? null : match.id)}
                                                                            header={
                                                                                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
                                                                                    {matchResult && (
                                                                                        <span className={`badge ${matchResult.isTie ? 'badge-secondary' : 'badge-success'}`}>
                                                                                            {matchResult.isTie ? 'TIE' : 'WIN'}
                                                                                        </span>
                                                                                    )}
                                                                                    <div style={{ textAlign: 'left' }}>
                                                                                        <p style={{ fontWeight: '600', marginBottom: '0.25rem' }}>
                                                                                            {playerAName} vs {playerBName}
                                                                                        </p>
                                                                                        {matchResult && !matchResult.isTie && (
                                                                                            <p style={{ fontSize: '0.75rem', color: 'var(--color-accent)' }}>
                                                                                                Winner: {matchResult.winner}
                                                                                            </p>
                                                                                        )}
                                                                                    </div>
                                                                                </div>
                                                                            }
                                                                            rightContent={
                                                                                playerAScore && playerBScore && (
                                                                                    <span style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>
                                                                                        {playerATotal} - {playerBTotal} pts
                                                                                    </span>
                                                                                )
                                                                            }
                                                                        >
                                                                            {playerAScore && playerBScore && scorecardRows.length > 0 ? (
                                                                                <>
                                                                                    <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text-secondary)' }}>
                                                                                        Match Scorecard
                                                                                    </h4>
                                                                                    <ScorecardTable rows={scorecardRows} />
                                                                                    <div style={{ marginTop: 'var(--spacing-md)', padding: 'var(--spacing-md)', background: 'rgba(255, 255, 255, 0.05)', borderRadius: 'var(--radius-md)' }}>
                                                                                        <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                                                            Points: 2 per hole (winner takes both, tie splits 1-1) + 4 for lowest total net (tie splits 2-2)
                                                                                        </p>
                                                                                    </div>
                                                                                </>
                                                                            ) : (
                                                                                <p style={{ color: 'var(--color-text-muted)' }}>
                                                                                    Scores not yet entered for this match.
                                                                                </p>
                                                                            )}
                                                                        </ExpandableCard>
                                                                    )
                                                                })}
                                                            </div>
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        )
                                    })}
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
