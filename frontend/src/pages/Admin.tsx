import { Link, useNavigate, useParams } from 'react-router-dom'
import { useEffect } from 'react'
import { useLeague } from '../contexts/LeagueContext'

export default function Admin() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading, selectLeague } = useLeague()
    const navigate = useNavigate()

    useEffect(() => {
        if (leagueId && (!currentLeague || currentLeague.id !== leagueId)) {
            selectLeague(leagueId)
        }
    }, [leagueId, currentLeague, selectLeague])

    useEffect(() => {
        if (!isLoading && !currentLeague && !leagueId) {
            navigate('/leagues')
        }
    }, [currentLeague, isLoading, navigate, leagueId])

    if (isLoading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!currentLeague) {
        return null // Will redirect in useEffect
    }

    if (userRole !== 'admin') {
        return (
            <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                <div className="alert alert-error">
                    <strong>Access Denied:</strong> You must be an admin of {currentLeague.name} to access this page.
                </div>
                <Link to="/" className="btn btn-secondary" style={{ marginTop: 'var(--spacing-lg)' }}>
                    Return Home
                </Link>
            </div>
        )
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <div style={{ marginBottom: 'var(--spacing-2xl)' }}>
                    <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                        ← Back to Home
                    </Link>
                    <h1 style={{ marginBottom: 'var(--spacing-sm)' }}>Admin Dashboard</h1>
                    <p style={{ color: 'var(--color-text-secondary)' }}>
                        Manage {currentLeague.name}
                    </p>
                </div>

                <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-xl)' }}>
                    <Link to={`/leagues/${currentLeague.id}/admin/league-setup`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>🏅</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>League Setup</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Create and manage seasons, set active season
                        </p>
                    </Link>

                    <Link to={`/leagues/${currentLeague.id}/admin/players`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>👥</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Players</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Add and manage league players
                        </p>
                    </Link>

                    <Link to={`/leagues/${currentLeague.id}/admin/courses`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>⛳</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Courses</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Add and manage golf courses
                        </p>
                    </Link>

                    <Link to={`/leagues/${currentLeague.id}/admin/matches`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>📅</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Match Scheduling</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Schedule matches and set matchups
                        </p>
                    </Link>

                    <Link to={`/leagues/${currentLeague.id}/admin/scores`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>✏️</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Score Entry</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Enter scores for matches with absence handling
                        </p>
                    </Link>

                    <Link to="/standings" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>📊</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>View Standings</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            View league standings and statistics
                        </p>
                    </Link>
                </div>
            </div>
        </div>
    )
}
