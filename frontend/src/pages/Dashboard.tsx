import { useState, useEffect, useCallback } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { SignedIn, SignedOut, SignInButton, UserButton } from '@clerk/clerk-react'
import { useLeague } from '../contexts/LeagueContext'
import { useCurrentUser } from '../hooks'
import { formatDateShort } from '../lib/utils'
import { Trophy, Calendar, MessageSquare, TrendingUp, ChevronRight } from 'lucide-react'
import api from '../lib/api'
import BulletinBoard from '../components/BulletinBoard'
import { LoadingSpinner } from '../components/Layout'
import type { StandingsEntry, Season, Match, LeagueMemberWithPlayer } from '../types'

export default function Dashboard() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const navigate = useNavigate()
    const { currentLeague, userRole, isLoading: leagueLoading } = useLeague()
    const { player: currentPlayer } = useCurrentUser()

    const [activeSeason, setActiveSeason] = useState<Season | null>(null)
    const [standings, setStandings] = useState<StandingsEntry[]>([])
    const [recentMatches, setRecentMatches] = useState<Match[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    const effectiveLeagueId = leagueId || currentLeague?.id

    useEffect(() => {
        async function loadDashboardData() {
            if (!effectiveLeagueId) return

            try {
                setLoading(true)
                setError('')

                // Load all data in parallel
                const [seasonsData, standingsData, matchesData, membersData] = await Promise.all([
                    api.listSeasons(effectiveLeagueId),
                    api.getStandings(effectiveLeagueId),
                    api.listMatches(effectiveLeagueId, 'completed'),
                    api.listLeagueMembers(effectiveLeagueId)
                ])

                // Find active season
                const active = seasonsData.find(s => s.active)
                setActiveSeason(active || null)

                // Set standings (already sorted by API)
                setStandings(standingsData.slice(0, 5)) // Top 5 for dashboard

                // Set recent matches (last 5 completed)
                const sortedMatches = matchesData
                    .sort((a, b) => new Date(b.matchDate).getTime() - new Date(a.matchDate).getTime())
                    .slice(0, 5)
                setRecentMatches(sortedMatches)

                setMembers(membersData)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load dashboard data')
            } finally {
                setLoading(false)
            }
        }

        if (!leagueLoading) {
            loadDashboardData()
        }
    }, [effectiveLeagueId, leagueLoading])

    // Memoize helper function to prevent recreation on each render
    const getPlayerName = useCallback((playerId: string): string => {
        const member = members.find(m => m.playerId === playerId)
        return member?.player?.name || 'Unknown'
    }, [members])

    if (leagueLoading || loading) {
        return <LoadingSpinner />
    }

    if (!effectiveLeagueId) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                    <div className="text-center">
                        <Trophy className="w-16 h-16 text-emerald-500 mx-auto mb-4" />
                        <h1 className="text-2xl font-bold text-white mb-2">No League Selected</h1>
                        <p className="text-gray-400 mb-6">Select a league to view your dashboard</p>
                        <button
                            onClick={() => navigate('/leagues')}
                            className="btn btn-primary"
                        >
                            Select League
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            {/* Header */}
            <header className="border-b" style={{ borderColor: 'var(--color-border)', background: 'rgba(30, 41, 59, 0.8)', backdropFilter: 'blur(10px)' }}>
                <div className="container" style={{ padding: 'var(--spacing-md) var(--spacing-lg)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div className="flex items-center gap-4">
                            <Link to="/" style={{ textDecoration: 'none' }}>
                                <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text', fontSize: '1.25rem' }}>
                                    ⛳ Golf League
                                </h2>
                            </Link>
                            <SignedIn>
                                {currentLeague && (
                                    <div className="hidden md:flex items-center text-gray-400 text-sm border-l border-gray-700 pl-4 ml-4">
                                        <Trophy className="w-4 h-4 mr-2 text-emerald-500" />
                                        <span className="text-gray-200 font-medium">{currentLeague.name}</span>
                                        <button
                                            onClick={() => navigate('/leagues')}
                                            className="btn btn-sm btn-outline ml-2"
                                            style={{ fontSize: '0.75rem', padding: '0.25rem 0.5rem', height: 'auto' }}
                                        >
                                            Switch
                                        </button>
                                    </div>
                                )}
                            </SignedIn>
                        </div>
                        <div>
                            <SignedOut>
                                <SignInButton mode="modal">
                                    <button className="btn btn-primary">Sign In</button>
                                </SignInButton>
                            </SignedOut>
                            <SignedIn>
                                <div className="flex items-center gap-4">
                                    <button
                                        onClick={() => navigate('/leagues')}
                                        className="btn btn-sm btn-outline md:hidden"
                                    >
                                        Leagues
                                    </button>
                                    <UserButton afterSignOutUrl="/" />
                                </div>
                            </SignedIn>
                        </div>
                    </div>
                </div>
            </header>

            {/* Main Content - Sports Page Layout */}
            <main className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-lg)', paddingBottom: 'var(--spacing-2xl)' }}>
                {error && (
                    <div className="alert alert-error" style={{ marginBottom: 'var(--spacing-lg)' }}>
                        {error}
                    </div>
                )}

                {/* Season Badge */}
                {activeSeason && (
                    <div style={{ marginBottom: 'var(--spacing-lg)', display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)', flexWrap: 'wrap' }}>
                        <span className="badge badge-success" style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}>
                            {activeSeason.name}
                        </span>
                        {userRole === 'admin' && (
                            <Link 
                                to={`/leagues/${effectiveLeagueId}/admin`}
                                className="btn btn-sm btn-outline"
                                style={{ fontSize: '0.75rem' }}
                            >
                                Admin Dashboard
                            </Link>
                        )}
                    </div>
                )}

                {/* Main Grid - Responsive */}
                <div style={{ 
                    display: 'grid', 
                    gridTemplateColumns: 'repeat(12, 1fr)',
                    gap: 'var(--spacing-lg)'
                }}>
                    {/* Left Column - Bulletin Board (takes more space on larger screens) */}
                    <div style={{ 
                        gridColumn: 'span 12',
                        order: 1
                    }} className="lg:col-span-7 lg:order-1">
                        {activeSeason ? (
                            <div className="card-glass" style={{ height: '450px', display: 'flex', flexDirection: 'column' }}>
                                <BulletinBoard
                                    leagueId={effectiveLeagueId}
                                    seasonId={activeSeason.id}
                                    currentPlayerId={currentPlayer?.id}
                                    isAdmin={userRole === 'admin'}
                                />
                            </div>
                        ) : (
                            <div className="card-glass" style={{ textAlign: 'center', padding: 'var(--spacing-xl)' }}>
                                <MessageSquare className="w-12 h-12 text-gray-500 mx-auto mb-4" />
                                <h3 className="text-lg font-semibold text-white mb-2">No Active Season</h3>
                                <p className="text-gray-400">The bulletin board will be available when a season is active.</p>
                            </div>
                        )}
                    </div>

                    {/* Right Column - Standings & Results */}
                    <div style={{ 
                        gridColumn: 'span 12',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 'var(--spacing-lg)',
                        order: 2
                    }} className="lg:col-span-5 lg:order-2">
                        {/* Standings Card */}
                        <div className="card-glass">
                            <div style={{ 
                                display: 'flex', 
                                justifyContent: 'space-between', 
                                alignItems: 'center',
                                marginBottom: 'var(--spacing-md)'
                            }}>
                                <h3 style={{ 
                                    margin: 0, 
                                    color: 'var(--color-text)', 
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    gap: '0.5rem',
                                    fontSize: '1.1rem'
                                }}>
                                    <Trophy className="w-5 h-5 text-yellow-500" />
                                    Standings
                                </h3>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/standings`}
                                    style={{ 
                                        color: 'var(--color-primary)', 
                                        textDecoration: 'none',
                                        display: 'flex',
                                        alignItems: 'center',
                                        fontSize: '0.875rem'
                                    }}
                                >
                                    View All <ChevronRight className="w-4 h-4" />
                                </Link>
                            </div>
                            
                            {standings.length === 0 ? (
                                <p style={{ color: 'var(--color-text-muted)', margin: 0 }}>No standings data yet.</p>
                            ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                    {standings.map((entry, index) => (
                                        <div 
                                            key={entry.playerId}
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'space-between',
                                                padding: '0.5rem 0.75rem',
                                                background: index === 0 ? 'rgba(245, 158, 11, 0.1)' : 'rgba(255, 255, 255, 0.02)',
                                                borderRadius: 'var(--radius-sm)',
                                                border: index === 0 ? '1px solid rgba(245, 158, 11, 0.3)' : '1px solid transparent'
                                            }}
                                        >
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                                <span style={{ 
                                                    fontWeight: 'bold',
                                                    color: index === 0 ? 'var(--color-warning)' : 'var(--color-text-muted)',
                                                    width: '1.5rem'
                                                }}>
                                                    {index === 0 && '🏆'}{index + 1}
                                                </span>
                                                <Link 
                                                    to={`/profile/${entry.playerId}`}
                                                    style={{ 
                                                        color: 'var(--color-text)', 
                                                        textDecoration: 'none',
                                                        fontWeight: index === 0 ? 600 : 400,
                                                        fontSize: '0.875rem'
                                                    }}
                                                >
                                                    {entry.playerName}
                                                </Link>
                                            </div>
                                            <span style={{ 
                                                fontWeight: 'bold',
                                                color: 'var(--color-primary)',
                                                fontSize: '0.875rem'
                                            }}>
                                                {entry.totalPoints} pts
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Recent Results Card */}
                        <div className="card-glass">
                            <div style={{ 
                                display: 'flex', 
                                justifyContent: 'space-between', 
                                alignItems: 'center',
                                marginBottom: 'var(--spacing-md)'
                            }}>
                                <h3 style={{ 
                                    margin: 0, 
                                    color: 'var(--color-text)', 
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    gap: '0.5rem',
                                    fontSize: '1.1rem'
                                }}>
                                    <Calendar className="w-5 h-5 text-blue-400" />
                                    Recent Results
                                </h3>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/results`}
                                    style={{ 
                                        color: 'var(--color-primary)', 
                                        textDecoration: 'none',
                                        display: 'flex',
                                        alignItems: 'center',
                                        fontSize: '0.875rem'
                                    }}
                                >
                                    View All <ChevronRight className="w-4 h-4" />
                                </Link>
                            </div>
                            
                            {recentMatches.length === 0 ? (
                                <p style={{ color: 'var(--color-text-muted)', margin: 0 }}>No completed matches yet.</p>
                            ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                    {recentMatches.map((match) => {
                                        const playerAName = getPlayerName(match.playerAId)
                                        const playerBName = getPlayerName(match.playerBId)
                                        const aWon = (match.playerAPoints || 0) > (match.playerBPoints || 0)
                                        const bWon = (match.playerBPoints || 0) > (match.playerAPoints || 0)
                                        const isTie = match.playerAPoints === match.playerBPoints

                                        return (
                                            <div 
                                                key={match.id}
                                                style={{
                                                    padding: '0.5rem 0.75rem',
                                                    background: 'rgba(255, 255, 255, 0.02)',
                                                    borderRadius: 'var(--radius-sm)',
                                                    border: '1px solid var(--color-border)'
                                                }}
                                            >
                                                <div style={{ 
                                                    display: 'flex', 
                                                    justifyContent: 'space-between',
                                                    alignItems: 'center',
                                                    fontSize: '0.75rem',
                                                    color: 'var(--color-text-muted)',
                                                    marginBottom: '0.25rem'
                                                }}>
                                                    <span>Week {match.weekNumber}</span>
                                                    <span>{formatDateShort(match.matchDate)}</span>
                                                </div>
                                                <div style={{ 
                                                    display: 'flex', 
                                                    justifyContent: 'space-between',
                                                    alignItems: 'center',
                                                    fontSize: '0.875rem'
                                                }}>
                                                    <span style={{ 
                                                        color: aWon ? 'var(--color-accent)' : 'var(--color-text)',
                                                        fontWeight: aWon ? 600 : 400
                                                    }}>
                                                        {playerAName}
                                                    </span>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <span style={{ 
                                                            fontWeight: 'bold',
                                                            color: aWon ? 'var(--color-accent)' : 'var(--color-text)'
                                                        }}>
                                                            {match.playerAPoints || 0}
                                                        </span>
                                                        <span style={{ color: 'var(--color-text-muted)' }}>-</span>
                                                        <span style={{ 
                                                            fontWeight: 'bold',
                                                            color: bWon ? 'var(--color-accent)' : 'var(--color-text)'
                                                        }}>
                                                            {match.playerBPoints || 0}
                                                        </span>
                                                    </div>
                                                    <span style={{ 
                                                        color: bWon ? 'var(--color-accent)' : 'var(--color-text)',
                                                        fontWeight: bWon ? 600 : 400
                                                    }}>
                                                        {playerBName}
                                                    </span>
                                                </div>
                                                {isTie && (
                                                    <div style={{ 
                                                        textAlign: 'center', 
                                                        fontSize: '0.75rem', 
                                                        color: 'var(--color-text-muted)',
                                                        marginTop: '0.25rem'
                                                    }}>
                                                        Tie
                                                    </div>
                                                )}
                                            </div>
                                        )
                                    })}
                                </div>
                            )}
                        </div>

                        {/* Quick Stats */}
                        <div className="card-glass">
                            <h3 style={{ 
                                margin: 0, 
                                color: 'var(--color-text)', 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '0.5rem',
                                fontSize: '1.1rem',
                                marginBottom: 'var(--spacing-md)'
                            }}>
                                <TrendingUp className="w-5 h-5 text-purple-400" />
                                Quick Links
                            </h3>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/standings`}
                                    className="btn btn-secondary"
                                    style={{ fontSize: '0.875rem', justifyContent: 'center' }}
                                >
                                    Full Standings
                                </Link>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/results`}
                                    className="btn btn-secondary"
                                    style={{ fontSize: '0.875rem', justifyContent: 'center' }}
                                >
                                    All Results
                                </Link>
                                {currentPlayer && (
                                    <Link 
                                        to={`/profile/${currentPlayer.id}`}
                                        className="btn btn-secondary"
                                        style={{ fontSize: '0.875rem', justifyContent: 'center', gridColumn: 'span 2' }}
                                    >
                                        My Profile
                                    </Link>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </main>

            {/* Add responsive styles */}
            <style>{`
                @media (min-width: 1024px) {
                    .lg\\:col-span-7 {
                        grid-column: span 7 !important;
                    }
                    .lg\\:col-span-5 {
                        grid-column: span 5 !important;
                    }
                    .lg\\:order-1 {
                        order: 1 !important;
                    }
                    .lg\\:order-2 {
                        order: 2 !important;
                    }
                }
            `}</style>
        </div>
    )
}
