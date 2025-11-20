import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../../lib/api'
import type { Player } from '../../types'

export default function PlayerManagement() {
    const { getToken } = useAuth()
    const [players, setPlayers] = useState<Player[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        is_admin: false,
    })

    useEffect(() => {
        loadPlayers()
    }, [])

    async function loadPlayers() {
        try {
            api.setAuthTokenProvider(getToken)
            const data = await api.listPlayers()
            setPlayers(data)
        } catch (error) {
            console.error('Failed to load players:', error)
        } finally {
            setLoading(false)
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        try {
            await api.createPlayer(formData)
            setShowForm(false)
            setFormData({ name: '', email: '', is_admin: false })
            loadPlayers()
        } catch (error) {
            alert('Failed to create player: ' + (error instanceof Error ? error.message : 'Unknown error'))
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
                    <h1>Player Management</h1>
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ Add Player'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Add New Player</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Player Name</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        required
                                        placeholder="John Doe"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Email</label>
                                    <input
                                        type="email"
                                        className="form-input"
                                        value={formData.email}
                                        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                        required
                                        placeholder="john@example.com"
                                    />
                                </div>
                            </div>
                            <div className="form-group">
                                <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)', cursor: 'pointer' }}>
                                    <input
                                        type="checkbox"
                                        className="form-checkbox"
                                        checked={formData.is_admin}
                                        onChange={(e) => setFormData({ ...formData, is_admin: e.target.checked })}
                                    />
                                    <span style={{ color: 'var(--color-text-secondary)' }}>Admin privileges</span>
                                </label>
                            </div>
                            <button type="submit" className="btn btn-success">
                                Add Player
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Players ({players.length})</h3>
                    {players.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No players added yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Email</th>
                                        <th>Status</th>
                                        <th>Established</th>
                                        <th>Admin</th>
                                        <th>Linked</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {players.map((player) => (
                                        <tr key={player.id}>
                                            <td style={{ fontWeight: '600' }}>{player.name}</td>
                                            <td>{player.email}</td>
                                            <td>
                                                <span className={`badge ${player.active ? 'badge-success' : 'badge-danger'}`}>
                                                    {player.active ? 'Active' : 'Inactive'}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`badge ${player.established ? 'badge-success' : 'badge-warning'}`}>
                                                    {player.established ? 'Yes' : 'No'}
                                                </span>
                                            </td>
                                            <td>{player.is_admin ? '✓' : ''}</td>
                                            <td>{player.clerk_user_id ? '✓' : ''}</td>
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
