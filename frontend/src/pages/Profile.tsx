import { useState, useEffect, useCallback } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type { Player, Score, Season, Match, LeagueMemberWithPlayer, Course, League } from '../types'
import { Trophy, Calendar, Target, TrendingUp, Users } from 'lucide-react'
import { 
    ExpandableCard, 
    ScorecardTable, 
    StatItem,
    AbsentBadge,
    EMPTY_STROKES_ARRAY 
} from '../components/Scorecard'

// Constants
const MAX_HANDICAP_ROUNDS = 20

// Helper function to get player/opponent points based on position
function getMatchPoints(match: Match, isPlayerA: boolean): { playerPoints?: number, opponentPoints?: number } {
    return {
        playerPoints: isPlayerA ? match.playerAPoints : match.playerBPoints,
        opponentPoints: isPlayerA ? match.playerBPoints : match.playerAPoints
    }
}

// Extended types for profile data
interface HandicapHistoryEntry {
    date: string
    handicapIndex: number
    roundId?: string
    isProvisional: boolean
    seasonId?: string
    seasonName?: string
}

interface MatchupDetail {
    match: Match
    opponentName: string
    playerScore?: Score
    opponentScore?: Score
    result: 'won' | 'lost' | 'tied' | 'pending'
    courseName: string
    date: string
    isPlayerA: boolean
}

export default function Profile() {
    const { playerId } = useParams<{ playerId: string }>()
    const navigate = useNavigate()
    const { currentLeague, userLeagues, isLoading: leagueLoading } = useLeague()
    
    const [currentUser, setCurrentUser] = useState<Player | null>(null)
    const [player, setPlayer] = useState<Player | null>(null)
    const [scores, setScores] = useState<Score[]>([])
    const [allMatchScores, setAllMatchScores] = useState<Score[]>([]) // All scores for matches the player is in
    const [seasons, setSeasons] = useState<Season[]>([])
    const [matches, setMatches] = useState<Match[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [provisionalHandicap, setProvisionalHandicap] = useState<number | null>(null)
    const [currentHandicapIndex, setCurrentHandicapIndex] = useState<number | null>(null) // From HandicapRecord API
    const [allLeagues, setAllLeagues] = useState<League[]>([])
    const [leagueHandicaps, setLeagueHandicaps] = useState<Map<string, number | null>>(new Map()) // Handicaps per league
    
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [accessDenied, setAccessDenied] = useState(false)
    
    // UI state
    const [selectedSeasonId, setSelectedSeasonId] = useState<string>('')
    const [expandedMatchId, setExpandedMatchId] = useState<string | null>(null)
    const [expandedRoundId, setExpandedRoundId] = useState<string | null>(null)
    const [activeTab, setActiveTab] = useState<'overview' | 'rounds' | 'handicaps' | 'matchups'>('overview')

    // Fetch all leagues to display names and handicaps for each league (only if user has league memberships)
    useEffect(() => {
        async function fetchLeaguesAndHandicaps() {
            if (userLeagues.length === 0 || !player) return
            try {
                const leagues = await api.listLeagues()
                setAllLeagues(leagues)
                
                // Fetch handicaps for each league the player is in
                const handicapsMap = new Map<string, number | null>()
                
                await Promise.all(userLeagues.map(async (membership) => {
                    try {
                        // Get the active season for this league
                        const leagueSeasons = await api.listSeasons(membership.leagueId)
                        const activeSeason = leagueSeasons.find(s => s.active)
                        
                        if (activeSeason) {
                            const handicapRecord = await api.getPlayerHandicap(
                                membership.leagueId, 
                                activeSeason.id, 
                                player.id
                            )
                            handicapsMap.set(membership.leagueId, handicapRecord.leagueHandicapIndex)
                        } else {
                            // No active season, use provisional handicap from membership
                            handicapsMap.set(membership.leagueId, membership.provisionalHandicap)
                        }
                    } catch {
                        // If handicap fetch fails, show provisional
                        handicapsMap.set(membership.leagueId, membership.provisionalHandicap)
                    }
                }))
                
                setLeagueHandicaps(handicapsMap)
            } catch (err) {
                console.error('Failed to load leagues:', err)
            }
        }
        fetchLeaguesAndHandicaps()
    }, [userLeagues, player])

    // Load current user and verify access
    useEffect(() => {
        async function loadUserAndVerifyAccess() {
            try {
                const userInfo = await api.getCurrentUser()
                
                if (!userInfo.linked || !userInfo.player) {
                    // User not linked to a player - redirect to link account
                    navigate('/link-account')
                    return
                }
                
                setCurrentUser(userInfo.player)
                
                // Verify playerId matches current user (only allow viewing own profile)
                if (playerId && playerId !== userInfo.player.id) {
                    setAccessDenied(true)
                    setLoading(false)
                    return
                }
                
                // If no playerId provided or matches current user, set player
                setPlayer(userInfo.player)
            } catch (err) {
                console.error('Failed to load user:', err)
                setError('Failed to load user information')
                setLoading(false)
            }
        }
        
        if (!leagueLoading) {
            loadUserAndVerifyAccess()
        }
    }, [playerId, leagueLoading, navigate])

    // Load league-specific data when player and league are available
    const loadLeagueData = useCallback(async () => {
        if (!player || !currentLeague) {
            setLoading(false)
            return
        }

        try {
            // First fetch seasons and other non-season-dependent data
            const [scoresData, seasonsData, matchesData, membersData, coursesData] = await Promise.all([
                api.getPlayerScores(currentLeague.id, player.id).catch(() => []),
                api.listSeasons(currentLeague.id).catch(() => []),
                api.listMatches(currentLeague.id).catch(() => []),
                api.listLeagueMembers(currentLeague.id).catch(() => []),
                api.listCourses(currentLeague.id).catch(() => []),
            ])

            setScores(scoresData)
            setSeasons(seasonsData)
            setMatches(matchesData)
            setMembers(membersData)
            setCourses(coursesData)

            // Find the active season
            const activeSeason = seasonsData.find(s => s.active)
            if (activeSeason) {
                setSelectedSeasonId(activeSeason.id)
                
                // Now fetch handicap using the active season ID
                const handicapRecord = await api.getPlayerHandicap(currentLeague.id, activeSeason.id, player.id).catch(() => null)
                if (handicapRecord) {
                    setCurrentHandicapIndex(handicapRecord.leagueHandicapIndex)
                }
            } else if (seasonsData.length > 0) {
                setSelectedSeasonId(seasonsData[0].id)
            }

            // Fetch all scores for matches the player is in (including opponent scores)
            const playerMatches = matchesData.filter(
                (m: Match) => m.playerAId === player.id || m.playerBId === player.id
            )
            const matchScorePromises = playerMatches.map((match: Match) =>
                api.getMatchScores(currentLeague.id, match.id).catch(() => [])
            )
            const matchScoresArrays = await Promise.all(matchScorePromises)
            const allScores = matchScoresArrays.flat()
            setAllMatchScores(allScores)

            // Get provisional handicap from league membership
            const playerMembership = membersData.find(m => m.playerId === player.id)
            if (playerMembership) {
                setProvisionalHandicap(playerMembership.provisionalHandicap)
            }
        } catch (err) {
            console.error('Failed to load league data:', err)
            setError('Failed to load profile data')
        } finally {
            setLoading(false)
        }
    }, [player, currentLeague])

    useEffect(() => {
        if (player && currentLeague) {
            loadLeagueData()
        } else if (player && !leagueLoading) {
            setLoading(false)
        }
    }, [player, currentLeague, leagueLoading, loadLeagueData])

    // Get current handicap from HandicapRecord API (preferred) or fallback to provisional
    const getCurrentHandicap = (): number | null => {
        // Use the current handicap index from HandicapRecord if available
        if (currentHandicapIndex !== null) {
            return currentHandicapIndex
        }
        // Fallback to provisional handicap if no HandicapRecord exists
        return provisionalHandicap
    }

    // Build handicap history including provisional
    // Only shows entries when handicap actually changes, excluding absent rounds
    const getHandicapHistory = (): HandicapHistoryEntry[] => {
        const history: HandicapHistoryEntry[] = []
        
        // Get filtered scores for season
        const filteredScores = selectedSeasonId 
            ? scores.filter(s => {
                const match = matches.find(m => m.id === s.matchId)
                return match?.seasonId === selectedSeasonId
            })
            : scores

        // Filter to non-absent rounds with valid data, sorted by date
        const playedRounds = filteredScores
            .filter(s => s.date && s.handicapIndex !== undefined && !s.playerAbsent)
            .sort((a, b) => new Date(a.date!).getTime() - new Date(b.date!).getTime())

        // Add provisional handicap at start if available
        if (provisionalHandicap !== null) {
            const activeSeason = seasons.find(s => s.active) || seasons[0]
            history.push({
                date: activeSeason?.startDate || new Date().toISOString(),
                handicapIndex: provisionalHandicap,
                isProvisional: true,
                seasonId: activeSeason?.id,
                seasonName: activeSeason?.name || 'Season Start'
            })
        }

        // Track the last handicap to only show changes
        let lastHandicap = provisionalHandicap

        // Add handicap from each played round (non-absent)
        playedRounds.forEach(score => {
            const match = matches.find(m => m.id === score.matchId)
            const season = match ? seasons.find(s => s.id === match.seasonId) : null
            
            // Safe access with fallback - we know handicapIndex exists from filter
            const currentHandicapIndex = score.handicapIndex ?? 0
            const currentDate = score.date ?? new Date().toISOString()
            
            // Only add if handicap changed from previous entry (or first played round)
            const handicapChanged = lastHandicap === null || 
                Math.abs(currentHandicapIndex - lastHandicap) >= 0.05
            
            if (handicapChanged) {
                history.push({
                    date: currentDate,
                    handicapIndex: currentHandicapIndex,
                    roundId: score.id,
                    isProvisional: false,
                    seasonId: season?.id,
                    seasonName: season?.name
                })
                lastHandicap = currentHandicapIndex
            }
        })

        return history
    }

    // Get player's matchups
    const getMatchups = (): MatchupDetail[] => {
        const playerMatches = matches.filter(
            m => m.playerAId === player?.id || m.playerBId === player?.id
        )

        const filteredMatches = selectedSeasonId
            ? playerMatches.filter(m => m.seasonId === selectedSeasonId)
            : playerMatches

        return filteredMatches.map(match => {
            const isPlayerA = match.playerAId === player?.id
            const opponentId = isPlayerA ? match.playerBId : match.playerAId
            const opponent = members.find(m => m.playerId === opponentId)
            const course = courses.find(c => c.id === match.courseId)
            
            // Use allMatchScores which includes both player and opponent scores
            const playerScore = allMatchScores.find(s => s.matchId === match.id && s.playerId === player?.id)
            const opponentScore = allMatchScores.find(s => s.matchId === match.id && s.playerId === opponentId)

            let result: 'won' | 'lost' | 'tied' | 'pending' = 'pending'
            if (match.status === 'completed') {
                // Use stored match points when available
                const { playerPoints, opponentPoints } = getMatchPoints(match, isPlayerA)
                
                if (playerPoints !== undefined && opponentPoints !== undefined) {
                    // Determine result based on stored match points
                    if (playerPoints > opponentPoints) {
                        result = 'won'
                    } else if (playerPoints < opponentPoints) {
                        result = 'lost'
                    } else {
                        result = 'tied'
                    }
                } else if (playerScore && opponentScore) {
                    // Fallback to calculating from net scores for backwards compatibility
                    const playerNet = playerScore.matchNetScore ?? playerScore.netScore
                    const opponentNet = opponentScore.matchNetScore ?? opponentScore.netScore
                    if (playerNet < opponentNet) {
                        result = 'won'
                    } else if (playerNet > opponentNet) {
                        result = 'lost'
                    } else {
                        result = 'tied'
                    }
                }
            }

            return {
                match,
                opponentName: opponent?.player?.name || 'Unknown',
                playerScore,
                opponentScore,
                result,
                courseName: course?.name || 'Unknown',
                date: match.matchDate,
                isPlayerA
            }
        }).sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
    }

    // Format date for display
    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'short', 
            day: 'numeric' 
        })
    }

    // Get match result stats
    const getMatchStats = () => {
        const matchups = getMatchups().filter(m => m.result !== 'pending')
        const won = matchups.filter(m => m.result === 'won').length
        const lost = matchups.filter(m => m.result === 'lost').length
        const tied = matchups.filter(m => m.result === 'tied').length
        return { won, lost, tied, total: matchups.length }
    }

    // Loading state
    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    // Access denied - trying to view another player's profile
    if (accessDenied) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                    <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                        ← Back to Home
                    </Link>
                    <div className="alert alert-error">
                        <strong>Access Denied:</strong> You can only view your own profile.
                    </div>
                    {currentUser && (
                        <Link 
                            to={`/profile/${currentUser.id}`} 
                            className="btn btn-primary" 
                            style={{ marginTop: 'var(--spacing-lg)' }}
                        >
                            View My Profile
                        </Link>
                    )}
                </div>
            </div>
        )
    }

    // Error state
    if (error) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                    <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                        ← Back to Home
                    </Link>
                    <div className="alert alert-error">{error}</div>
                </div>
            </div>
        )
    }

    // Not linked state
    if (!player) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                    <div className="alert alert-warning">
                        <strong>Not Linked:</strong> Please link your account to a player profile first.
                    </div>
                    <Link to="/link-account" className="btn btn-primary" style={{ marginTop: 'var(--spacing-lg)' }}>
                        Link Account
                    </Link>
                </div>
            </div>
        )
    }

    const currentHandicap = getCurrentHandicap()
    const matchStats = getMatchStats()
    const handicapHistory = getHandicapHistory()
    const matchups = getMatchups()
    const filteredScores = selectedSeasonId
        ? scores.filter(s => {
            const match = matches.find(m => m.id === s.matchId)
            return match?.seasonId === selectedSeasonId
        })
        : scores

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Home
                </Link>

                {/* Header */}
                <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h1>My Profile</h1>
                    {currentLeague && <p className="text-gray-400 mt-1">{currentLeague.name}</p>}
                </div>

                {/* Player Info Card */}
                <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 'var(--spacing-lg)' }}>
                        <div>
                            <h2 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text)' }}>{player.name}</h2>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Email</p>
                                    <p style={{ color: 'var(--color-text)' }}>{player.email}</p>
                                </div>
                                <div>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Status</p>
                                    <span className={`badge ${player.active ? 'badge-success' : 'badge-danger'}`}>
                                        {player.active ? 'Active' : 'Inactive'}
                                    </span>
                                </div>
                            </div>
                        </div>
                        
                        {/* Quick Stats */}
                        {currentLeague && currentHandicap !== null && (
                            <div style={{ 
                                display: 'flex', 
                                gap: 'var(--spacing-xl)',
                                padding: 'var(--spacing-md)',
                                background: 'rgba(16, 185, 129, 0.1)',
                                borderRadius: 'var(--radius-lg)'
                            }}>
                                <div style={{ textAlign: 'center' }}>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem', marginBottom: '0.25rem' }}>Handicap</p>
                                    <p style={{ fontSize: '1.75rem', fontWeight: 'bold', color: 'var(--color-primary)' }}>
                                        {currentHandicap.toFixed(1)}
                                    </p>
                                </div>
                                <div style={{ textAlign: 'center' }}>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem', marginBottom: '0.25rem' }}>Rounds</p>
                                    <p style={{ fontSize: '1.75rem', fontWeight: 'bold', color: 'var(--color-text)' }}>
                                        {filteredScores.length}
                                    </p>
                                </div>
                                <div style={{ textAlign: 'center' }}>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem', marginBottom: '0.25rem' }}>W-L-T</p>
                                    <p style={{ fontSize: '1.75rem', fontWeight: 'bold', color: 'var(--color-text)' }}>
                                        {matchStats.won}-{matchStats.lost}-{matchStats.tied}
                                    </p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Leagues Enrolled */}
                {userLeagues.length > 0 && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Users className="w-5 h-5" /> Leagues Enrolled
                        </h3>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--spacing-md)' }}>
                            {userLeagues.map(membership => {
                                const league = allLeagues.find(l => l.id === membership.leagueId)
                                const handicap = leagueHandicaps.get(membership.leagueId)
                                return (
                                    <div 
                                        key={membership.leagueId}
                                        style={{ 
                                            padding: 'var(--spacing-sm) var(--spacing-md)',
                                            background: membership.leagueId === currentLeague?.id 
                                                ? 'rgba(16, 185, 129, 0.2)' 
                                                : 'rgba(255, 255, 255, 0.05)',
                                            borderRadius: 'var(--radius-md)',
                                            border: membership.leagueId === currentLeague?.id 
                                                ? '1px solid var(--color-accent)' 
                                                : '1px solid var(--color-border)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '0.5rem'
                                        }}
                                    >
                                        <Trophy className="w-4 h-4" style={{ color: 'var(--color-primary)' }} />
                                        <span style={{ color: 'var(--color-text)' }}>
                                            {league?.name || 'Unknown League'}
                                        </span>
                                        <span className={`badge ${membership.role === 'admin' ? 'badge-primary' : 'badge-secondary'}`} style={{ fontSize: '0.65rem' }}>
                                            {membership.role}
                                        </span>
                                        {handicap !== undefined && handicap !== null && (
                                            <span 
                                                style={{ 
                                                    marginLeft: 'auto',
                                                    color: 'var(--color-primary)', 
                                                    fontWeight: 'bold',
                                                    fontSize: '0.875rem'
                                                }}
                                                title="Current handicap index for this league"
                                            >
                                                HC: {handicap.toFixed(1)}
                                            </span>
                                        )}
                                    </div>
                                )
                            })}
                        </div>
                        {userLeagues.length > 1 && (
                            <p style={{ 
                                marginTop: 'var(--spacing-md)', 
                                color: 'var(--color-text-muted)', 
                                fontSize: '0.75rem',
                                fontStyle: 'italic'
                            }}>
                                Each league tracks handicaps independently based on rounds played in that league.
                            </p>
                        )}
                    </div>
                )}

                {!currentLeague ? (
                    <div className="alert alert-info">
                        <strong>No League Selected:</strong> Select a league from the home page to view your handicap history, rounds, and match results.
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

                        {/* Tabs */}
                        <div style={{ marginBottom: 'var(--spacing-lg)', borderBottom: '1px solid var(--color-border)', overflowX: 'auto', WebkitOverflowScrolling: 'touch' }}>
                            <div style={{ display: 'flex', gap: 'var(--spacing-md)', minWidth: 'max-content' }}>
                                {[
                                    { id: 'overview', label: 'Overview', icon: Target },
                                    { id: 'rounds', label: 'Round History', icon: Calendar },
                                    { id: 'handicaps', label: 'Handicap History', icon: TrendingUp },
                                    { id: 'matchups', label: 'Matchups', icon: Trophy },
                                ].map(tab => {
                                    const Icon = tab.icon
                                    return (
                                        <button
                                            key={tab.id}
                                            onClick={() => setActiveTab(tab.id as typeof activeTab)}
                                            style={{
                                                padding: 'var(--spacing-md) var(--spacing-lg)',
                                                background: 'transparent',
                                                border: 'none',
                                                borderBottom: activeTab === tab.id ? '2px solid var(--color-primary)' : '2px solid transparent',
                                                color: activeTab === tab.id ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: '0.5rem',
                                                fontWeight: activeTab === tab.id ? '600' : '400',
                                                transition: 'all 0.2s ease',
                                                whiteSpace: 'nowrap',
                                                flexShrink: 0
                                            }}
                                        >
                                            <Icon className="w-4 h-4" />
                                            {tab.label}
                                        </button>
                                    )
                                })}
                            </div>
                        </div>

                        {/* Overview Tab */}
                        {activeTab === 'overview' && (
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-xl)' }}>
                                {/* Current Handicap Card */}
                                <div className="card-glass">
                                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Current Handicap</h3>
                                    {currentHandicap !== null ? (
                                        <div>
                                            <p style={{ fontSize: '3rem', fontWeight: 'bold', color: 'var(--color-primary)', marginBottom: 'var(--spacing-sm)' }}>
                                                {currentHandicap.toFixed(1)}
                                            </p>
                                            {provisionalHandicap !== null && scores.length === 0 && (
                                                <span className="badge badge-warning">Provisional</span>
                                            )}
                                            {scores.length > 0 && (
                                                <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                                    Based on {Math.min(scores.length, MAX_HANDICAP_ROUNDS)} round{scores.length !== 1 ? 's' : ''}
                                                </p>
                                            )}
                                        </div>
                                    ) : (
                                        <p style={{ color: 'var(--color-text-muted)' }}>No handicap established yet</p>
                                    )}
                                </div>

                                {/* Match Record Card */}
                                <div className="card-glass">
                                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Match Record</h3>
                                    {matchStats.total > 0 ? (
                                        <div>
                                            <div style={{ display: 'flex', gap: 'var(--spacing-xl)', marginBottom: 'var(--spacing-md)' }}>
                                                <div>
                                                    <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-accent)' }}>{matchStats.won}</p>
                                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>Wins</p>
                                                </div>
                                                <div>
                                                    <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-danger)' }}>{matchStats.lost}</p>
                                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>Losses</p>
                                                </div>
                                                <div>
                                                    <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-text-secondary)' }}>{matchStats.tied}</p>
                                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>Ties</p>
                                                </div>
                                            </div>
                                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                                Win Rate: {matchStats.total > 0 ? ((matchStats.won / matchStats.total) * 100).toFixed(0) : 0}%
                                            </p>
                                        </div>
                                    ) : (
                                        <p style={{ color: 'var(--color-text-muted)' }}>No completed matches yet</p>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* Rounds Tab */}
                        {activeTab === 'rounds' && (
                            <div className="card-glass">
                                <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Round History</h3>
                                {filteredScores.length === 0 ? (
                                    <p style={{ color: 'var(--color-text-muted)' }}>No rounds recorded yet.</p>
                                ) : (
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                        {[...filteredScores]
                                            .filter(s => s.date)
                                            .sort((a, b) => new Date(b.date!).getTime() - new Date(a.date!).getTime())
                                            .map((score) => {
                                                const course = courses.find(c => c.id === score.courseId)
                                                const isExpanded = expandedRoundId === score.id
                                                const isAbsent = score.playerAbsent
                                                const scorecardRows = [
                                                    ...(course?.holePars ? [{ label: 'Par', scores: course.holePars, total: course.par, withBorder: false }] : []),
                                                    ...(course?.holeHandicaps ? [{ label: 'Hole Hdcp', scores: course.holeHandicaps, total: '', withBorder: true }] : []),
                                                    { 
                                                        label: isAbsent ? 'Gross (Auto)' : 'Gross', 
                                                        scores: score.holeScores, 
                                                        total: score.grossScore, 
                                                        withBorder: false,
                                                        // Add golf scoring symbols for non-absent rounds
                                                        showGolfSymbols: !isAbsent && !!course?.holePars,
                                                        pars: course?.holePars
                                                    }
                                                ]
                                                return (
                                                    <ExpandableCard
                                                        key={score.id}
                                                        isExpanded={isExpanded}
                                                        onToggle={() => setExpandedRoundId(isExpanded ? null : score.id)}
                                                        header={
                                                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-lg)' }}>
                                                                {isAbsent && <AbsentBadge />}
                                                                <div style={{ textAlign: 'left' }}>
                                                                    <p style={{ fontWeight: '600', marginBottom: '0.25rem' }}>
                                                                        {formatDate(score.date!)}
                                                                    </p>
                                                                    <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                                        {course?.name || 'Unknown'}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        }
                                                        rightContent={
                                                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-xl)' }}>
                                                                <div style={{ textAlign: 'center' }}>
                                                                    <p style={{ fontSize: '0.65rem', color: 'var(--color-text-muted)' }}>Gross</p>
                                                                    <p style={{ fontWeight: 'bold' }}>{score.grossScore}</p>
                                                                </div>
                                                                <div style={{ textAlign: 'center' }}>
                                                                    <p style={{ fontSize: '0.65rem', color: 'var(--color-text-muted)' }}>Differential</p>
                                                                    <p style={{ fontWeight: 'bold', color: isAbsent ? 'var(--color-warning)' : 'var(--color-primary)' }}>
                                                                        {isAbsent ? 'N/A' : (score.handicapDifferential?.toFixed(1) || 'N/A')}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        }
                                                    >
                                                        {isAbsent && (
                                                            <div style={{ 
                                                                marginBottom: 'var(--spacing-md)', 
                                                                padding: 'var(--spacing-sm) var(--spacing-md)',
                                                                backgroundColor: 'rgba(234, 179, 8, 0.1)',
                                                                borderRadius: 'var(--radius-md)',
                                                                borderLeft: '3px solid var(--color-warning)'
                                                            }}>
                                                                <p style={{ fontSize: '0.8rem', color: 'var(--color-warning)' }}>
                                                                    Player was absent. Scores were auto-calculated and do not affect handicap.
                                                                </p>
                                                            </div>
                                                        )}
                                                        <div style={{ marginBottom: 'var(--spacing-md)', display: 'flex', gap: 'var(--spacing-xl)', flexWrap: 'wrap' }}>
                                                            <StatItem label="Handicap Index" value={score.handicapIndex?.toFixed(1) || 'N/A'} />
                                                        </div>
                                                        <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text-secondary)' }}>
                                                            Hole Scores
                                                        </h4>
                                                        <ScorecardTable rows={scorecardRows} />
                                                    </ExpandableCard>
                                                )
                                            })}
                                    </div>
                                )}
                            </div>
                        )}

                        {/* Handicap History Tab */}
                        {activeTab === 'handicaps' && (
                            <div className="card-glass">
                                <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Handicap History</h3>
                                {handicapHistory.length === 0 ? (
                                    <p style={{ color: 'var(--color-text-muted)' }}>No handicap history available.</p>
                                ) : (
                                    <div className="table-container">
                                        <table className="table">
                                            <thead>
                                                <tr>
                                                    <th>Date</th>
                                                    <th>Season</th>
                                                    <th>Handicap Index</th>
                                                    <th>Type</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {handicapHistory.map((entry, index) => (
                                                    <tr key={index}>
                                                        <td>{formatDate(entry.date)}</td>
                                                        <td>{entry.seasonName || 'N/A'}</td>
                                                        <td style={{ 
                                                            fontWeight: 'bold', 
                                                            color: entry.isProvisional ? 'var(--color-warning)' : 'var(--color-primary)' 
                                                        }}>
                                                            {entry.handicapIndex.toFixed(1)}
                                                        </td>
                                                        <td>
                                                            <span className={`badge ${entry.isProvisional ? 'badge-warning' : 'badge-primary'}`}>
                                                                {entry.isProvisional ? 'Provisional' : 'Calculated'}
                                                            </span>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                )}
                            </div>
                        )}

                        {/* Matchups Tab */}
                        {activeTab === 'matchups' && (
                            <div className="card-glass">
                                <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Match History</h3>
                                {matchups.length === 0 ? (
                                    <p style={{ color: 'var(--color-text-muted)' }}>No matches scheduled yet.</p>
                                ) : (
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                                        {matchups.map((matchup) => {
                                            const isExpanded = expandedMatchId === matchup.match.id
                                            const course = courses.find(c => c.id === matchup.match.courseId)
                                            
                                            // Calculate points per hole based on net scores
                                            const calculateHolePoints = (): { playerPoints: number[], opponentPoints: number[], playerTotal: number, opponentTotal: number } => {
                                                const playerPoints: number[] = []
                                                const opponentPoints: number[] = []
                                                let playerTotal = 0
                                                let opponentTotal = 0
                                                
                                                // Use server-stored match points when available
                                                const { playerPoints: storedPlayerTotal, opponentPoints: storedOpponentTotal } = getMatchPoints(matchup.match, matchup.isPlayerA)
                                                
                                                if (matchup.playerScore && matchup.opponentScore) {
                                                    const playerNetScores = matchup.playerScore.matchNetHoleScores || matchup.playerScore.holeScores
                                                    const opponentNetScores = matchup.opponentScore.matchNetHoleScores || matchup.opponentScore.holeScores
                                                    
                                                    for (let i = 0; i < 9; i++) {
                                                        const playerNet = playerNetScores[i]
                                                        const opponentNet = opponentNetScores[i]
                                                        
                                                        if (playerNet < opponentNet) {
                                                            playerPoints.push(2)
                                                            opponentPoints.push(0)
                                                            playerTotal += 2
                                                        } else if (playerNet > opponentNet) {
                                                            playerPoints.push(0)
                                                            opponentPoints.push(2)
                                                            opponentTotal += 2
                                                        } else {
                                                            playerPoints.push(1)
                                                            opponentPoints.push(1)
                                                            playerTotal += 1
                                                            opponentTotal += 1
                                                        }
                                                    }
                                                    
                                                    // Use server-stored totals if available, otherwise calculate
                                                    if (storedPlayerTotal !== undefined && storedOpponentTotal !== undefined) {
                                                        playerTotal = storedPlayerTotal
                                                        opponentTotal = storedOpponentTotal
                                                    } else {
                                                        // Add overall match points (4 points for lower total)
                                                        const playerMatchNet = matchup.playerScore.matchNetScore ?? matchup.playerScore.netScore
                                                        const opponentMatchNet = matchup.opponentScore.matchNetScore ?? matchup.opponentScore.netScore
                                                        
                                                        if (playerMatchNet < opponentMatchNet) {
                                                            playerTotal += 4
                                                        } else if (playerMatchNet > opponentMatchNet) {
                                                            opponentTotal += 4
                                                        } else {
                                                            playerTotal += 2
                                                            opponentTotal += 2
                                                        }
                                                    }
                                                }
                                                
                                                return { playerPoints, opponentPoints, playerTotal, opponentTotal }
                                            }
                                            
                                            const { playerPoints, opponentPoints, playerTotal, opponentTotal } = calculateHolePoints()
                                            
                                            // Calculate cell colors for net scores based on hole results
                                            const getPlayerNetCellColors = (): ('win' | 'loss' | 'tie' | 'none')[] => {
                                                if (!matchup.playerScore || !matchup.opponentScore) return []
                                                const playerNetScores = matchup.playerScore.matchNetHoleScores || matchup.playerScore.holeScores
                                                const opponentNetScores = matchup.opponentScore.matchNetHoleScores || matchup.opponentScore.holeScores
                                                return playerNetScores.map((playerNet, i) => {
                                                    const opponentNet = opponentNetScores[i]
                                                    if (playerNet < opponentNet) return 'win'
                                                    if (playerNet > opponentNet) return 'loss'
                                                    return 'tie'
                                                })
                                            }
                                            
                                            const getOpponentNetCellColors = (): ('win' | 'loss' | 'tie' | 'none')[] => {
                                                if (!matchup.playerScore || !matchup.opponentScore) return []
                                                const playerNetScores = matchup.playerScore.matchNetHoleScores || matchup.playerScore.holeScores
                                                const opponentNetScores = matchup.opponentScore.matchNetHoleScores || matchup.opponentScore.holeScores
                                                return opponentNetScores.map((opponentNet, i) => {
                                                    const playerNet = playerNetScores[i]
                                                    if (opponentNet < playerNet) return 'win'
                                                    if (opponentNet > playerNet) return 'loss'
                                                    return 'tie'
                                                })
                                            }

                                            const playerNetCellColors = getPlayerNetCellColors()
                                            const opponentNetCellColors = getOpponentNetCellColors()
                                            
                                            // Build comprehensive scorecard rows - organized with player data grouped together
                                            const scorecardRows = matchup.playerScore && matchup.opponentScore ? [
                                                ...(course?.holePars ? [{ label: 'Par', scores: course.holePars, total: course.par, withBorder: false }] : []),
                                                ...(course?.holeHandicaps ? [{ label: 'Hole Hdcp', scores: course.holeHandicaps, total: '', withBorder: true }] : []),
                                                // Player rows grouped together
                                                { 
                                                    label: matchup.playerScore.playerAbsent ? 'Your Gross (Absent)' : 'Your Gross', 
                                                    scores: matchup.playerScore.holeScores, 
                                                    total: matchup.playerScore.grossScore, 
                                                    withBorder: false,
                                                    showGolfSymbols: !matchup.playerScore.playerAbsent && !!course?.holePars,
                                                    pars: course?.holePars
                                                },
                                                { label: 'Your Strokes', scores: matchup.playerScore.matchStrokes || EMPTY_STROKES_ARRAY, total: matchup.playerScore.strokesReceived, withBorder: false, color: 'var(--color-accent)' },
                                                { 
                                                    label: 'Your Net', 
                                                    scores: matchup.playerScore.matchNetHoleScores || matchup.playerScore.holeScores, 
                                                    total: matchup.playerScore.matchNetScore ?? matchup.playerScore.netScore, 
                                                    withBorder: false, 
                                                    cellColors: playerNetCellColors
                                                },
                                                { label: 'Your Points', scores: playerPoints, total: playerTotal, withBorder: true, color: 'var(--color-primary)', bgColor: 'rgba(16, 185, 129, 0.15)' },
                                                // Opponent rows grouped together
                                                { 
                                                    label: matchup.opponentScore.playerAbsent ? `${matchup.opponentName} Gross (Absent)` : `${matchup.opponentName} Gross`, 
                                                    scores: matchup.opponentScore.holeScores, 
                                                    total: matchup.opponentScore.grossScore, 
                                                    withBorder: false,
                                                    showGolfSymbols: !matchup.opponentScore.playerAbsent && !!course?.holePars,
                                                    pars: course?.holePars
                                                },
                                                { label: `${matchup.opponentName} Strokes`, scores: matchup.opponentScore.matchStrokes || EMPTY_STROKES_ARRAY, total: matchup.opponentScore.strokesReceived, withBorder: false, color: 'var(--color-warning)' },
                                                { 
                                                    label: `${matchup.opponentName} Net`, 
                                                    scores: matchup.opponentScore.matchNetHoleScores || matchup.opponentScore.holeScores, 
                                                    total: matchup.opponentScore.matchNetScore ?? matchup.opponentScore.netScore, 
                                                    withBorder: false, 
                                                    cellColors: opponentNetCellColors
                                                },
                                                { label: `${matchup.opponentName} Points`, scores: opponentPoints, total: opponentTotal, withBorder: false, color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.15)' }
                                            ] : []

                                            return (
                                                <ExpandableCard
                                                    key={matchup.match.id}
                                                    isExpanded={isExpanded}
                                                    onToggle={() => setExpandedMatchId(isExpanded ? null : matchup.match.id)}
                                                    header={
                                                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
                                                            <span className={`badge ${
                                                                matchup.result === 'won' ? 'badge-success' :
                                                                matchup.result === 'lost' ? 'badge-danger' :
                                                                matchup.result === 'tied' ? 'badge-secondary' :
                                                                'badge-primary'
                                                            }`}>
                                                                {matchup.result.toUpperCase()}
                                                            </span>
                                                            {matchup.playerScore?.playerAbsent && <AbsentBadge small />}
                                                            <div style={{ textAlign: 'left' }}>
                                                                <p style={{ fontWeight: '600', marginBottom: '0.25rem' }}>
                                                                    vs {matchup.opponentName}
                                                                    {matchup.opponentScore?.playerAbsent && (
                                                                        <span style={{ marginLeft: '0.5rem' }}><AbsentBadge small /></span>
                                                                    )}
                                                                </p>
                                                                <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                                    {formatDate(matchup.date)} • {matchup.courseName}
                                                                </p>
                                                            </div>
                                                        </div>
                                                    }
                                                    rightContent={
                                                        matchup.playerScore && matchup.opponentScore && (
                                                            <span style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>
                                                                {playerTotal} - {opponentTotal} pts
                                                            </span>
                                                        )
                                                    }
                                                >
                                                    {matchup.playerScore && matchup.opponentScore && (
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
                                                    )}
                                                </ExpandableCard>
                                            )
                                        })}
                                    </div>
                                )}
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    )
}
