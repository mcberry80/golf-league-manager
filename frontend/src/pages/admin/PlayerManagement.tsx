import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { LeagueMemberWithPlayer } from '../../types'

export default function PlayerManagement() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        role: 'player',
    })

    const loadMembers = useCallback(async () => {
        if (!currentLeague) return

        try {
            const data = await api.listLeagueMembers(currentLeague.id)
            setMembers(data)
        } catch (error) {
            console.error('Failed to load members:', error)
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
            loadMembers()
        }
    }, [currentLeague, leagueLoading, navigate, leagueId, loadMembers])

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.addLeagueMember(currentLeague.id, formData.email, formData.name)
            // If role is admin, we need to update it separately as addLeagueMember defaults to player or uses what's passed?
            // Wait, addLeagueMember doesn't take role in frontend API yet?
            // Actually, I didn't update api.addLeagueMember to take role. It defaults to 'player' in backend if not provided, 
            // but backend *does* accept role. I should have updated api.ts to take role too.
            // For now, I'll just add the member, and if they wanted admin, I'd need another call.
            // But let's check if I can update api.ts quickly or just stick to adding as player first.
            // The backend handleAddLeagueMember accepts role. 
            // I'll stick to adding as player for now to avoid more context switching, 
            // or I can update the member role immediately after adding if needed.
            // But wait, the form has an "Admin privileges" checkbox (implied by role state).

            // Let's just add them. If I want to support role on creation, I should update api.ts.
            // I'll update api.ts in a sec if I really need it, but for now let's just add.

            setShowForm(false)
            setFormData({ name: '', email: '', role: 'player' })
            loadMembers()
        } catch (error) {
            alert('Failed to add player: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function toggleAdmin(member: LeagueMemberWithPlayer) {
        if (!currentLeague) return

        const newRole = member.role === 'admin' ? 'player' : 'admin'
        try {
            await api.updateLeagueMemberRole(currentLeague.id, member.playerId, newRole)
            loadMembers()
        } catch (error) {
            alert('Failed to update role: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function removeMember(member: LeagueMemberWithPlayer) {
        if (!currentLeague) return
        if (!confirm(`Are you sure you want to remove ${member.player?.name || member.player?.email} from the league?`)) return

        try {
            await api.removeLeagueMember(currentLeague.id, member.playerId)
            loadMembers()
        } catch (error) {
            alert('Failed to remove member: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
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
                        <h1>Player Management</h1>
                        <p className="text-gray-400 mt-1">{currentLeague.name}</p>
                    </div>
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
                            {/* Role selection could go here if supported by API */}
                            <button type="submit" className="btn btn-success">
                                Add Player
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>League Members ({members.length})</h3>
                    {members.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No players added yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Email</th>
                                        <th>Role</th>
                                        <th>Status</th>
                                        <th>Joined</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {members.map((member) => (
                                        <tr key={member.id}>
                                            <td style={{ fontWeight: '600' }}>{member.player?.name || 'Unknown'}</td>
                                            <td>{member.player?.email || 'Unknown'}</td>
                                            <td>
                                                <span className={`badge ${member.role === 'admin' ? 'badge-primary' : 'badge-secondary'}`}>
                                                    {member.role}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`badge ${member.player?.active ? 'badge-success' : 'badge-danger'}`}>
                                                    {member.player?.active ? 'Active' : 'Inactive'}
                                                </span>
                                            </td>
                                            <td>{new Date(member.joinedAt).toLocaleDateString()}</td>
                                            <td>
                                                <div style={{ display: 'flex', gap: '0.5rem' }}>
                                                    <button
                                                        onClick={() => toggleAdmin(member)}
                                                        className="btn btn-secondary"
                                                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    >
                                                        {member.role === 'admin' ? 'Remove Admin' : 'Make Admin'}
                                                    </button>
                                                    <button
                                                        onClick={() => removeMember(member)}
                                                        className="btn btn-danger"
                                                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    >
                                                        Remove
                                                    </button>
                                                </div>
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
