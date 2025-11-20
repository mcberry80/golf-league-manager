import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../lib/api'
import type { StandingsEntry } from '../types'

export default function Standings() {
    const { getToken } = useAuth()
    const [standings, setStandings] = useState<StandingsEntry[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    useEffect(() => {
        async function loadStandings() {
            try {
                api.setAuthTokenProvider(getToken)
                const data = await api.getStandings()
                // Sort by total points descending
                const sorted = data.sort((a, b) => b.total_points - a.total_points)
                setStandings(sorted)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load standings')
            } finally {
                setLoading(false)
            }
        }
        loadStandings()
    }, [getToken])

    if (loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Home
                </Link>

                <h1 style={{ marginBottom: 'var(--spacing-xl)' }}>League Standings</h1>

                {error && (
                    <div className="alert alert-error">
                        {error}
                    </div>
                )}

                <div className="card-glass">
                    {standings.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No standings data available yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Rank</th>
                                        <th>Player</th>
                                        <th>Matches Played</th>
                                        <th>Won</th>
                                        <th>Lost</th>
                                        <th>Tied</th>
                                        <th>Total Points</th>
                                        <th>Handicap</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {standings.map((entry, index) => (
                                        <tr key={entry.player_id}>
                                            <td>
                                                <span style={{
                                                    fontWeight: 'bold',
                                                    color: index === 0 ? 'var(--color-warning)' : 'var(--color-text)'
                                                }}>
                                                    {index === 0 && '🏆 '}#{index + 1}
                                                </span>
                                            </td>
                                            <td style={{ fontWeight: '600' }}>{entry.player_name}</td>
                                            <td>{entry.matches_played}</td>
                                            <td style={{ color: 'var(--color-accent)' }}>{entry.matches_won}</td>
                                            <td style={{ color: 'var(--color-danger)' }}>{entry.matches_lost}</td>
                                            <td style={{ color: 'var(--color-text-muted)' }}>{entry.matches_tied}</td>
                                            <td>
                                                <span style={{
                                                    fontWeight: 'bold',
                                                    fontSize: '1.1rem',
                                                    color: 'var(--color-primary)'
                                                }}>
                                                    {entry.total_points}
                                                </span>
                                            </td>
                                            <td>{entry.league_handicap.toFixed(1)}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
