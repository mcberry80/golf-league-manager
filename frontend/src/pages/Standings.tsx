import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type { StandingsEntry } from '../types'

export default function Standings() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()
    const [standings, setStandings] = useState<StandingsEntry[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    // Use leagueId from URL params, fallback to currentLeague from context
    const effectiveLeagueId = leagueId || currentLeague?.id

    useEffect(() => {
        async function loadStandings() {
            if (!effectiveLeagueId) return

            try {
                const data = await api.getStandings(effectiveLeagueId)
                // Sort by total points descending
                const sorted = data.sort((a, b) => b.totalPoints - a.totalPoints)
                setStandings(sorted)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load standings')
            } finally {
                setLoading(false)
            }
        }

        if (!leagueLoading) {
            loadStandings()
        }
    }, [effectiveLeagueId, leagueLoading])

    if (leagueLoading || loading) {
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

                <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h1>League Standings</h1>
                    {currentLeague && <p className="text-gray-400 mt-1">{currentLeague.name}</p>}
                </div>

                {error && (
                    <div className="alert alert-error">
                        {error}
                    </div>
                )}

                {!effectiveLeagueId ? (
                    <div className="alert alert-info">
                        Select a league to view standings.
                    </div>
                ) : (
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
                                            <tr key={entry.playerId}>
                                                <td>
                                                    <span style={{
                                                        fontWeight: 'bold',
                                                        color: index === 0 ? 'var(--color-warning)' : 'var(--color-text)'
                                                    }}>
                                                        {index === 0 && '🏆 '}#{index + 1}
                                                    </span>
                                                </td>
                                                <td style={{ fontWeight: '600' }}>
                                                    <Link 
                                                        to={`/profile/${entry.playerId}`}
                                                        style={{ 
                                                            color: 'var(--color-primary)', 
                                                            textDecoration: 'none'
                                                        }}
                                                    >
                                                        {entry.playerName}
                                                    </Link>
                                                </td>
                                                <td>{entry.matchesPlayed}</td>
                                                <td style={{ color: 'var(--color-accent)' }}>{entry.matchesWon}</td>
                                                <td style={{ color: 'var(--color-danger)' }}>{entry.matchesLost}</td>
                                                <td style={{ color: 'var(--color-text-muted)' }}>{entry.matchesTied}</td>
                                                <td>
                                                    <span style={{
                                                        fontWeight: 'bold',
                                                        fontSize: '1.1rem',
                                                        color: 'var(--color-primary)'
                                                    }}>
                                                        {entry.totalPoints}
                                                    </span>
                                                </td>
                                                <td>{entry.leagueHandicap?.toFixed(1) || 'N/A'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    )
}
