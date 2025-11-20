import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../../lib/api'
import type { Match, Season, Player, Course } from '../../types'

export default function MatchScheduling() {
    const { getToken } = useAuth()
    const [matches, setMatches] = useState<Match[]>([])
    const [seasons, setSeasons] = useState<Season[]>([])
    const [players, setPlayers] = useState<Player[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        season_id: '',
        week_number: 1,
        player_a_id: '',
        player_b_id: '',
        course_id: '',
        match_date: '',
    })

    useEffect(() => {
        loadData()
    }, [])

    async function loadData() {
        try {
            api.setAuthTokenProvider(getToken)
            const [seasonsData, playersData, coursesData, matchesData] = await Promise.all([
                api.listSeasons(),
                api.listPlayers(true),
                api.listCourses(),
                api.listMatches(),
            ])
            setSeasons(seasonsData)
            setPlayers(playersData)
            setCourses(coursesData)
            setMatches(matchesData)

            // Set active season as default
            const activeSeason = seasonsData.find(s => s.active)
            if (activeSeason) {
                setFormData(prev => ({ ...prev, season_id: activeSeason.id }))
            }
        } catch (error) {
            console.error('Failed to load data:', error)
        } finally {
            setLoading(false)
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        try {
            await api.createMatch(formData)
            setShowForm(false)
            setFormData({
                season_id: formData.season_id,
                week_number: formData.week_number + 1,
                player_a_id: '',
                player_b_id: '',
                course_id: '',
                match_date: '',
            })
            loadData()
        } catch (error) {
            alert('Failed to create match: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    const getPlayerName = (id: string) => players.find(p => p.id === id)?.name || 'Unknown'
    const getCourseName = (id: string) => courses.find(c => c.id === id)?.name || 'Unknown'
    const getSeasonName = (id: string) => seasons.find(s => s.id === id)?.name || 'Unknown'

    // Group matches by week
    const matchesByWeek = matches.reduce((acc, match) => {
        if (!acc[match.week_number]) acc[match.week_number] = []
        acc[match.week_number].push(match)
        return acc
    }, {} as Record<number, Match[]>)

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

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <h1>Match Scheduling</h1>
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
                                        value={formData.season_id}
                                        onChange={(e) => setFormData({ ...formData, season_id: e.target.value })}
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
                                        value={formData.week_number}
                                        onChange={(e) => setFormData({ ...formData, week_number: parseInt(e.target.value) })}
                                        required
                                        min="1"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Player A</label>
                                    <select
                                        className="form-select"
                                        value={formData.player_a_id}
                                        onChange={(e) => setFormData({ ...formData, player_a_id: e.target.value })}
                                        required
                                    >
                                        <option value="">Select Player</option>
                                        {players.map(player => (
                                            <option key={player.id} value={player.id}>{player.name}</option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Player B</label>
                                    <select
                                        className="form-select"
                                        value={formData.player_b_id}
                                        onChange={(e) => setFormData({ ...formData, player_b_id: e.target.value })}
                                        required
                                    >
                                        <option value="">Select Player</option>
                                        {players.map(player => (
                                            <option key={player.id} value={player.id}>{player.name}</option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Course</label>
                                    <select
                                        className="form-select"
                                        value={formData.course_id}
                                        onChange={(e) => setFormData({ ...formData, course_id: e.target.value })}
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
                                        value={formData.match_date}
                                        onChange={(e) => setFormData({ ...formData, match_date: e.target.value })}
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
                                                        <td>{new Date(match.match_date).toLocaleDateString()}</td>
                                                        <td style={{ fontWeight: '600' }}>
                                                            {getPlayerName(match.player_a_id)} vs {getPlayerName(match.player_b_id)}
                                                        </td>
                                                        <td>{getCourseName(match.course_id)}</td>
                                                        <td>{getSeasonName(match.season_id)}</td>
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
