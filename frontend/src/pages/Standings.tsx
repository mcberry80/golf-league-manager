import { Link, useParams } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import { useStandings } from '../hooks'
import { LoadingSpinner } from '../components/Layout'

export default function Standings() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()

    // Use leagueId from URL params, fallback to currentLeague from context
    const effectiveLeagueId = leagueId || currentLeague?.id
    
    const { standings, loading, error } = useStandings(effectiveLeagueId)

    if (leagueLoading || loading) {
        return <LoadingSpinner />
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
                                            <th>Won</th>
                                            <th>Lost</th>
                                            <th>Tied</th>
                                            <th>Total Points</th>
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
