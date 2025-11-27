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
                <div style={{ textAlign: 'center', marginBottom: 'var(--spacing-2xl)' }}>
                    <h1 style={{ fontSize: '3.5rem', marginBottom: 'var(--spacing-lg)' }}>
                        Golf League Manager
                    </h1>
                    <p style={{ fontSize: '1.25rem', color: 'var(--color-text-secondary)', maxWidth: '600px', margin: '0 auto' }}>
                        Manage your golf league with precision handicap calculations and comprehensive match play scoring
                    </p>
                </div>

                <SignedIn>
                    {!isLoading && !currentLeague ? (
                        <div className="text-center mb-12">
                            <div className="bg-gray-800/50 rounded-lg p-8 max-w-md mx-auto border border-gray-700">
                                <Trophy className="w-12 h-12 text-emerald-500 mx-auto mb-4" />
                                <h3 className="text-xl font-semibold text-white mb-2">No League Selected</h3>
                                <p className="text-gray-400 mb-6">Select a league to view stats and manage matches, or create a new one.</p>
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
                            <div className="text-center mb-8">
                                <Link 
                                    to={`/leagues/${currentLeague?.id}/dashboard`}
                                    className="btn btn-primary"
                                    style={{ padding: '1rem 2rem', fontSize: '1.1rem' }}
                                >
                                    <LayoutDashboard className="w-5 h-5 mr-2" />
                                    Go to Dashboard
                                </Link>
                                <p style={{ color: 'var(--color-text-muted)', marginTop: 'var(--spacing-sm)', fontSize: '0.875rem' }}>
                                    View standings, recent results, and talk trash with your league mates
                                </p>
                            </div>

                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-xl)', maxWidth: '900px', margin: '0 auto' }}>
                                <Link to={`/leagues/${currentLeague?.id}/dashboard`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>
                                        <MessageSquare className="w-10 h-10 text-blue-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Bulletin Board</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        Post messages, talk trash, and stay connected with your league
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/standings`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>
                                        <Trophy className="w-10 h-10 text-yellow-500" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Standings</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        View league standings, rankings, and player statistics
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/results`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>
                                        <Calendar className="w-10 h-10 text-purple-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Results</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        View match results and detailed scorecards by week
                                    </p>
                                </Link>

                                <Link to={currentPlayer ? `/profile/${currentPlayer.id}` : '/link-account'} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>
                                        <TrendingUp className="w-10 h-10 text-emerald-400" />
                                    </div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>My Profile</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        View your scores, handicap history, and match results
                                    </p>
                                </Link>

                                <Link to={`/leagues/${currentLeague?.id}/admin`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>⚙️</div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Admin</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        Manage league, players, courses, matches, and enter scores
                                    </p>
                                </Link>

                                <Link to="/link-account" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                    <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>🔗</div>
                                    <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Link Account</h3>
                                    <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                        Connect your account to your player profile
                                    </p>
                                </Link>
                            </div>
                        </>
                    )}
                </SignedIn>

                <SignedOut>
                    <div className="card-glass" style={{ maxWidth: '500px', margin: '0 auto', textAlign: 'center' }}>
                        <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)' }}>
                            Please sign in to access the league management features
                        </p>
                        <SignInButton mode="modal">
                            <button className="btn btn-primary">
                                Get Started
                            </button>
                        </SignInButton>
                    </div>
                </SignedOut>

                {/* Features */}
                <div style={{ marginTop: 'var(--spacing-2xl)', paddingTop: 'var(--spacing-2xl)', borderTop: '1px solid var(--color-border)' }}>
                    <h2 style={{ textAlign: 'center', marginBottom: 'var(--spacing-xl)', color: 'var(--color-text)' }}>Features</h2>
                    <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-lg)' }}>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-md)' }}>📊</div>
                            <h4 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>USGA Handicaps</h4>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                Automatic handicap calculation using USGA-compliant formulas
                            </p>
                        </div>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-md)' }}>⚡</div>
                            <h4 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Match Play</h4>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                Full 9-hole match play with stroke allocation and point calculation
                            </p>
                        </div>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-md)' }}>💬</div>
                            <h4 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Bulletin Board</h4>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                Stay connected with your league - post updates and talk trash
                            </p>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
