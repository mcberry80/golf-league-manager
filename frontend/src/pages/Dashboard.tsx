import { useState, useEffect } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { SignedIn, UserButton } from '@clerk/clerk-react'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type { 
    StandingsEntry, 
    Match, 
    BulletinPost, 
    Season, 
    LeagueMemberWithPlayer,
    Player 
} from '../types'
import { Trophy, MessageSquare, Calendar, Send, Trash2, ChevronRight, Users } from 'lucide-react'

export default function Dashboard() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()
    const navigate = useNavigate()
    
    const [standings, setStandings] = useState<StandingsEntry[]>([])
    const [recentMatches, setRecentMatches] = useState<Match[]>([])
    const [bulletinPosts, setBulletinPosts] = useState<BulletinPost[]>([])
    const [activeSeason, setActiveSeason] = useState<Season | null>(null)
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null)
    
    const [newMessage, setNewMessage] = useState('')
    const [posting, setPosting] = useState(false)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    const effectiveLeagueId = leagueId || currentLeague?.id

    // Load current user's player info
    useEffect(() => {
        async function loadCurrentPlayer() {
            try {
                const userInfo = await api.getCurrentUser();
                if (userInfo.linked && userInfo.player) {
                    setCurrentPlayer(userInfo.player);
                }
            } catch {
                // User not linked yet
            }
        }
        loadCurrentPlayer();
    }, []);

    // Load all dashboard data
    useEffect(() => {
        async function loadDashboardData() {
            if (!effectiveLeagueId) return

            try {
                setLoading(true)
                setError('')

                // Load standings, matches, members, and active season in parallel
                const [standingsData, matchesData, membersData] = await Promise.all([
                    api.getStandings(effectiveLeagueId),
                    api.listMatches(effectiveLeagueId, 'completed'),
                    api.listLeagueMembers(effectiveLeagueId)
                ])

                // Sort standings by total points
                const sortedStandings = standingsData.sort((a, b) => b.totalPoints - a.totalPoints)
                setStandings(sortedStandings.slice(0, 5)) // Top 5 for dashboard

                // Sort matches by date and take recent ones
                const sortedMatches = matchesData.sort((a, b) => 
                    new Date(b.matchDate).getTime() - new Date(a.matchDate).getTime()
                )
                setRecentMatches(sortedMatches.slice(0, 5)) // Last 5 matches

                setMembers(membersData)

                // Try to get active season
                try {
                    const season = await api.getActiveSeason(effectiveLeagueId)
                    setActiveSeason(season)
                    
                    // Load bulletin posts for active season
                    const posts = await api.listBulletinPosts(effectiveLeagueId, season.id, 20)
                    setBulletinPosts(posts)
                } catch {
                    // No active season, that's okay
                    setActiveSeason(null)
                    setBulletinPosts([])
                }

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

    // Get player name by ID
    const getPlayerName = (playerId: string): string => {
        const member = members.find(m => m.playerId === playerId)
        return member?.player?.name || 'Unknown'
    }

    // Format date for display
    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric' 
        })
    }

    // Format timestamp for bulletin posts
    const formatTimestamp = (dateString: string) => {
        const date = new Date(dateString)
        const now = new Date()
        const diffMs = now.getTime() - date.getTime()
        const diffMins = Math.floor(diffMs / 60000)
        const diffHours = Math.floor(diffMs / 3600000)
        const diffDays = Math.floor(diffMs / 86400000)

        if (diffMins < 1) return 'Just now'
        if (diffMins < 60) return `${diffMins}m ago`
        if (diffHours < 24) return `${diffHours}h ago`
        if (diffDays < 7) return `${diffDays}d ago`
        
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric',
            hour: 'numeric',
            minute: '2-digit'
        })
    }

    // Handle posting new message
    const handlePostMessage = async () => {
        if (!newMessage.trim() || !effectiveLeagueId || !activeSeason) return

        setPosting(true)
        try {
            const post = await api.createBulletinPost(effectiveLeagueId, activeSeason.id, newMessage.trim())
            setBulletinPosts([post, ...bulletinPosts])
            setNewMessage('')
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to post message')
        } finally {
            setPosting(false)
        }
    }

    // Handle deleting a message
    const handleDeletePost = async (postId: string) => {
        if (!effectiveLeagueId || !activeSeason) return

        try {
            await api.deleteBulletinPost(effectiveLeagueId, activeSeason.id, postId)
            setBulletinPosts(bulletinPosts.filter(p => p.id !== postId))
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to delete message')
        }
    }

    // Get match result display
    const getMatchResultDisplay = (match: Match): string => {
        const playerAName = getPlayerName(match.playerAId)
        const playerBName = getPlayerName(match.playerBId)
        const aPoints = match.playerAPoints ?? 0
        const bPoints = match.playerBPoints ?? 0
        
        if (aPoints > bPoints) {
            return `${playerAName} def. ${playerBName} (${aPoints}-${bPoints})`
        } else if (bPoints > aPoints) {
            return `${playerBName} def. ${playerAName} (${bPoints}-${aPoints})`
        } else {
            return `${playerAName} tied ${playerBName} (${aPoints}-${bPoints})`
        }
    }

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!effectiveLeagueId) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                    <div className="text-center">
                        <Trophy className="w-16 h-16 text-emerald-500 mx-auto mb-4" />
                        <h2 className="text-2xl font-bold text-white mb-4">No League Selected</h2>
                        <p className="text-gray-400 mb-6">Select a league to view your dashboard.</p>
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
                <div className="container" style={{ padding: 'var(--spacing-lg)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div className="flex items-center gap-4">
                            <Link to="/" style={{ textDecoration: 'none' }}>
                                <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>
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
                        <div className="flex items-center gap-4">
                            <button
                                onClick={() => navigate('/leagues')}
                                className="btn btn-sm btn-outline md:hidden"
                            >
                                Leagues
                            </button>
                            <SignedIn>
                                <UserButton afterSignOutUrl="/" />
                            </SignedIn>
                        </div>
                    </div>
                </div>
            </header>

            <main className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                {error && (
                    <div className="alert alert-error mb-6">
                        {error}
                    </div>
                )}

                {/* Main Dashboard Grid - Sports Page Layout */}
                <div className="grid gap-6" style={{ 
                    gridTemplateColumns: 'repeat(12, 1fr)',
                    gridTemplateRows: 'auto'
                }}>
                    {/* Bulletin Board - Main Feature (Left Column on Desktop) */}
                    <div style={{ gridColumn: 'span 12' }} className="lg:col-span-8">
                        <div className="card-glass" style={{ height: '100%' }}>
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="flex items-center gap-2 text-lg font-bold text-white">
                                    <MessageSquare className="w-5 h-5 text-blue-400" />
                                    League Talk
                                    {activeSeason && (
                                        <span className="badge badge-primary ml-2">{activeSeason.name}</span>
                                    )}
                                </h3>
                            </div>

                            {!activeSeason ? (
                                <p className="text-gray-400">No active season. Bulletin board will be available when a season is active.</p>
                            ) : (
                                <>
                                    {/* New Post Form */}
                                    <div className="mb-4">
                                        <div className="flex gap-2">
                                            <textarea
                                                value={newMessage}
                                                onChange={(e) => setNewMessage(e.target.value)}
                                                placeholder="Talk some trash, share news, or congratulate a great round..."
                                                className="form-textarea flex-1"
                                                style={{ 
                                                    minHeight: '60px', 
                                                    resize: 'none',
                                                    fontSize: '0.95rem'
                                                }}
                                                maxLength={1000}
                                                onKeyDown={(e) => {
                                                    if (e.key === 'Enter' && !e.shiftKey) {
                                                        e.preventDefault()
                                                        handlePostMessage()
                                                    }
                                                }}
                                            />
                                            <button
                                                onClick={handlePostMessage}
                                                disabled={posting || !newMessage.trim()}
                                                className="btn btn-primary self-end"
                                                style={{ height: '44px', padding: '0 1rem' }}
                                            >
                                                {posting ? (
                                                    <div className="spinner" style={{ width: '20px', height: '20px', borderWidth: '2px' }}></div>
                                                ) : (
                                                    <Send className="w-5 h-5" />
                                                )}
                                            </button>
                                        </div>
                                        <div className="text-xs text-gray-500 mt-1 text-right">
                                            {newMessage.length}/1000
                                        </div>
                                    </div>

                                    {/* Posts List */}
                                    <div className="space-y-3" style={{ 
                                        maxHeight: '400px', 
                                        overflowY: 'auto',
                                        paddingRight: 'var(--spacing-sm)'
                                    }}>
                                        {bulletinPosts.length === 0 ? (
                                            <p className="text-gray-400 text-center py-8">
                                                No posts yet. Be the first to share something!
                                            </p>
                                        ) : (
                                            bulletinPosts.map((post) => (
                                                <div 
                                                    key={post.id}
                                                    className="p-3 rounded-lg"
                                                    style={{ 
                                                        background: 'rgba(255, 255, 255, 0.03)',
                                                        border: '1px solid rgba(255, 255, 255, 0.05)'
                                                    }}
                                                >
                                                    <div className="flex justify-between items-start gap-2">
                                                        <div className="flex-1 min-w-0">
                                                            <div className="flex items-center gap-2 mb-1">
                                                                <span className="font-semibold text-white text-sm">
                                                                    {post.playerName}
                                                                </span>
                                                                <span className="text-xs text-gray-500">
                                                                    {formatTimestamp(post.createdAt)}
                                                                </span>
                                                            </div>
                                                            <p className="text-gray-300 text-sm whitespace-pre-wrap break-words">
                                                                {post.message}
                                                            </p>
                                                        </div>
                                                        {currentPlayer && post.playerId === currentPlayer.id && (
                                                            <button
                                                                onClick={() => handleDeletePost(post.id)}
                                                                className="text-gray-500 hover:text-red-400 transition-colors p-1"
                                                                title="Delete post"
                                                            >
                                                                <Trash2 className="w-4 h-4" />
                                                            </button>
                                                        )}
                                                    </div>
                                                </div>
                                            ))
                                        )}
                                    </div>
                                </>
                            )}
                        </div>
                    </div>

                    {/* Right Sidebar - Standings & Results */}
                    <div style={{ gridColumn: 'span 12' }} className="lg:col-span-4 space-y-6">
                        {/* Quick Standings */}
                        <div className="card-glass">
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="flex items-center gap-2 text-lg font-bold text-white">
                                    <Trophy className="w-5 h-5 text-yellow-400" />
                                    Standings
                                </h3>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/standings`}
                                    className="text-blue-400 text-sm hover:text-blue-300 flex items-center gap-1"
                                >
                                    View All <ChevronRight className="w-4 h-4" />
                                </Link>
                            </div>

                            {standings.length === 0 ? (
                                <p className="text-gray-400 text-sm">No standings data yet.</p>
                            ) : (
                                <div className="space-y-2">
                                    {standings.map((entry, index) => (
                                        <div 
                                            key={entry.playerId}
                                            className="flex items-center justify-between p-2 rounded"
                                            style={{ 
                                                background: index === 0 ? 'rgba(234, 179, 8, 0.1)' : 'transparent'
                                            }}
                                        >
                                            <div className="flex items-center gap-3">
                                                <span className={`font-bold w-6 text-center ${
                                                    index === 0 ? 'text-yellow-400' : 'text-gray-400'
                                                }`}>
                                                    {index === 0 ? '🏆' : `#${index + 1}`}
                                                </span>
                                                <Link 
                                                    to={`/profile/${entry.playerId}`}
                                                    className="text-white hover:text-blue-400 font-medium text-sm truncate"
                                                    style={{ maxWidth: '120px' }}
                                                >
                                                    {entry.playerName}
                                                </Link>
                                            </div>
                                            <span className="font-bold text-emerald-400">
                                                {entry.totalPoints} pts
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Recent Results */}
                        <div className="card-glass">
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="flex items-center gap-2 text-lg font-bold text-white">
                                    <Calendar className="w-5 h-5 text-green-400" />
                                    Recent Results
                                </h3>
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/results`}
                                    className="text-blue-400 text-sm hover:text-blue-300 flex items-center gap-1"
                                >
                                    View All <ChevronRight className="w-4 h-4" />
                                </Link>
                            </div>

                            {recentMatches.length === 0 ? (
                                <p className="text-gray-400 text-sm">No completed matches yet.</p>
                            ) : (
                                <div className="space-y-3">
                                    {recentMatches.map((match) => (
                                        <div 
                                            key={match.id}
                                            className="p-2 rounded text-sm"
                                            style={{ 
                                                background: 'rgba(255, 255, 255, 0.03)'
                                            }}
                                        >
                                            <div className="flex items-center gap-2 text-gray-400 text-xs mb-1">
                                                <Calendar className="w-3 h-3" />
                                                {formatDate(match.matchDate)}
                                            </div>
                                            <p className="text-gray-200 text-sm">
                                                {getMatchResultDisplay(match)}
                                            </p>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Quick Links */}
                        <div className="card-glass">
                            <h3 className="flex items-center gap-2 text-lg font-bold text-white mb-4">
                                <Users className="w-5 h-5 text-purple-400" />
                                Quick Links
                            </h3>
                            <div className="space-y-2">
                                <Link 
                                    to={`/leagues/${effectiveLeagueId}/admin`}
                                    className="block p-3 rounded bg-gray-800/50 hover:bg-gray-700/50 transition-colors text-gray-200 text-sm"
                                >
                                    ⚙️ League Admin
                                </Link>
                                {currentPlayer && (
                                    <Link 
                                        to={`/profile/${currentPlayer.id}`}
                                        className="block p-3 rounded bg-gray-800/50 hover:bg-gray-700/50 transition-colors text-gray-200 text-sm"
                                    >
                                        👤 My Profile
                                    </Link>
                                )}
                                <Link 
                                    to="/link-account"
                                    className="block p-3 rounded bg-gray-800/50 hover:bg-gray-700/50 transition-colors text-gray-200 text-sm"
                                >
                                    🔗 Link Account
                                </Link>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
