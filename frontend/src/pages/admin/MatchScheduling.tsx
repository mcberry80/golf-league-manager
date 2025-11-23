import { useState, useEffect } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Match, Season, LeagueMemberWithPlayer, Course } from '../../types'

export default function MatchScheduling() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [matches, setMatches] = useState<Match[]>([])
    const [seasons, setSeasons] = useState<Season[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        seasonId: '',
        weekNumber: 1,
        playerAId: '',
        playerBId: '',
        courseId: '',
        matchDate: '',
    })

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
            const [seasonsData, membersData, coursesData, matchesData] = await Promise.all([
                api.listSeasons(currentLeague.id),
                api.listLeagueMembers(currentLeague.id),
                api.listCourses(currentLeague.id),
                api.listMatches(currentLeague.id),
            ])
            setSeasons(seasonsData)
            setMembers(membersData)
            setCourses(coursesData)
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
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.createMatch(currentLeague.id, formData)
            setShowForm(false)
            setFormData({
                seasonId: formData.seasonId,
                weekNumber: formData.weekNumber + 1,
                playerAId: '',
                playerBId: '',
                courseId: '',
                matchDate: '',
            })
            loadData()
        } catch (error) {
            alert('Failed to create match: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    const getPlayerName = (id: string) => {
        // Match stores player_id, so we look up in members list
        // members have player_id and player object
        const member = members.find(m => m.playerId === id)
        return member?.player?.name || 'Unknown'
    }
    const getCourseName = (id: string) => courses.find(c => c.id === id)?.name || 'Unknown'
    const getSeasonName = (id: string) => seasons.find(s => s.id === id)?.name || 'Unknown'

    // Group matches by week
    const matchesByWeek = matches.reduce((acc, match) => {
        if (!acc[match.weekNumber]) acc[match.weekNumber] = []
        acc[match.weekNumber].push(match)
        return acc
    }, {} as Record<number, Match[]>)

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
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ Schedule Match'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Schedule New Match</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Season</label>
                                    <select
                                        className="form-select"
                                        value={formData.seasonId}
                                        onChange={(e) => setFormData({ ...formData, seasonId: e.target.value })}
                                        required
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
                                    <label className="form-label">Week Number</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={formData.weekNumber}
                                        onChange={(e) => setFormData({ ...formData, weekNumber: parseInt(e.target.value) })}
                                        required
                                        min="1"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Player A</label>
                                    <select
                                        className="form-select"
                                        value={formData.playerAId}
                                        onChange={(e) => setFormData({ ...formData, playerAId: e.target.value })}
                                        required
                                    >
                                        <option value="">Select Player</option>
                                        {members.map(member => (
                                            <option key={member.id} value={member.playerId}>{member.player?.name || member.player?.email}</option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Player B</label>
                                    <select
                                        className="form-select"
                                        value={formData.playerBId}
                                        onChange={(e) => setFormData({ ...formData, playerBId: e.target.value })}
                                        required
                                    >
                                        <option value="">Select Player</option>
                                        {members.map(member => (
                                            <option key={member.id} value={member.playerId}>{member.player?.name || member.player?.email}</option>
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
                                    <label className="form-label">Match Date</label>
                                    <input
                                        type="date"
                                        className="form-input"
                                        value={formData.matchDate}
                                        onChange={(e) => setFormData({ ...formData, matchDate: e.target.value })}
                                        required
                                    />
                                </div>
                            </div>
                            <button type="submit" className="btn btn-success">
                                Schedule Match
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Scheduled Matches ({matches.length})</h3>
                    {matches.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No matches scheduled yet.</p>
                    ) : (
                        <div>
                            {Object.keys(matchesByWeek).sort((a, b) => parseInt(a) - parseInt(b)).map(week => (
                                <div key={week} style={{ marginBottom: 'var(--spacing-xl)' }}>
                                    <h4 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text-secondary)' }}>
                                        Week {week}
                                    </h4>
                                    <div className="table-container">
                                        <table className="table">
                                            <thead>
                                                <tr>
                                                    <th>Date</th>
                                                    <th>Matchup</th>
                                                    <th>Course</th>
                                                    <th>Season</th>
                                                    <th>Status</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {matchesByWeek[parseInt(week)].map(match => (
                                                    <tr key={match.id}>
                                                        <td>{new Date(match.matchDate).toLocaleDateString()}</td>
                                                        <td style={{ fontWeight: '600' }}>
                                                            {getPlayerName(match.playerAId)} vs {getPlayerName(match.playerBId)}
                                                        </td>
                                                        <td>{getCourseName(match.courseId)}</td>
                                                        <td>{getSeasonName(match.seasonId)}</td>
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
            </div>
        </div>
    )
}
