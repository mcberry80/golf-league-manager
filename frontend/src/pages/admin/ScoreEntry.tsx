import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Match, LeagueMemberWithPlayer, Course, MatchDay } from '../../types'

export default function ScoreEntry() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()

    const [matchDays, setMatchDays] = useState<MatchDay[]>([])
    const [selectedMatchDay, setSelectedMatchDay] = useState<MatchDay | null>(null)
    const [matches, setMatches] = useState<Match[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)

    // Scores state: matchId_playerId -> number[]
    const [scores, setScores] = useState<Record<string, number[]>>({})
    // Absent state: matchId_playerId -> boolean
    const [absentPlayers, setAbsentPlayers] = useState<Record<string, boolean>>({})
    const [submitting, setSubmitting] = useState(false)

    const loadData = useCallback(async () => {
        if (!currentLeague) return

        try {
            const [matchDaysData, matchesData, membersData, coursesData] = await Promise.all([
                api.listMatchDays(currentLeague.id),
                api.listMatches(currentLeague.id),
                api.listLeagueMembers(currentLeague.id),
                api.listCourses(currentLeague.id),
            ])
            setMatchDays(matchDaysData)
            setMatches(matchesData)
            setMembers(membersData)
            setCourses(coursesData)
        } catch (error) {
            console.error('Failed to load data:', error)
        } finally {
            setLoading(false)
        }
    }, [currentLeague])

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
    }, [currentLeague, leagueLoading, navigate, leagueId, loadData])

    function handleMatchDaySelect(matchDayId: string) {
        const matchDay = matchDays.find(m => m.id === matchDayId)
        setSelectedMatchDay(matchDay || null)

        if (matchDay) {
            const dayMatches = matches.filter(m => m.matchDayId === matchDay.id)
            const course = courses.find(c => c.id === matchDay.courseId)
            const defaultScores = course?.holePars || Array(9).fill(4)

            const initialScores: Record<string, number[]> = {}
            const initialAbsent: Record<string, boolean> = {}
            dayMatches.forEach(match => {
                initialScores[`${match.id}_${match.playerAId}`] = [...defaultScores]
                initialScores[`${match.id}_${match.playerBId}`] = [...defaultScores]
                initialAbsent[`${match.id}_${match.playerAId}`] = false
                initialAbsent[`${match.id}_${match.playerBId}`] = false
            })
            setScores(initialScores)
            setAbsentPlayers(initialAbsent)
        }
    }

    function handleScoreChange(matchId: string, playerId: string, holeIndex: number, value: number) {
        const key = `${matchId}_${playerId}`
        const currentScores = scores[key] || Array(9).fill(0)
        const newScores = [...currentScores]
        newScores[holeIndex] = value
        setScores(prev => ({ ...prev, [key]: newScores }))
    }

    function handleAbsentChange(matchId: string, playerId: string, isAbsent: boolean) {
        const key = `${matchId}_${playerId}`
        setAbsentPlayers(prev => ({ ...prev, [key]: isAbsent }))
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!selectedMatchDay || !currentLeague) return

        setSubmitting(true)
        try {
            const dayMatches = matches.filter(m => m.matchDayId === selectedMatchDay.id)
            const scoreSubmissions = []

            for (const match of dayMatches) {
                const keyA = `${match.id}_${match.playerAId}`
                const keyB = `${match.id}_${match.playerBId}`

                // Player A
                scoreSubmissions.push({
                    matchId: match.id,
                    playerId: match.playerAId,
                    holeScores: scores[keyA],
                    playerAbsent: absentPlayers[keyA] || false
                })
                // Player B
                scoreSubmissions.push({
                    matchId: match.id,
                    playerId: match.playerBId,
                    holeScores: scores[keyB],
                    playerAbsent: absentPlayers[keyB] || false
                })
            }

            await api.enterMatchDayScores(currentLeague.id, scoreSubmissions)

            alert('Scores entered successfully! Matches completed and handicaps updated.')
            setSelectedMatchDay(null)
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

    // Format date to display only the date part without timezone conversion
    const formatDateOnly = (dateString: string) => {
        const date = new Date(dateString)
        const year = date.getUTCFullYear()
        const month = date.getUTCMonth() + 1
        const day = date.getUTCDate()
        return `${month}/${day}/${year}`
    }

    // Render player row with absent checkbox and score inputs
    const renderPlayerRow = (matchId: string, playerId: string) => {
        const key = `${matchId}_${playerId}`
        const isAbsent = absentPlayers[key] || false
        const playerScores = scores[key] || Array(9).fill(0)

        return (
            <tr key={key}>
                <td style={{ padding: '0.5rem', fontWeight: '600' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <label 
                            style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '0.25rem',
                                cursor: 'pointer',
                                fontSize: '0.75rem',
                                color: isAbsent ? 'var(--color-warning)' : 'var(--color-text-muted)'
                            }}
                            title="Mark player as absent"
                        >
                            <input
                                type="checkbox"
                                checked={isAbsent}
                                onChange={(e) => handleAbsentChange(matchId, playerId, e.target.checked)}
                                style={{ 
                                    cursor: 'pointer',
                                    accentColor: 'var(--color-warning)'
                                }}
                            />
                            <span style={{ fontSize: '0.65rem' }}>Absent</span>
                        </label>
                        <span style={{ 
                            textDecoration: isAbsent ? 'line-through' : 'none',
                            opacity: isAbsent ? 0.7 : 1
                        }}>
                            {getPlayerName(playerId)}
                        </span>
                        {isAbsent && (
                            <span 
                                style={{ 
                                    fontSize: '0.65rem', 
                                    backgroundColor: 'var(--color-warning)',
                                    color: '#000',
                                    padding: '0.1rem 0.3rem',
                                    borderRadius: '3px',
                                    fontWeight: '500'
                                }}
                            >
                                ABSENT
                            </span>
                        )}
                    </div>
                </td>
                {playerScores.map((score, i) => (
                    <td key={i} style={{ padding: '0.25rem', textAlign: 'center' }}>
                        <input
                            type="number"
                            className="form-input"
                            value={score}
                            onChange={(e) => handleScoreChange(matchId, playerId, i, parseInt(e.target.value) || 0)}
                            min="0"
                            max="15"
                            disabled={isAbsent}
                            style={{ 
                                width: '40px', 
                                padding: '0.25rem', 
                                textAlign: 'center',
                                opacity: isAbsent ? 0.5 : 1,
                                backgroundColor: isAbsent ? 'var(--color-bg-tertiary)' : undefined
                            }}
                            title={isAbsent ? 'Score will be calculated automatically for absent players' : ''}
                        />
                    </td>
                ))}
                <td style={{ 
                    padding: '0.5rem', 
                    textAlign: 'center', 
                    fontWeight: '700', 
                    color: isAbsent ? 'var(--color-warning)' : 'var(--color-primary)'
                }}>
                    {isAbsent ? (
                        <span title="Score will be calculated based on playing handicap + par + 3">Auto</span>
                    ) : (
                        playerScores.reduce((a, b) => a + b, 0)
                    )}
                </td>
            </tr>
        )
    }

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

    const course = selectedMatchDay ? getCourse(selectedMatchDay.courseId) : null
    const dayMatches = selectedMatchDay ? matches.filter(m => m.matchDayId === selectedMatchDay.id) : []

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
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Select Match Day</h3>
                    {matchDays.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No match days available.</p>
                    ) : (
                        <select
                            className="form-input"
                            value={selectedMatchDay?.id || ''}
                            onChange={(e) => handleMatchDaySelect(e.target.value)}
                            style={{ maxWidth: '500px' }}
                        >
                            <option value="">-- Select a Match Day --</option>
                            {matchDays.map((day) => (
                                <option key={day.id} value={day.id}>
                                    {formatDateOnly(day.date)} @ {getCourseName(day.courseId)}
                                </option>
                            ))}
                        </select>
                    )}
                </div>

                {selectedMatchDay && (
                    <form onSubmit={handleSubmit}>
                        {/* Absent player info box */}
                        <div className="card-glass" style={{ 
                            marginBottom: 'var(--spacing-lg)', 
                            padding: 'var(--spacing-md)',
                            backgroundColor: 'rgba(234, 179, 8, 0.1)',
                            borderLeft: '3px solid var(--color-warning)'
                        }}>
                            <h4 style={{ color: 'var(--color-warning)', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
                                ℹ️ Absent Player Rules
                            </h4>
                            <ul style={{ 
                                fontSize: '0.8rem', 
                                color: 'var(--color-text-muted)', 
                                marginLeft: '1rem',
                                lineHeight: '1.6'
                            }}>
                                <li>Absent player scores are calculated as: <strong>Playing Handicap + Par + 3</strong></li>
                                <li>Strokes are distributed evenly across holes, with extras on hardest holes</li>
                                <li>Absent rounds do <strong>not</strong> affect handicap calculations</li>
                                <li>Net scores for match play are calculated normally based on strokes received</li>
                            </ul>
                        </div>

                        <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)', overflow: 'auto' }}>
                            <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>
                                Scores for {formatDateOnly(selectedMatchDay.date)}
                            </h3>

                            {dayMatches.map(match => (
                                <div key={match.id} style={{ marginBottom: 'var(--spacing-xl)', borderBottom: '1px solid var(--color-border)', paddingBottom: 'var(--spacing-lg)' }}>
                                    <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text-secondary)' }}>
                                        Match: {getPlayerName(match.playerAId)} vs {getPlayerName(match.playerBId)}
                                    </h4>

                                    <div style={{ overflowX: 'auto' }}>
                                        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.875rem' }}>
                                            <thead>
                                                <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
                                                    <th style={{ padding: '0.5rem', textAlign: 'left', width: '200px' }}>Player</th>
                                                    {[...Array(9)].map((_, i) => (
                                                        <th key={i} style={{ padding: '0.5rem', textAlign: 'center' }}>{i + 1}</th>
                                                    ))}
                                                    <th style={{ padding: '0.5rem', textAlign: 'center', fontWeight: '700' }}>Total</th>
                                                </tr>
                                                <tr style={{ borderBottom: '1px solid var(--color-border)', backgroundColor: 'rgba(255,255,255,0.02)' }}>
                                                    <td style={{ padding: '0.5rem', color: 'var(--color-text-secondary)' }}>Par</td>
                                                    {(course?.holePars || Array(9).fill(4)).map((par, i) => (
                                                        <td key={i} style={{ padding: '0.5rem', textAlign: 'center', color: 'var(--color-text-secondary)' }}>{par}</td>
                                                    ))}
                                                    <td style={{ padding: '0.5rem', textAlign: 'center', color: 'var(--color-text-secondary)' }}>{course?.par || 36}</td>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {renderPlayerRow(match.id, match.playerAId)}
                                                {renderPlayerRow(match.id, match.playerBId)}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            ))}
                        </div>

                        <div style={{ display: 'flex', gap: 'var(--spacing-md)' }}>
                            <button
                                type="submit"
                                className="btn btn-success"
                                disabled={submitting}
                            >
                                {submitting ? 'Submitting...' : 'Save All Scores'}
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    )
}
