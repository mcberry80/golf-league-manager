import { useState, useEffect } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Match, LeagueMemberWithPlayer, Course } from '../../types'

export default function ScoreEntry() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [matches, setMatches] = useState<Match[]>([])
    const [selectedMatch, setSelectedMatch] = useState<Match | null>(null)
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
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
        if (leagueId && (!currentLeague || currentLeague.id !== leagueId)) {
            selectLeague(leagueId)
        }
    }, [leagueId, currentLeague, selectLeague])

    useEffect(() => {
        if (!leagueLoading && !currentLeague && !leagueId) {
            navigate('/leagues')
            return
        }

        if (currentLeague) {
            loadData()
        }
    }, [currentLeague, leagueLoading, navigate, leagueId])

    async function loadData() {
        if (!currentLeague) return

        try {
            const [matchesData, membersData, coursesData] = await Promise.all([
                api.listMatches(currentLeague.id),
                api.listLeagueMembers(currentLeague.id),
                api.listCourses(currentLeague.id),
            ])
            setMatches(matchesData.filter(m => m.status === 'scheduled'))
            setMembers(membersData)
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

        // Get course par for default values
        const course = match ? courses.find(c => c.id === match.courseId) : null
        const defaultScores = course?.holePars || Array(9).fill(4)

        setPlayerAScores(defaultScores)
        setPlayerBScores(defaultScores)
        setPlayerAAbsent(false)
        setPlayerBAbsent(false)
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!selectedMatch || !currentLeague) return

        setSubmitting(true)
        try {
            // Batch all scores into a single API call for each player
            const playerAScoreData = playerAScores.map((score, index) => ({
                matchId: selectedMatch.id,
                playerId: selectedMatch.playerAId,
                holeNumber: index + 1,
                grossScore: playerAAbsent ? 0 : score,
                netScore: 0, // Backend will calculate
                strokesReceived: 0, // Backend will calculate
                playerAbsent: playerAAbsent,
            }))

            const playerBScoreData = playerBScores.map((score, index) => ({
                matchId: selectedMatch.id,
                playerId: selectedMatch.playerBId,
                holeNumber: index + 1,
                grossScore: playerBAbsent ? 0 : score,
                netScore: 0, // Backend will calculate
                strokesReceived: 0, // Backend will calculate
                playerAbsent: playerBAbsent,
            }))

            // Submit all scores in parallel (2 API calls instead of 18)
            await Promise.all([
                api.enterScoreBatch(currentLeague.id, playerAScoreData),
                api.enterScoreBatch(currentLeague.id, playerBScoreData),
            ])

            // Process the match to calculate points and update handicaps
            await api.processMatch(currentLeague.id, selectedMatch.id)

            alert('Scores entered successfully! Match completed and handicaps updated.')
            setSelectedMatch(null)
            loadData()
        } catch (error) {
            alert('Failed to enter scores: ' + (error instanceof Error ? error.message : 'Unknown error'))
        } finally {
            setSubmitting(false)
        }
    }

    const getPlayerName = (id: string) => {
        const member = members.find(m => m.playerId === id)
        return member?.player?.name || 'Unknown'
    }
    const getCourseName = (id: string) => courses.find(c => c.id === id)?.name || 'Unknown'
    const getCourse = (id: string) => courses.find(c => c.id === id)

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!currentLeague || userRole !== 'admin') {
        return (
            <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                <div className="alert alert-error">
                    <strong>Access Denied:</strong> You must be an admin of {currentLeague?.name || 'this league'} to access this page.
                </div>
                <Link to="/" className="btn btn-secondary" style={{ marginTop: 'var(--spacing-lg)' }}>
                    Return Home
                </Link>
            </div>
        )
    }

    const course = selectedMatch ? getCourse(selectedMatch.courseId) : null

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to={`/leagues/${currentLeague.id}/admin`} style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Admin
                </Link>

                <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h1>Score Entry</h1>
                    <p className="text-gray-400 mt-1">{currentLeague.name}</p>
                </div>

                <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Select Match</h3>
                    {matches.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No scheduled matches available.</p>
                    ) : (
                        <select
                            className="form-input"
                            value={selectedMatch?.id || ''}
                            onChange={(e) => handleMatchSelect(e.target.value)}
                            style={{ maxWidth: '500px' }}
                        >
                            <option value="">-- Select a match --</option>
                            {matches.map((match) => (
                                <option key={match.id} value={match.id}>
                                    Week {match.weekNumber}: {getPlayerName(match.playerAId)} vs {getPlayerName(match.playerBId)} - {getCourseName(match.courseId)}
                                </option>
                            ))}
                        </select>
                    )}
                </div>

                {selectedMatch && (
                    <form onSubmit={handleSubmit}>
                        <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)', overflow: 'auto' }}>
                            <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Scorecard</h3>

                            <div style={{ fontSize: '0.875rem', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)' }}>
                                <strong>{getCourseName(selectedMatch.courseId)}</strong> • Par {course?.par || 36}
                            </div>

                            {/* Scorecard Table */}
                            <div style={{ overflowX: 'auto' }}>
                                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.875rem' }}>
                                    <thead>
                                        <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
                                            <th style={{ padding: '0.75rem', textAlign: 'left', color: 'var(--color-text-secondary)', fontWeight: '600' }}>Hole</th>
                                            {[...Array(9)].map((_, i) => (
                                                <th key={i} style={{ padding: '0.75rem', textAlign: 'center', color: 'var(--color-text)', fontWeight: '600' }}>{i + 1}</th>
                                            ))}
                                            <th style={{ padding: '0.75rem', textAlign: 'center', color: 'var(--color-text)', fontWeight: '700', borderLeft: '2px solid var(--color-border)' }}>Total</th>
                                        </tr>
                                        <tr style={{ borderBottom: '1px solid var(--color-border)', backgroundColor: 'rgba(255,255,255,0.02)' }}>
                                            <td style={{ padding: '0.75rem', color: 'var(--color-text-secondary)' }}>Par</td>
                                            {(course?.holePars || Array(9).fill(4)).map((par, i) => (
                                                <td key={i} style={{ padding: '0.75rem', textAlign: 'center', color: 'var(--color-text-secondary)' }}>{par}</td>
                                            ))}
                                            <td style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '600', borderLeft: '2px solid var(--color-border)', color: 'var(--color-text-secondary)' }}>
                                                {course?.par || 36}
                                            </td>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {/* Player A */}
                                        <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                                            <td style={{ padding: '0.75rem' }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)' }}>
                                                    <strong style={{ color: 'var(--color-text)' }}>{getPlayerName(selectedMatch.playerAId)}</strong>
                                                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>
                                                        <input
                                                            type="checkbox"
                                                            checked={playerAAbsent}
                                                            onChange={(e) => setPlayerAAbsent(e.target.checked)}
                                                        />
                                                        Absent
                                                    </label>
                                                </div>
                                            </td>
                                            {playerAScores.map((score, i) => (
                                                <td key={i} style={{ padding: '0.5rem', textAlign: 'center' }}>
                                                    <input
                                                        type="number"
                                                        className="form-input"
                                                        value={score}
                                                        onChange={(e) => {
                                                            const newScores = [...playerAScores]
                                                            newScores[i] = parseInt(e.target.value) || 0
                                                            setPlayerAScores(newScores)
                                                        }}
                                                        disabled={playerAAbsent}
                                                        min="0"
                                                        max="15"
                                                        style={{ width: '50px', padding: '0.5rem', textAlign: 'center', fontSize: '0.875rem' }}
                                                    />
                                                </td>
                                            ))}
                                            <td style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '700', fontSize: '1rem', borderLeft: '2px solid var(--color-border)', color: 'var(--color-primary)' }}>
                                                {playerAAbsent ? '-' : playerAScores.reduce((a, b) => a + b, 0)}
                                            </td>
                                        </tr>

                                        {/* Player B */}
                                        <tr>
                                            <td style={{ padding: '0.75rem' }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)' }}>
                                                    <strong style={{ color: 'var(--color-text)' }}>{getPlayerName(selectedMatch.playerBId)}</strong>
                                                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>
                                                        <input
                                                            type="checkbox"
                                                            checked={playerBAbsent}
                                                            onChange={(e) => setPlayerBAbsent(e.target.checked)}
                                                        />
                                                        Absent
                                                    </label>
                                                </div>
                                            </td>
                                            {playerBScores.map((score, i) => (
                                                <td key={i} style={{ padding: '0.5rem', textAlign: 'center' }}>
                                                    <input
                                                        type="number"
                                                        className="form-input"
                                                        value={score}
                                                        onChange={(e) => {
                                                            const newScores = [...playerBScores]
                                                            newScores[i] = parseInt(e.target.value) || 0
                                                            setPlayerBScores(newScores)
                                                        }}
                                                        disabled={playerBAbsent}
                                                        min="0"
                                                        max="15"
                                                        style={{ width: '50px', padding: '0.5rem', textAlign: 'center', fontSize: '0.875rem' }}
                                                    />
                                                </td>
                                            ))}
                                            <td style={{ padding: '0.75rem', textAlign: 'center', fontWeight: '700', fontSize: '1rem', borderLeft: '2px solid var(--color-border)', color: 'var(--color-primary)' }}>
                                                {playerBAbsent ? '-' : playerBScores.reduce((a, b) => a + b, 0)}
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        <div style={{ display: 'flex', gap: 'var(--spacing-md)' }}>
                            <button
                                type="submit"
                                className="btn btn-success"
                                disabled={submitting}
                            >
                                {submitting ? 'Submitting...' : 'Submit Scores & Complete Match'}
                            </button>
                            <button
                                type="button"
                                className="btn btn-secondary"
                                onClick={() => setSelectedMatch(null)}
                                disabled={submitting}
                            >
                                Cancel
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    )
}
