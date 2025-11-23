import { SignInButton, SignedIn, SignedOut, UserButton } from '@clerk/clerk-react'
import { Link, useNavigate } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import { Trophy } from 'lucide-react'

export default function Home() {
    const { currentLeague, isLoading } = useLeague();
    const navigate = useNavigate();

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
                                            className="ml-2 text-xs text-emerald-500 hover:text-emerald-400 transition-colors"
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
                                        className="md:hidden text-sm text-emerald-500 hover:text-emerald-400"
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
                        <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-xl)', maxWidth: '900px', margin: '0 auto' }}>
                            <Link to={`/leagues/${currentLeague?.id}/admin`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>⚙️</div>
                                <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Admin</h3>
                                <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                    Manage league, players, courses, matches, and enter scores
                                </p>
                            </Link>

                            <Link to="/standings" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>🏆</div>
                                <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Standings</h3>
                                <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                    View league standings, rankings, and player statistics
                                </p>
                            </Link>

                            <Link to="/players" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                                <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>👤</div>
                                <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>My Profile</h3>
                                <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                    View your scores, handicap history, and match results
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
                            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-md)' }}>📱</div>
                            <h4 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Real-time Updates</h4>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                Instant handicap recalculation after each round
                            </p>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
