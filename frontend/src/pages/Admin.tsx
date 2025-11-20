import { Link } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { useAuth } from '@clerk/clerk-react'
import api from '../lib/api'

export default function Admin() {
    const { getToken } = useAuth()
    const [isAdmin, setIsAdmin] = useState(false)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        async function checkAdmin() {
            try {
                api.setAuthTokenProvider(getToken)
                const userInfo = await api.getCurrentUser()
                setIsAdmin(userInfo.player?.is_admin || false)
            } catch (error) {
                console.error('Failed to check admin status:', error)
            } finally {
                setLoading(false)
            }
        }
        checkAdmin()
    }, [getToken])

    if (loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!isAdmin) {
        return (
            <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                <div className="alert alert-error">
                    <strong>Access Denied:</strong> You must be an admin to access this page.
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
                        Manage your golf league, players, courses, and matches
                    </p>
                </div>

                <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-xl)' }}>
                    <Link to="/admin/league-setup" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>🏅</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>League Setup</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Create and manage seasons, set active season
                        </p>
                    </Link>

                    <Link to="/admin/players" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>👥</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Players</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Add and manage league players
                        </p>
                    </Link>

                    <Link to="/admin/courses" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>⛳</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Courses</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Add and manage golf courses
                        </p>
                    </Link>

                    <Link to="/admin/matches" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
                        <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-md)' }}>📅</div>
                        <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>Match Scheduling</h3>
                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                            Schedule matches and set matchups
                        </p>
                    </Link>

                    <Link to="/admin/scores" className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
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
