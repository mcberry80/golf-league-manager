import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { MatchDay, Season, LeagueMemberWithPlayer, Course, Match } from '../../types'

export default function MatchScheduling() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()

    const [matchDays, setMatchDays] = useState<MatchDay[]>([])
    const [matches, setMatches] = useState<Match[]>([])
    const [seasons, setSeasons] = useState<Season[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])

    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)

    // Edit mode state
    const [editingMatchDay, setEditingMatchDay] = useState<MatchDay | null>(null)
    const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null)

    // Form state
    const [formData, setFormData] = useState({
        seasonId: '',
        courseId: '',
        date: '',
    })

    const [matchups, setMatchups] = useState<{ id?: string; playerAId: string; playerBId: string }[]>([
        { playerAId: '', playerBId: '' }
    ])

    const loadData = useCallback(async () => {
        if (!currentLeague) return

        try {
            const [seasonsData, membersData, coursesData, matchDaysData, matchesData] = await Promise.all([
                api.listSeasons(currentLeague.id),
                api.listLeagueMembers(currentLeague.id),
                api.listCourses(currentLeague.id),
                api.listMatchDays(currentLeague.id),
                api.listMatches(currentLeague.id),
            ])
            setSeasons(seasonsData)
            setMembers(membersData)
            setCourses(coursesData)
            setMatchDays(matchDaysData)
            setMatches(matchesData)

            // Set active season as default
            const activeSeason = seasonsData.find(s => s.active)
            if (activeSeason) {
                setFormData(prev => ({ ...prev, seasonId: activeSeason.id }))
            }
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

    const handleAddMatchup = () => {
        setMatchups([...matchups, { playerAId: '', playerBId: '' }])
    }

    const handleRemoveMatchup = (index: number) => {
        setMatchups(matchups.filter((_, i) => i !== index))
    }

    const handleMatchupChange = (index: number, field: 'playerAId' | 'playerBId', value: string) => {
        const newMatchups = [...matchups]
        newMatchups[index][field] = value
        setMatchups(newMatchups)
    }

    const resetForm = () => {
        setShowForm(false)
        setEditingMatchDay(null)
        const activeSeason = seasons.find(s => s.active)
        setFormData({
            seasonId: activeSeason?.id || '',
            courseId: '',
            date: '',
        })
        setMatchups([{ playerAId: '', playerBId: '' }])
    }

    const handleEdit = (day: MatchDay) => {
        // Can't edit completed or locked match days
        if (day.status === 'completed' || day.status === 'locked') {
            alert('Cannot edit a completed or locked match day')
            return
        }

        setEditingMatchDay(day)
        // Convert UTC date to YYYY-MM-DD format for the input
        const date = new Date(day.date)
        const dateStr = `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, '0')}-${String(date.getUTCDate()).padStart(2, '0')}`

        setFormData({
            seasonId: day.seasonId,
            courseId: day.courseId,
            date: dateStr,
        })

        // Load existing matchups
        const dayMatches = matches.filter(m => m.matchDayId === day.id)
        if (dayMatches.length > 0) {
            setMatchups(dayMatches.map(m => ({
                id: m.id,
                playerAId: m.playerAId,
                playerBId: m.playerBId,
            })))
        } else {
            setMatchups([{ playerAId: '', playerBId: '' }])
        }

        setShowForm(true)
    }

    const handleDelete = async (day: MatchDay) => {
        if (!currentLeague) return

        // Can't delete completed or locked match days
        if (day.status === 'completed' || day.status === 'locked') {
            alert('Cannot delete a completed or locked match day')
            return
        }

        try {
            await api.deleteMatchDay(currentLeague.id, day.id)
            setShowDeleteConfirm(null)
            loadData()
        } catch (error) {
            alert('Failed to delete match day: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        // Validate matchups
        const validMatchups = matchups.filter(m => m.playerAId && m.playerBId)
        if (validMatchups.length === 0) {
            alert('Please add at least one valid matchup')
            return
        }

        try {
            if (editingMatchDay) {
                // Update existing match day
                await api.updateMatchDay(currentLeague.id, editingMatchDay.id, {
                    date: formData.date,
                    courseId: formData.courseId,
                })

                // Update matchups
                await api.updateMatchDayMatchups(currentLeague.id, editingMatchDay.id, validMatchups.map(m => ({
                    id: m.id,
                    playerAId: m.playerAId,
                    playerBId: m.playerBId,
                })))
            } else {
                // Create new match day
                await api.createMatchDay(currentLeague.id, {
                    seasonId: formData.seasonId,
                    courseId: formData.courseId,
                    date: formData.date,
                    matches: validMatchups.map((m) => ({
                        playerAId: m.playerAId,
                        playerBId: m.playerBId,
                        weekNumber: 1, // Default, maybe should be calculated or input?
                    }))
                })
            }

            resetForm()
            loadData()
        } catch (error) {
            alert('Failed to save match day: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    const getPlayerName = (id: string) => {
        const member = members.find(m => m.playerId === id)
        return member?.player?.name || 'Unknown'
    }
    const getCourseName = (id: string) => courses.find(c => c.id === id)?.name || 'Unknown'
    const getSeasonName = (id: string) => seasons.find(s => s.id === id)?.name || 'Unknown'

    // Format date to display only the date part without timezone conversion
    const formatDateOnly = (dateString: string) => {
        // Parse the date string and extract just the date part (YYYY-MM-DD)
        const date = new Date(dateString)
        const year = date.getUTCFullYear()
        const month = date.getUTCMonth() + 1
        const day = date.getUTCDate()
        return `${month}/${day}/${year}`
    }

    // Check if match day can be modified
    const canModify = (day: MatchDay) => {
        return day.status !== 'completed' && day.status !== 'locked'
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

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to={`/leagues/${currentLeague.id}/admin`} style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Admin
                </Link>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <div>
                        <h1>Match Scheduling</h1>
                        <p className="text-gray-400 mt-1">{currentLeague.name}</p>
                    </div>
                    <button onClick={() => { resetForm(); setShowForm(!showForm) }} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ Create Match Day'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>
                            {editingMatchDay ? 'Edit Match Day' : 'Create Match Day'}
                        </h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-lg)', marginBottom: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Season</label>
                                    <select
                                        className="form-select"
                                        value={formData.seasonId}
                                        onChange={(e) => setFormData({ ...formData, seasonId: e.target.value })}
                                        required
                                        disabled={!!editingMatchDay}
                                    >
                                        <option value="">Select Season</option>
                                        {seasons.map(season => (
                                            <option key={season.id} value={season.id}>
                                                {season.name} {season.active && '(Active)'}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Course</label>
                                    <select
                                        className="form-select"
                                        value={formData.courseId}
                                        onChange={(e) => setFormData({ ...formData, courseId: e.target.value })}
                                        required
                                    >
                                        <option value="">Select Course</option>
                                        {courses.map(course => (
                                            <option key={course.id} value={course.id}>{course.name}</option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Date</label>
                                    <input
                                        type="date"
                                        className="form-input"
                                        value={formData.date}
                                        onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                                        required
                                    />
                                </div>
                            </div>

                            <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text)' }}>Matchups</h4>
                            {matchups.map((matchup, index) => (
                                <div key={matchup.id || index} className="grid grid-cols-2" style={{ gap: 'var(--spacing-md)', marginBottom: 'var(--spacing-md)', alignItems: 'end' }}>
                                    <div className="form-group">
                                        <label className="form-label">Player A</label>
                                        <select
                                            className="form-select"
                                            value={matchup.playerAId}
                                            onChange={(e) => handleMatchupChange(index, 'playerAId', e.target.value)}
                                            required
                                        >
                                            <option value="">Select Player</option>
                                            {members.map(member => (
                                                <option key={member.id} value={member.playerId} disabled={matchups.some((m, i) => i !== index && (m.playerAId === member.playerId || m.playerBId === member.playerId)) || matchup.playerBId === member.playerId}>
                                                    {member.player?.name || member.player?.email}
                                                </option>
                                            ))}
                                        </select>
                                    </div>
                                    <div style={{ display: 'flex', gap: 'var(--spacing-sm)' }}>
                                        <div className="form-group" style={{ flex: 1 }}>
                                            <label className="form-label">Player B</label>
                                            <select
                                                className="form-select"
                                                value={matchup.playerBId}
                                                onChange={(e) => handleMatchupChange(index, 'playerBId', e.target.value)}
                                                required
                                            >
                                                <option value="">Select Player</option>
                                                {members.map(member => (
                                                    <option key={member.id} value={member.playerId} disabled={matchups.some((m, i) => i !== index && (m.playerAId === member.playerId || m.playerBId === member.playerId)) || matchup.playerAId === member.playerId}>
                                                        {member.player?.name || member.player?.email}
                                                    </option>
                                                ))}
                                            </select>
                                        </div>
                                        <button type="button" className="btn btn-danger" onClick={() => handleRemoveMatchup(index)} disabled={matchups.length === 1} style={{ padding: '0.5rem 0.75rem', height: 'auto', fontSize: '0.875rem', alignSelf: 'flex-end' }}>
                                            X
                                        </button>
                                    </div>
                                </div>
                            ))}

                            <button type="button" className="btn btn-secondary" onClick={handleAddMatchup} style={{ marginBottom: 'var(--spacing-lg)' }}>
                                + Add Matchup
                            </button>

                            <div style={{ borderTop: '1px solid var(--color-border)', paddingTop: 'var(--spacing-lg)', display: 'flex', gap: 'var(--spacing-md)' }}>
                                <button type="submit" className="btn btn-success">
                                    {editingMatchDay ? 'Update Match Day' : 'Save Match Day'}
                                </button>
                                {editingMatchDay && (
                                    <button type="button" className="btn btn-secondary" onClick={resetForm}>
                                        Cancel
                                    </button>
                                )}
                            </div>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Scheduled Match Days</h3>
                    {matchDays.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No match days scheduled yet.</p>
                    ) : (
                        <div>
                            {matchDays.map(day => (
                                <div key={day.id} style={{ marginBottom: 'var(--spacing-xl)', borderBottom: '1px solid var(--color-border)', paddingBottom: 'var(--spacing-lg)' }}>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-md)' }}>
                                        <h4 style={{ color: 'var(--color-text)' }}>
                                            {formatDateOnly(day.date)} @ {getCourseName(day.courseId)}
                                        </h4>
                                        <div style={{ display: 'flex', gap: 'var(--spacing-sm)', alignItems: 'center' }}>
                                            <span className={`badge ${day.status === 'completed' ? 'badge-success' : day.status === 'locked' ? 'badge-warning' : 'badge-info'}`}>
                                                {day.status}
                                            </span>
                                            <span className="badge badge-info">{getSeasonName(day.seasonId)}</span>
                                            {canModify(day) && (
                                                <>
                                                    <button
                                                        className="btn btn-secondary"
                                                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                        onClick={() => handleEdit(day)}
                                                    >
                                                        Edit
                                                    </button>
                                                    <button
                                                        className="btn btn-danger"
                                                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                        onClick={() => setShowDeleteConfirm(day.id)}
                                                    >
                                                        Delete
                                                    </button>
                                                </>
                                            )}
                                        </div>
                                    </div>

                                    {/* Delete confirmation */}
                                    {showDeleteConfirm === day.id && (
                                        <div style={{ 
                                            backgroundColor: 'rgba(239, 68, 68, 0.1)', 
                                            border: '1px solid var(--color-error)', 
                                            borderRadius: 'var(--radius-md)', 
                                            padding: 'var(--spacing-md)', 
                                            marginBottom: 'var(--spacing-md)' 
                                        }}>
                                            <p style={{ color: 'var(--color-error)', marginBottom: 'var(--spacing-sm)' }}>
                                                Are you sure you want to delete this match day? This will also delete all associated matches.
                                            </p>
                                            <div style={{ display: 'flex', gap: 'var(--spacing-sm)' }}>
                                                <button
                                                    className="btn btn-danger"
                                                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    onClick={() => handleDelete(day)}
                                                >
                                                    Yes, Delete
                                                </button>
                                                <button
                                                    className="btn btn-secondary"
                                                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    onClick={() => setShowDeleteConfirm(null)}
                                                >
                                                    Cancel
                                                </button>
                                            </div>
                                        </div>
                                    )}

                                    <div className="table-container">
                                        <table className="table">
                                            <thead>
                                                <tr>
                                                    <th>Matchup</th>
                                                    <th>Status</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {matches.filter(m => m.matchDayId === day.id).map(match => (
                                                    <tr key={match.id}>
                                                        <td style={{ fontWeight: '600' }}>
                                                            {getPlayerName(match.playerAId)} vs {getPlayerName(match.playerBId)}
                                                        </td>
                                                        <td>
                                                            <span className={`badge ${match.status === 'completed' ? 'badge-success' : 'badge-warning'}`}>
                                                                {match.status}
                                                            </span>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div >
        </div >
    )
}
