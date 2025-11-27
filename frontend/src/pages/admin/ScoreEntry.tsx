import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Match, LeagueMemberWithPlayer, Course, MatchDay, ScoreResponse } from '../../types'
import { LoadingSpinner, AccessDenied } from '../../components/Layout'

type MessageType = 'success' | 'error' | 'warning' | 'info'

interface Message {
    type: MessageType
    text: string
    details?: string[]
}

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
    const [loadingScores, setLoadingScores] = useState(false)

    // Scores state: matchId_playerId -> number[]
    const [scores, setScores] = useState<Record<string, number[]>>({})
    // Absent state: matchId_playerId -> boolean
    const [absentPlayers, setAbsentPlayers] = useState<Record<string, boolean>>({})
    const [submitting, setSubmitting] = useState(false)
    const [hasExistingScores, setHasExistingScores] = useState(false)
    
    // Message state
    const [message, setMessage] = useState<Message | null>(null)

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
            setMessage({
                type: 'error',
                text: 'Failed to load data',
                details: [error instanceof Error ? error.message : 'Unknown error']
            })
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

    async function handleMatchDaySelect(matchDayId: string) {
        setMessage(null)
        
        if (!matchDayId) {
            setSelectedMatchDay(null)
            setScores({})
            setAbsentPlayers({})
            setHasExistingScores(false)
            return
        }

        const matchDay = matchDays.find(m => m.id === matchDayId)
        setSelectedMatchDay(matchDay || null)

        if (matchDay && currentLeague) {
            const dayMatches = matches.filter(m => m.matchDayId === matchDay.id)
            const course = courses.find(c => c.id === matchDay.courseId)
            const defaultScores = course?.holePars || Array(9).fill(4)

            // Initialize with default scores
            const initialScores: Record<string, number[]> = {}
            const initialAbsent: Record<string, boolean> = {}
            dayMatches.forEach(match => {
                initialScores[`${match.id}_${match.playerAId}`] = [...defaultScores]
                initialScores[`${match.id}_${match.playerBId}`] = [...defaultScores]
                initialAbsent[`${match.id}_${match.playerAId}`] = false
                initialAbsent[`${match.id}_${match.playerBId}`] = false
            })

            // Try to load existing scores
            if (matchDay.hasScores || matchDay.status === 'completed') {
                setLoadingScores(true)
                try {
                    const response = await api.getMatchDayScores(currentLeague.id, matchDay.id)
                    if (response.scores && response.scores.length > 0) {
                        setHasExistingScores(true)
                        // Populate with existing scores
                        response.scores.forEach((score: ScoreResponse) => {
                            const key = `${score.matchId}_${score.playerId}`
                            if (score.holeScores && score.holeScores.length > 0) {
                                initialScores[key] = score.holeScores
                            }
                            initialAbsent[key] = score.playerAbsent || false
                        })
                        
                        if (matchDay.status === 'completed') {
                            setMessage({
                                type: 'info',
                                text: 'This match week has been completed. You can update scores if needed.'
                            })
                        }
                    }
                } catch (error) {
                    console.error('Failed to load existing scores:', error)
                    // Continue with default scores
                }
                setLoadingScores(false)
            } else {
                setHasExistingScores(false)
            }

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

        setMessage(null)
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

            const response = await api.enterMatchDayScores(currentLeague.id, selectedMatchDay.id, scoreSubmissions)

            if (response.status === 'success') {
                const actionText = response.updated ? 'updated' : 'saved'
                setMessage({
                    type: 'success',
                    text: `Scores ${actionText} successfully!`,
                    details: response.warnings && response.warnings.length > 0 
                        ? [`${response.count} scores processed`, ...response.warnings]
                        : [`${response.count} scores processed. Matches completed and handicaps updated.`]
                })
                setHasExistingScores(true)
                // Reload match days to get updated status
                loadData()
            } else {
                setMessage({
                    type: 'error',
                    text: response.message || 'Failed to save scores',
                    details: response.warnings
                })
            }
        } catch (error) {
            setMessage({
                type: 'error',
                text: 'Failed to save scores',
                details: [error instanceof Error ? error.message : 'Unknown error']
            })
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

    // Get status badge styling
    const getStatusBadge = (status: string | undefined, hasScores: boolean | undefined) => {
        if (status === 'locked') {
            return { text: 'Locked', color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.2)' }
        }
        if (status === 'completed' || hasScores) {
            return { text: 'Completed', color: 'var(--color-accent)', bgColor: 'rgba(16, 185, 129, 0.2)' }
        }
        return { text: 'Scheduled', color: 'var(--color-primary)', bgColor: 'rgba(59, 130, 246, 0.2)' }
    }

    // Render player row with absent checkbox and score inputs
    const renderPlayerRow = (matchId: string, playerId: string, isLocked: boolean) => {
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
                                cursor: isLocked ? 'not-allowed' : 'pointer',
                                fontSize: '0.75rem',
                                color: isAbsent ? 'var(--color-warning)' : 'var(--color-text-muted)',
                                opacity: isLocked ? 0.6 : 1
                            }}
                            title={isLocked ? 'This match week is locked' : 'Mark player as absent'}
                        >
                            <input
                                type="checkbox"
                                checked={isAbsent}
                                onChange={(e) => handleAbsentChange(matchId, playerId, e.target.checked)}
                                disabled={isLocked}
                                style={{ 
                                    cursor: isLocked ? 'not-allowed' : 'pointer',
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
                            disabled={isAbsent || isLocked}
                            style={{ 
                                width: '40px', 
                                padding: '0.25rem', 
                                textAlign: 'center',
                                opacity: (isAbsent || isLocked) ? 0.5 : 1,
                                backgroundColor: (isAbsent || isLocked) ? 'var(--color-bg-tertiary)' : undefined,
                                cursor: isLocked ? 'not-allowed' : undefined
                            }}
                            title={isLocked ? 'This match week is locked' : (isAbsent ? 'Score will be calculated automatically for absent players' : '')}
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

    // Message component
    const renderMessage = () => {
        if (!message) return null

        const alertClass = `alert alert-${message.type}`
        
        return (
            <div className={alertClass} style={{ marginBottom: 'var(--spacing-lg)' }}>
                <strong>{message.text}</strong>
                {message.details && message.details.length > 0 && (
                    <ul style={{ marginTop: '0.5rem', marginBottom: 0, paddingLeft: '1.5rem' }}>
                        {message.details.map((detail, i) => (
                            <li key={i} style={{ fontSize: '0.9rem' }}>{detail}</li>
                        ))}
                    </ul>
                )}
            </div>
        )
    }

    if (leagueLoading || loading) {
        return <LoadingSpinner />
    }

    if (!currentLeague || userRole !== 'admin') {
        return <AccessDenied leagueName={currentLeague?.name} />
    }

    const course = selectedMatchDay ? getCourse(selectedMatchDay.courseId) : null
    const dayMatches = selectedMatchDay ? matches.filter(m => m.matchDayId === selectedMatchDay.id) : []
    const isLocked = selectedMatchDay?.status === 'locked'

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

                {renderMessage()}

                <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Select Match Week</h3>
                    {matchDays.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No match weeks available. Create a match day first.</p>
                    ) : (
                        <>
                            <select
                                className="form-input"
                                value={selectedMatchDay?.id || ''}
                                onChange={(e) => handleMatchDaySelect(e.target.value)}
                                style={{ maxWidth: '500px' }}
                            >
                                <option value="">-- Select a Match Week --</option>
                                {matchDays.map((day) => {
                                    const badge = getStatusBadge(day.status, day.hasScores)
                                    return (
                                        <option key={day.id} value={day.id}>
                                            {day.weekNumber ? `Week ${day.weekNumber}: ` : ''}{formatDateOnly(day.date)} @ {getCourseName(day.courseId)} [{badge.text}]
                                        </option>
                                    )
                                })}
                            </select>
                            
                            {/* Match day status legend */}
                            <div style={{ marginTop: 'var(--spacing-md)', display: 'flex', gap: 'var(--spacing-lg)', flexWrap: 'wrap' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}>
                                    <span style={{ 
                                        display: 'inline-block',
                                        width: '12px',
                                        height: '12px',
                                        borderRadius: '50%',
                                        backgroundColor: 'var(--color-primary)'
                                    }}></span>
                                    <span style={{ color: 'var(--color-text-muted)' }}>Scheduled - Ready for score entry</span>
                                </div>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}>
                                    <span style={{ 
                                        display: 'inline-block',
                                        width: '12px',
                                        height: '12px',
                                        borderRadius: '50%',
                                        backgroundColor: 'var(--color-accent)'
                                    }}></span>
                                    <span style={{ color: 'var(--color-text-muted)' }}>Completed - Scores can be updated</span>
                                </div>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}>
                                    <span style={{ 
                                        display: 'inline-block',
                                        width: '12px',
                                        height: '12px',
                                        borderRadius: '50%',
                                        backgroundColor: 'var(--color-danger)'
                                    }}></span>
                                    <span style={{ color: 'var(--color-text-muted)' }}>Locked - View only</span>
                                </div>
                            </div>
                        </>
                    )}
                </div>

                {loadingScores && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)', marginBottom: 'var(--spacing-lg)' }}>
                        <div className="spinner" style={{ width: '24px', height: '24px' }}></div>
                        <span style={{ color: 'var(--color-text-muted)' }}>Loading existing scores...</span>
                    </div>
                )}

                {selectedMatchDay && !loadingScores && (
                    <form onSubmit={handleSubmit}>
                        {/* Locked match day warning */}
                        {isLocked && (
                            <div className="alert alert-error" style={{ marginBottom: 'var(--spacing-lg)' }}>
                                <strong>🔒 This match week is locked.</strong>
                                <p style={{ marginTop: '0.5rem', marginBottom: 0 }}>
                                    Scores for locked match weeks cannot be modified. A match week becomes locked when 
                                    scores are entered for a later match week in the same season.
                                </p>
                            </div>
                        )}

                        {/* Update notice for completed match days */}
                        {hasExistingScores && !isLocked && (
                            <div className="alert alert-info" style={{ marginBottom: 'var(--spacing-lg)' }}>
                                <strong>ℹ️ Updating Existing Scores</strong>
                                <p style={{ marginTop: '0.5rem', marginBottom: 0 }}>
                                    This match week already has scores entered. Saving will update the existing scores 
                                    and recalculate handicaps.
                                </p>
                            </div>
                        )}

                        {/* Absent player info box */}
                        {!isLocked && (
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
                        )}

                        <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)', overflow: 'auto' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-lg)' }}>
                                <h3 style={{ color: 'var(--color-text)', margin: 0 }}>
                                    {selectedMatchDay.weekNumber ? `Week ${selectedMatchDay.weekNumber}: ` : ''}
                                    {formatDateOnly(selectedMatchDay.date)}
                                </h3>
                                {(() => {
                                    const badge = getStatusBadge(selectedMatchDay.status, selectedMatchDay.hasScores)
                                    return (
                                        <span style={{
                                            padding: '0.25rem 0.75rem',
                                            borderRadius: 'var(--radius-full)',
                                            fontSize: '0.75rem',
                                            fontWeight: '600',
                                            backgroundColor: badge.bgColor,
                                            color: badge.color,
                                            textTransform: 'uppercase'
                                        }}>
                                            {badge.text}
                                        </span>
                                    )
                                })()}
                            </div>

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
                                                {renderPlayerRow(match.id, match.playerAId, isLocked)}
                                                {renderPlayerRow(match.id, match.playerBId, isLocked)}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            ))}
                        </div>

                        {!isLocked && (
                            <div style={{ display: 'flex', gap: 'var(--spacing-md)', alignItems: 'center' }}>
                                <button
                                    type="submit"
                                    className="btn btn-success"
                                    disabled={submitting}
                                >
                                    {submitting ? 'Saving...' : (hasExistingScores ? 'Update Scores' : 'Save All Scores')}
                                </button>
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    onClick={() => {
                                        setSelectedMatchDay(null)
                                        setMessage(null)
                                    }}
                                    disabled={submitting}
                                >
                                    Cancel
                                </button>
                            </div>
                        )}
                    </form>
                )}
            </div>
        </div>
    )
}
