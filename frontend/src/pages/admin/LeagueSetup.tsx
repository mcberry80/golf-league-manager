import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../../lib/api'
import type { Season } from '../../types'

export default function LeagueSetup() {
    const { getToken } = useAuth()
    const [seasons, setSeasons] = useState<Season[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        start_date: '',
        end_date: '',
        description: '',
        active: false,
    })

    useEffect(() => {
        loadSeasons()
    }, [])

    async function loadSeasons() {
        try {
            api.setAuthTokenProvider(getToken)
            const data = await api.listSeasons()
            setSeasons(data)
        } catch (error) {
            console.error('Failed to load seasons:', error)
        } finally {
            setLoading(false)
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        try {
            await api.createSeason(formData)
            setShowForm(false)
            setFormData({ name: '', start_date: '', end_date: '', description: '', active: false })
            loadSeasons()
        } catch (error) {
            alert('Failed to create season: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function toggleActive(season: Season) {
        try {
            await api.updateSeason(season.id, { ...season, active: !season.active })
            loadSeasons()
        } catch (error) {
            alert('Failed to update season: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

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
                    <h1>League Setup</h1>
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ New Season'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Create New Season</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Season Name</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        required
                                        placeholder="Fall 2024"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Description</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={formData.description}
                                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                        placeholder="Optional description"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Start Date</label>
                                    <input
                                        type="date"
                                        className="form-input"
                                        value={formData.start_date}
                                        onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">End Date</label>
                                    <input
                                        type="date"
                                        className="form-input"
                                        value={formData.end_date}
                                        onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                                        required
                                    />
                                </div>
                            </div>
                            <div className="form-group">
                                <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)', cursor: 'pointer' }}>
                                    <input
                                        type="checkbox"
                                        className="form-checkbox"
                                        checked={formData.active}
                                        onChange={(e) => setFormData({ ...formData, active: e.target.checked })}
                                    />
                                    <span style={{ color: 'var(--color-text-secondary)' }}>Set as active season</span>
                                </label>
                            </div>
                            <button type="submit" className="btn btn-success">
                                Create Season
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Seasons</h3>
                    {seasons.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No seasons created yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Start Date</th>
                                        <th>End Date</th>
                                        <th>Description</th>
                                        <th>Status</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {seasons.map((season) => (
                                        <tr key={season.id}>
                                            <td style={{ fontWeight: '600' }}>{season.name}</td>
                                            <td>{new Date(season.start_date).toLocaleDateString()}</td>
                                            <td>{new Date(season.end_date).toLocaleDateString()}</td>
                                            <td>{season.description}</td>
                                            <td>
                                                <span className={`badge ${season.active ? 'badge-success' : 'badge-secondary'}`}>
                                                    {season.active ? 'Active' : 'Inactive'}
                                                </span>
                                            </td>
                                            <td>
                                                <button
                                                    onClick={() => toggleActive(season)}
                                                    className="btn btn-secondary"
                                                    style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}
                                                >
                                                    {season.active ? 'Deactivate' : 'Activate'}
                                                </button>
                                            </td>
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
