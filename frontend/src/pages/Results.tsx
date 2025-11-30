import { useState, useEffect, useCallback } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import { formatDateWithWeekday } from '../lib/utils'
import type { MatchDay, Match, Score, LeagueMemberWithPlayer, Course, Season } from '../types'
import { Calendar, ChevronDown, ChevronUp } from 'lucide-react'
import {
    ExpandableCard,
    ScorecardTable,
    AbsentBadge,
    EMPTY_STROKES_ARRAY
} from '../components/Scorecard'
import { LoadingSpinner } from '../components/Layout'

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
    playerAAbsent?: boolean
    playerBAbsent?: boolean
}

export default function Results() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()

    // Data State
    const [matchDays, setMatchDays] = useState<MatchDay[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [seasons, setSeasons] = useState<Season[]>([])

    // Cache State
    const [matchesCache, setMatchesCache] = useState<Record<string, Match[]>>({})
    const [scoresCache, setScoresCache] = useState<Record<string, Score[]>>({})

    // Loading State
    const [loading, setLoading] = useState(true)
    const [loadingMatchDayId, setLoadingMatchDayId] = useState<string | null>(null)
    const [loadingMatchId, setLoadingMatchId] = useState<string | null>(null)
    const [error, setError] = useState('')

    // UI State
    const [selectedSeasonId, setSelectedSeasonId] = useState<string>('')
    const [expandedMatchDayId, setExpandedMatchDayId] = useState<string | null>(null)
    const [expandedMatchId, setExpandedMatchId] = useState<string | null>(null)

    // Load initial data (MatchDays, Members, Courses, Seasons)
    useEffect(() => {
        async function loadData() {
            const effectiveLeagueId = leagueId || currentLeague?.id
            if (!effectiveLeagueId) return

            try {
                setLoading(true)
                const [matchDaysData, membersData, coursesData, seasonsData] = await Promise.all([
                    api.listMatchDays(effectiveLeagueId),
                    api.listLeagueMembers(effectiveLeagueId),
                    api.listCourses(effectiveLeagueId),
                    api.listSeasons(effectiveLeagueId)
                ])

                setMatchDays(matchDaysData)
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
    const getPlayerName = useCallback((playerId: string): string => {
        const member = members.find(m => m.playerId === playerId)
        return member?.player?.name || 'Unknown'
    }, [members])

    // Helper to get course by ID
    const getCourse = useCallback((courseId: string): Course | undefined => {
        return courses.find(c => c.id === courseId)
    }, [courses])

    const handleExpandMatchDay = async (matchDayId: string) => {
        if (expandedMatchDayId === matchDayId) {
            setExpandedMatchDayId(null)
            return
        }

        setExpandedMatchDayId(matchDayId)

        // If not in cache, fetch matches
        if (!matchesCache[matchDayId]) {
            const effectiveLeagueId = leagueId || currentLeague?.id
            if (!effectiveLeagueId) return

            setLoadingMatchDayId(matchDayId)
            try {
                const matches = await api.getMatchDayMatches(effectiveLeagueId, matchDayId)
                setMatchesCache(prev => ({ ...prev, [matchDayId]: matches }))
            } catch (err) {
                console.error(`Failed to load matches for match day ${matchDayId}:`, err)
                // Optionally show error toast
            } finally {
                setLoadingMatchDayId(null)
            }
        }
    }

    const handleExpandMatch = async (matchId: string) => {
        if (expandedMatchId === matchId) {
            setExpandedMatchId(null)
            return
        }

        setExpandedMatchId(matchId)

        // If not in cache, fetch scores
        if (!scoresCache[matchId]) {
            const effectiveLeagueId = leagueId || currentLeague?.id
            if (!effectiveLeagueId) return

            setLoadingMatchId(matchId)
            try {
                const scores = await api.getMatchScores(effectiveLeagueId, matchId)
                setScoresCache(prev => ({ ...prev, [matchId]: scores }))
            } catch (err) {
                console.error(`Failed to load scores for match ${matchId}:`, err)
            } finally {
                setLoadingMatchId(null)
            }
        }
    }

    // Calculate points for a match (reused logic)
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
            const playerANetScores = playerAScore.matchNetHoleScores
            const playerBNetScores = playerBScore.matchNetHoleScores

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
        }

        return { playerAHolePoints, playerBHolePoints, playerATotal, playerBTotal }
    }

    // Build scorecard rows (reused logic)
    const buildScorecardRows = (
        matchData: MatchWithScores,
        course: Course | undefined
    ) => {
        const { match, playerAName, playerBName, playerAScore, playerBScore } = matchData

        if (!playerAScore || !playerBScore) return []

        const { playerAHolePoints, playerBHolePoints, playerATotal, playerBTotal } =
            calculateMatchPoints(playerAScore, playerBScore, match)

        const getPlayerANetCellColors = (): ('win' | 'loss' | 'tie' | 'none')[] => {
            const playerANetScores = playerAScore.matchNetHoleScores || playerAScore.holeScores
            const playerBNetScores = playerBScore.matchNetHoleScores || playerBScore.holeScores
            return playerANetScores.map((aNet, i) => {
                const bNet = playerBNetScores[i]
                if (aNet < bNet) return 'win'
                if (aNet > bNet) return 'loss'
                return 'tie'
            })
        }

        const getPlayerBNetCellColors = (): ('win' | 'loss' | 'tie' | 'none')[] => {
            const playerANetScores = playerAScore.matchNetHoleScores || playerAScore.holeScores
            const playerBNetScores = playerBScore.matchNetHoleScores || playerBScore.holeScores
            return playerBNetScores.map((bNet, i) => {
                const aNet = playerANetScores[i]
                if (bNet < aNet) return 'win'
                if (bNet > aNet) return 'loss'
                return 'tie'
            })
        }

        const playerANetCellColors = getPlayerANetCellColors()
        const playerBNetCellColors = getPlayerBNetCellColors()

        return [
            ...(course?.holePars ? [{ label: 'Par', scores: course.holePars, total: course.par, withBorder: false }] : []),
            ...(course?.holeHandicaps ? [{ label: 'Hole Hdcp', scores: course.holeHandicaps, total: '', withBorder: true }] : []),
            {
                label: playerAScore.playerAbsent ? `${playerAName} Gross (Absent)` : `${playerAName} Gross`,
                scores: playerAScore.holeScores,
                total: playerAScore.grossScore,
                withBorder: false,
                showGolfSymbols: !playerAScore.playerAbsent && !!course?.holePars,
                pars: course?.holePars
            },
            { label: `${playerAName} Strokes`, scores: playerAScore.matchStrokes || EMPTY_STROKES_ARRAY, total: playerAScore.strokesReceived, withBorder: false, color: 'var(--color-accent)' },
            {
                label: `${playerAName} Net`,
                scores: playerAScore.matchNetHoleScores || playerAScore.holeScores,
                total: playerAScore.matchNetScore ?? playerAScore.netScore,
                withBorder: false,
                cellColors: playerANetCellColors
            },
            { label: `${playerAName} Pts`, scores: playerAHolePoints, total: playerATotal, withBorder: true, color: 'var(--color-primary)', bgColor: 'rgba(16, 185, 129, 0.15)' },
            {
                label: playerBScore.playerAbsent ? `${playerBName} Gross (Absent)` : `${playerBName} Gross`,
                scores: playerBScore.holeScores,
                total: playerBScore.grossScore,
                withBorder: false,
                showGolfSymbols: !playerBScore.playerAbsent && !!course?.holePars,
                pars: course?.holePars
            },
            { label: `${playerBName} Strokes`, scores: playerBScore.matchStrokes || EMPTY_STROKES_ARRAY, total: playerBScore.strokesReceived, withBorder: false, color: 'var(--color-warning)' },
            {
                label: `${playerBName} Net`,
                scores: playerBScore.matchNetHoleScores || playerBScore.holeScores,
                total: playerBScore.matchNetScore ?? playerBScore.netScore,
                withBorder: false,
                cellColors: playerBNetCellColors
            },
            { label: `${playerBName} Pts`, scores: playerBHolePoints, total: playerBTotal, withBorder: false, color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.15)' }
        ]
    }

    // Determine winner of match (reused logic)
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
        return <LoadingSpinner />
    }

    const effectiveLeague = currentLeague

    // Filter match days
    const filteredMatchDays = selectedSeasonId
        ? matchDays.filter(md => md.seasonId === selectedSeasonId)
        : matchDays

    const sortedMatchDays = filteredMatchDays
        .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())

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

                            {sortedMatchDays.length === 0 ? (
                                <p style={{ color: 'var(--color-text-muted)' }}>No completed match results yet.</p>
                            ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                    {sortedMatchDays.map((matchDay) => {
                                        const isMatchDayExpanded = expandedMatchDayId === matchDay.id
                                        const course = getCourse(matchDay.courseId)
                                        const matches = matchesCache[matchDay.id] || []
                                        const isLoadingMatches = loadingMatchDayId === matchDay.id

                                        return (
                                            <div
                                                key={matchDay.id}
                                                style={{
                                                    border: '1px solid var(--color-border)',
                                                    borderRadius: 'var(--radius-md)',
                                                    overflow: 'hidden'
                                                }}
                                            >
                                                {/* Match Day Header */}
                                                <button
                                                    onClick={() => handleExpandMatchDay(matchDay.id)}
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
                                                        <div style={{ textAlign: 'left' }}>
                                                            <p style={{ fontWeight: '600', marginBottom: '0.25rem' }}>
                                                                {formatDateWithWeekday(matchDay.date)}
                                                            </p>
                                                            <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                                {course?.name || 'Unknown Course'}
                                                            </p>
                                                        </div>
                                                    </div>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
                                                        <span className={`badge ${matchDay.status === MATCH_DAY_STATUS.LOCKED ? 'badge-secondary' : 'badge-success'}`}>
                                                            {matchDay.status}
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
                                                        {isLoadingMatches ? (
                                                            <div style={{ display: 'flex', justifyContent: 'center', padding: '1rem' }}>
                                                                <LoadingSpinner />
                                                            </div>
                                                        ) : matches.length === 0 ? (
                                                            <p style={{ color: 'var(--color-text-muted)' }}>No matches found for this week.</p>
                                                        ) : (
                                                            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                                                {matches.map((match) => {
                                                                    const isMatchExpanded = expandedMatchId === match.id
                                                                    const matchResult = getMatchResult(match)
                                                                    const playerAName = getPlayerName(match.playerAId)
                                                                    const playerBName = getPlayerName(match.playerBId)

                                                                    const matchScores = scoresCache[match.id] || []
                                                                    const playerAScore = matchScores.find(s => s.playerId === match.playerAId)
                                                                    const playerBScore = matchScores.find(s => s.playerId === match.playerBId)

                                                                    const matchData: MatchWithScores = {
                                                                        match,
                                                                        playerAName,
                                                                        playerBName,
                                                                        playerAScore,
                                                                        playerBScore
                                                                    }

                                                                    const scorecardRows = buildScorecardRows(matchData, course)
                                                                    const isLoadingScores = loadingMatchId === match.id

                                                                    // Use absence from match object if available, otherwise from score
                                                                    const playerAAbsent = match.playerAAbsent || playerAScore?.playerAbsent
                                                                    const playerBAbsent = match.playerBAbsent || playerBScore?.playerAbsent

                                                                    // Calculate totals for display on card
                                                                    const { playerATotal, playerBTotal } = calculateMatchPoints(playerAScore, playerBScore, match)

                                                                    return (
                                                                        <ExpandableCard
                                                                            key={match.id}
                                                                            isExpanded={isMatchExpanded}
                                                                            onToggle={() => handleExpandMatch(match.id)}
                                                                            header={
                                                                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                                                                    {/* Player A Row */}
                                                                                    <div style={{ display: 'grid', gridTemplateColumns: '40px 1fr auto', alignItems: 'center', gap: '0.5rem' }}>
                                                                                        <div style={{ display: 'flex', justifyContent: 'center' }}>
                                                                                            {matchResult && !matchResult.isTie && matchResult.winner === playerAName && (
                                                                                                <span className="badge badge-success" style={{ fontSize: '0.7rem', padding: '0.1rem 0.3rem' }}>WIN</span>
                                                                                            )}
                                                                                        </div>
                                                                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                                                            <span style={{ fontWeight: matchResult?.winner === playerAName ? 'bold' : 'normal' }}>{playerAName}</span>
                                                                                            {playerAAbsent && <AbsentBadge small />}
                                                                                        </div>
                                                                                        <span style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>{playerATotal}</span>
                                                                                    </div>

                                                                                    {/* Player B Row */}
                                                                                    <div style={{ display: 'grid', gridTemplateColumns: '40px 1fr auto', alignItems: 'center', gap: '0.5rem' }}>
                                                                                        <div style={{ display: 'flex', justifyContent: 'center' }}>
                                                                                            {matchResult && !matchResult.isTie && matchResult.winner === playerBName && (
                                                                                                <span className="badge badge-success" style={{ fontSize: '0.7rem', padding: '0.1rem 0.3rem' }}>WIN</span>
                                                                                            )}
                                                                                        </div>
                                                                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                                                            <span style={{ fontWeight: matchResult?.winner === playerBName ? 'bold' : 'normal' }}>{playerBName}</span>
                                                                                            {playerBAbsent && <AbsentBadge small />}
                                                                                        </div>
                                                                                        <span style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>{playerBTotal}</span>
                                                                                    </div>

                                                                                    {/* Winner Footer */}
                                                                                    {matchResult && !matchResult.isTie && (
                                                                                        <div style={{ fontSize: '0.75rem', color: 'var(--color-accent)', marginTop: '0.25rem', paddingLeft: '48px' }}>
                                                                                            Winner: {matchResult.winner}
                                                                                        </div>
                                                                                    )}
                                                                                </div>
                                                                            }
                                                                            rightContent={null}
                                                                        >
                                                                            {
                                                                                isLoadingScores ? (
                                                                                    <div style={{ display: 'flex', justifyContent: 'center', padding: '1rem' }} >
                                                                                        <LoadingSpinner />
                                                                                    </div>
                                                                                ) : playerAScore && playerBScore && scorecardRows.length > 0 ? (
                                                                                    <>
                                                                                        <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text-secondary)' }}>
                                                                                            Match Scorecard
                                                                                        </h4>
                                                                                        <ScorecardTable rows={scorecardRows} />
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
                                                        )
                                                        }
                                                    </div>
                                                )}
                                            </div>
                                        )
                                    })}
                                </div>
                            )}
                        </div>
                    </>
                )
                }
            </div >
        </div >
    )
}
