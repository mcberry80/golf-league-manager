import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../../lib/api'
import type { Match, Player, Course } from '../../types'

export default function ScoreEntry() {
    const { getToken } = useAuth()
    const [matches, setMatches] = useState<Match[]>([])
    const [selectedMatch, setSelectedMatch] = useState<Match | null>(null)
    const [players, setPlayers] = useState<Player[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)

    // Player absence status
    const [playerAAbsent, setPlayerAAbsent] = useState(false)
    const [playerBAbsent, setPlayerBAbsent] = useState(false)

    // Scores for 9 holes
    const [playerAScores, setPlayerAScores] = useState<number[]>(Array(9).fill(0))
    const [playerBScores, setPlayerBScores] = useState<number[]>(Array(9).fill(0))

    const [submitting, setSubmitting] = useState(false)

    useEffect(() => {
        loadData()
    }, [])

    async function loadData() {
        try {
            api.setAuthTokenProvider(getToken)
            const [matchesData, playersData, coursesData] = await Promise.all([
                api.listMatches('scheduled'),
                api.listPlayers(true),
                api.listCourses(),
            ])
            setMatches(matchesData)
            setPlayers(playersData)
            setCourses(coursesData)
        } catch (error) {
            console.error('Failed to load data:', error)
        } finally {
            setLoading(false)
        }
    }

    function handleMatchSelect(matchId: string) {
        const match = matches.find(m => m.id === matchId)
        setSelectedMatch(match || null)
        setPlayerAScores(Array(9).fill(0))
        setPlayerBScores(Array(9).fill(0))
        setPlayerAAbsent(false)
        setPlayerBAbsent(false)
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!selectedMatch) return

        setSubmitting(true)
        try {
            // For absent players, we still need to create score records
            // The backend will handle absence handicap calculation

            // Enter scores for all 9 holes for both players
            for (let hole = 1; hole <= 9; hole++) {
                // Player A score
                await api.enterScore({
                    match_id: selectedMatch.id,
                    player_id: selectedMatch.player_a_id,
                    hole_number: hole,
                    gross_score: playerAAbsent ? 0 : playerAScores[hole - 1],
                    net_score: 0, // Backend will calculate
                    strokes_received: 0, // Backend will calculate
                    player_absent: playerAAbsent,
                })

                // Player B score
                await api.enterScore({
                    match_id: selectedMatch.id,
                    player_id: selectedMatch.player_b_id,
                    hole_number: hole,
                    gross_score: playerBAbsent ? 0 : playerBScores[hole - 1],
                    net_score: 0, // Backend will calculate
                    strokes_received: 0, // Backend will calculate
                    player_absent: playerBAbsent,
                })
            }

            // Process the match to calculate points and update handicaps
            await api.processMatch(selectedMatch.id)

            alert('Scores entered successfully! Match completed and handicaps updated.')
            setSelectedMatch(null)
            loadData()
        } catch (error) {
            alert('Failed to enter scores: ' + (error instanceof Error ? error.message : 'Unknown error'))
        } finally {
            setSubmitting(false)
        }
    }

    const getPlayerName = (id: string) => players.find(p => p.id === id)?.name || 'Unknown'
    const getCourseName = (id: string) => courses.find(c => c.id === id)?.name || 'Unknown'

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
                <Link to="/admin" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Admin
                </Link>

                <h1 style={{ marginBottom: 'var(--spacing-xl)' }}>Score Entry</h1>

                {!selectedMatch ? (
                    <div className="card-glass">
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Select a Match</h3>
                        {matches.length === 0 ? (
                            <p style={{ color: 'var(--color-text-muted)' }}>No scheduled matches available.</p>
                        ) : (
                            <div className="table-container">
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th>Date</th>
                                            <th>Week</th>
                                            <th>Matchup</th>
                                            <th>Course</th>
                                            <th>Action</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {matches.map(match => (
                                            <tr key={match.id}>
                                                <td>{new Date(match.match_date).toLocaleDateString()}</td>
                                                <td>Week {match.week_number}</td>
                                                <td style={{ fontWeight: '600' }}>
                                                    {getPlayerName(match.player_a_id)} vs {getPlayerName(match.player_b_id)}
                                                </td>
                                                <td>{getCourseName(match.course_id)}</td>
                                                <td>
                                                    <button
                                                        onClick={() => handleMatchSelect(match.id)}
                                                        className="btn btn-primary"
                                                        style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}
                                                    >
                                                        Enter Scores
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>
                ) : (
                    <div className="card-glass">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                            <div>
                                <h3 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>
                                    {getPlayerName(selectedMatch.player_a_id)} vs {getPlayerName(selectedMatch.player_b_id)}
                                </h3>
                                <p style={{ color: 'var(--color-text-muted)' }}>
                                    {getCourseName(selectedMatch.course_id)} • Week {selectedMatch.week_number}
                                </p>
                            </div>
                            <button
                                onClick={() => setSelectedMatch(null)}
                                className="btn btn-secondary"
                            >
                                Cancel
                            </button>
                        </div>

                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-xl)' }}>
                                {/* Player A */}
                                <div>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-lg)' }}>
                                        <h4 style={{ color: 'var(--color-text)' }}>{getPlayerName(selectedMatch.player_a_id)}</h4>
                                        <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)', cursor: 'pointer' }}>
                                            <input
                                                type="checkbox"
                                                className="form-checkbox"
                                                checked={playerAAbsent}
                                                onChange={(e) => setPlayerAAbsent(e.target.checked)}
                                            />
                                            <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>Absent</span>
                                        </label>
                                    </div>
                                    {!playerAAbsent && (
                                        <div style={{ display: 'grid', gap: 'var(--spacing-sm)' }}>
                                            {[...Array(9)].map((_, i) => (
                                                <div key={i} className="form-group" style={{ marginBottom: 0 }}>
                                                    <label className="form-label" style={{ fontSize: '0.75rem' }}>Hole {i + 1}</label>
                                                    <input
                                                        type="number"
                                                        className="form-input"
                                                        value={playerAScores[i] || ''}
                                                        onChange={(e) => {
                                                            const newScores = [...playerAScores]
                                                            newScores[i] = parseInt(e.target.value) || 0
                                                            setPlayerAScores(newScores)
                                                        }}
                                                        required={!playerAAbsent}
                                                        min="1"
                                                        max="15"
                                                        style={{ padding: '0.5rem' }}
                                                    />
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                    {playerAAbsent && (
                                        <div className="alert alert-warning">
                                            Player marked as absent. Absence handicap will be applied automatically.
                                        </div>
                                    )}
                                </div>

                                {/* Player B */}
                                <div>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-lg)' }}>
                                        <h4 style={{ color: 'var(--color-text)' }}>{getPlayerName(selectedMatch.player_b_id)}</h4>
                                        <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)', cursor: 'pointer' }}>
                                            <input
                                                type="checkbox"
                                                className="form-checkbox"
                                                checked={playerBAbsent}
                                                onChange={(e) => setPlayerBAbsent(e.target.checked)}
                                            />
                                            <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>Absent</span>
                                        </label>
                                    </div>
                                    {!playerBAbsent && (
                                        <div style={{ display: 'grid', gap: 'var(--spacing-sm)' }}>
                                            {[...Array(9)].map((_, i) => (
                                                <div key={i} className="form-group" style={{ marginBottom: 0 }}>
                                                    <label className="form-label" style={{ fontSize: '0.75rem' }}>Hole {i + 1}</label>
                                                    <input
                                                        type="number"
                                                        className="form-input"
                                                        value={playerBScores[i] || ''}
                                                        onChange={(e) => {
                                                            const newScores = [...playerBScores]
                                                            newScores[i] = parseInt(e.target.value) || 0
                                                            setPlayerBScores(newScores)
                                                        }}
                                                        required={!playerBAbsent}
                                                        min="1"
                                                        max="15"
                                                        style={{ padding: '0.5rem' }}
                                                    />
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                    {playerBAbsent && (
                                        <div className="alert alert-warning">
                                            Player marked as absent. Absence handicap will be applied automatically.
                                        </div>
                                    )}
                                </div>
                            </div>

                            <div style={{ marginTop: 'var(--spacing-xl)', paddingTop: 'var(--spacing-xl)', borderTop: '1px solid var(--color-border)' }}>
                                <button
                                    type="submit"
                                    className="btn btn-success"
                                    disabled={submitting}
                                    style={{ width: '100%' }}
                                >
                                    {submitting ? 'Submitting...' : 'Submit Scores & Complete Match'}
                                </button>
                                <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginTop: 'var(--spacing-md)', textAlign: 'center' }}>
                                    This will calculate net scores, match points (22 total), and update handicaps automatically.
                                </p>
                            </div>
                        </form>
                    </div>
                )}
            </div>
        </div>
    )
}
