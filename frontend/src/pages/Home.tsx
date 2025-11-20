import { SignInButton, SignedIn, SignedOut, UserButton } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'

export default function Home() {
    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            {/* Header */}
            <header className="border-b" style={{ borderColor: 'var(--color-border)', background: 'rgba(30, 41, 59, 0.8)', backdropFilter: 'blur(10px)' }}>
                <div className="container" style={{ padding: 'var(--spacing-lg)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>
                            ⛳ Golf League
                        </h2>
                        <div>
                            <SignedOut>
                                <SignInButton mode="modal">
                                    <button className="btn btn-primary">
                                        Sign In
                                    </button>
                                </SignInButton>
                            </SignedOut>
                            <SignedIn>
                                <UserButton afterSignOutUrl="/" />
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
                    <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-xl)', maxWidth: '900px', margin: '0 auto' }}>
                        <Link to="/admin" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
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
