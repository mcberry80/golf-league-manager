import { SignInButton, SignedIn, SignedOut, UserButton } from '@clerk/clerk-react'
import { Link, useNavigate } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import { useCurrentUser } from '../hooks'
import { Trophy, LayoutDashboard, MessageSquare, Calendar, TrendingUp } from 'lucide-react'

export default function Home() {
    const { currentLeague, isLoading } = useLeague();
    const navigate = useNavigate();
    const { player: currentPlayer } = useCurrentUser();

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            {/* Header */}
            <header className="border-b" style={{ borderColor: 'var(--color-border)', background: 'rgba(30, 41, 59, 0.8)', backdropFilter: 'blur(10px)' }}>
                <div className="container" style={{ padding: 'var(--spacing-lg)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div className="flex items-center gap-4">
                            <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>
                                ⛳ Golf League
                            </h2>
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
                                    <button className="btn btn-primary">
                                        Sign In
                                    </button>
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

            {/* Hero Section */}
            <main className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <div style={{ textAlign: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <h1 style={{ marginBottom: 'var(--spacing-sm)' }}>
                        Golf League Manager
                    </h1>
                </div>

                <SignedIn>
                    {!isLoading && !currentLeague ? (
                        <div className="text-center mb-12">
                            <div className="bg-gray-800/50 rounded-lg p-8 max-w-md mx-auto border border-gray-700">
                                <Trophy className="w-12 h-12 text-emerald-500 mx-auto mb-4" />
                                <h3 className="text-xl font-semibold text-white mb-2">No League Selected</h3>
                                <p className="text-gray-400 mb-6">Select or create a league to get started.</p>
                                <button
                                    onClick={() => navigate('/leagues')}
                                    className="btn btn-primary w-full justify-center"
                                >
                                    Select or Create League
                                </button>
                            </div>
                        </div>
                    ) : (
                        <>
                            {/* Primary CTA - Go to Dashboard */}
                            <div className="text-center mb-6">
                                <Link 
                                    to={`/leagues/${currentLeague?.id}/dashboard`}
                                    className="btn btn-primary"
                                    style={{ padding: '0.75rem 1.5rem' }}
                                >
                                    <LayoutDashboard className="w-4 h-4 mr-2" />
                                    Go to Dashboard
                                </Link>
                            </div>

                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)', maxWidth: '800px', margin: '0 auto' }}>
                                <Link to={`/leagues/${currentLeague?.id}/dashboard`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)' }}>
                                        <MessageSquare className="w-6 h-6 text-blue-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>Bulletin Board</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        League announcements and messages
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/standings`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)' }}>
                                        <Trophy className="w-6 h-6 text-yellow-500" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>Standings</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        League rankings and statistics
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/results`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)' }}>
                                        <Calendar className="w-6 h-6 text-purple-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>Results</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        Match results and scorecards
                                    </p>
                                </Link>

                                <Link to={currentPlayer ? `/profile/${currentPlayer.id}` : '/link-account'} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)' }}>
                                        <TrendingUp className="w-6 h-6 text-emerald-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>My Profile</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        Scores and handicap history
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/admin`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)', fontSize: '1.5rem' }}>⚙️</div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>Admin</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        League management
                                    </p>
                                </Link>

                                <Link to="/link-account" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ marginBottom: 'var(--spacing-sm)', fontSize: '1.5rem' }}>🔗</div>
                                    <h3 style={{ marginBottom: 'var(--spacing-xs)', color: 'var(--color-text)', fontSize: '1rem' }}>Link Account</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.813rem' }}>
                                        Connect to player profile
                                    </p>
                                </Link>
                            </div>
                        </>
                    )}
                </SignedIn>

                <SignedOut>
                    <div className="card-glass" style={{ maxWidth: '500px', margin: '0 auto', textAlign: 'center' }}>
                        <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)' }}>
                            Sign in to access your league
                        </p>
                        <SignInButton mode="modal">
                            <button className="btn btn-primary">
                                Sign In
                            </button>
                        </SignInButton>
                    </div>
                </SignedOut>

            </main>
        </div>
    )
}
