import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { LeagueMemberWithPlayer, LeagueInvite } from '../../types'
import { Copy, Trash2, Link as LinkIcon, Users, Clock } from 'lucide-react'
import { LoadingSpinner, AccessDenied } from '../../components/Layout'

// Number of characters to show from the invite token in the UI
const INVITE_TOKEN_PREVIEW_LENGTH = 8

export default function PlayerManagement() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [invites, setInvites] = useState<LeagueInvite[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [showInviteForm, setShowInviteForm] = useState(false)
    const [copiedInvite, setCopiedInvite] = useState<string | null>(null)
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        role: 'player',
        provisionalHandicap: 0,
    })
    const [inviteFormData, setInviteFormData] = useState({
        expiresInDays: 7,
        maxUses: 0,
    })
    const [editingHandicap, setEditingHandicap] = useState<string | null>(null)
    const [editHandicapValue, setEditHandicapValue] = useState<number>(0)

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

    const loadInvites = useCallback(async () => {
        if (!currentLeague) return

        try {
            const data = await api.listLeagueInvites(currentLeague.id)
            setInvites(data)
        } catch (error) {
            console.error('Failed to load invites:', error)
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
            loadInvites()
        }
    }, [currentLeague, leagueLoading, navigate, leagueId, loadMembers, loadInvites])

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.addLeagueMember(currentLeague.id, formData.email, formData.name, formData.provisionalHandicap)
            setShowForm(false)
            setFormData({ name: '', email: '', role: 'player', provisionalHandicap: 0 })
            loadMembers()
        } catch (error) {
            alert('Failed to add player: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function handleCreateInvite(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.createLeagueInvite(
                currentLeague.id,
                inviteFormData.expiresInDays,
                inviteFormData.maxUses
            )
            setShowInviteForm(false)
            setInviteFormData({ expiresInDays: 7, maxUses: 0 })
            loadInvites()
        } catch (error) {
            alert('Failed to create invite: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function revokeInvite(inviteId: string) {
        if (!currentLeague) return
        if (!confirm('Are you sure you want to revoke this invite link?')) return

        try {
            await api.revokeLeagueInvite(currentLeague.id, inviteId)
            loadInvites()
        } catch (error) {
            alert('Failed to revoke invite: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    function copyInviteLink(token: string) {
        const url = `${window.location.origin}/invite/${token}`
        navigator.clipboard.writeText(url)
        setCopiedInvite(token)
        setTimeout(() => setCopiedInvite(null), 2000)
    }

    function getInviteStatus(invite: LeagueInvite): { status: string; color: string } {
        if (invite.revokedAt) {
            return { status: 'Revoked', color: 'var(--color-danger)' }
        }
        if (new Date(invite.expiresAt) < new Date()) {
            return { status: 'Expired', color: 'var(--color-warning)' }
        }
        if (invite.maxUses > 0 && invite.useCount >= invite.maxUses) {
            return { status: 'Max Uses Reached', color: 'var(--color-warning)' }
        }
        return { status: 'Active', color: 'var(--color-accent)' }
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

    function startEditHandicap(member: LeagueMemberWithPlayer) {
        setEditingHandicap(member.playerId)
        setEditHandicapValue(member.provisionalHandicap || 0)
    }

    function cancelEditHandicap() {
        setEditingHandicap(null)
        setEditHandicapValue(0)
    }

    async function saveHandicap(member: LeagueMemberWithPlayer) {
        if (!currentLeague) return

        try {
            await api.updateLeagueMember(currentLeague.id, member.playerId, { provisionalHandicap: editHandicapValue })
            setEditingHandicap(null)
            setEditHandicapValue(0)
            loadMembers()
        } catch (error) {
            alert('Failed to update handicap: ' + (error instanceof Error ? error.message : 'Unknown error'))
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
        return <LoadingSpinner />
    }

    if (!currentLeague || userRole !== 'admin') {
        return <AccessDenied leagueName={currentLeague?.name} />
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
                            <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-lg)' }}>
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
                                <div className="form-group">
                                    <label className="form-label">Provisional Handicap</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        className="form-input"
                                        value={formData.provisionalHandicap}
                                        onChange={(e) => setFormData({ ...formData, provisionalHandicap: parseFloat(e.target.value) || 0 })}
                                        placeholder="12.5"
                                    />
                                    <small className="text-gray-400" style={{ fontSize: '0.75rem', marginTop: '0.25rem', display: 'block' }}>
                                        Starting handicap for this season (see League Rules 3.2)
                                    </small>
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
                                        <th>Provisional Handicap</th>
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
                                                {editingHandicap === member.playerId ? (
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <input
                                                            type="number"
                                                            step="0.1"
                                                            className="form-input"
                                                            value={editHandicapValue}
                                                            onChange={(e) => setEditHandicapValue(parseFloat(e.target.value) || 0)}
                                                            style={{ width: '80px', padding: '0.25rem 0.5rem' }}
                                                            autoFocus
                                                        />
                                                        <button
                                                            onClick={() => saveHandicap(member)}
                                                            className="btn btn-success"
                                                            style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                        >
                                                            Save
                                                        </button>
                                                        <button
                                                            onClick={cancelEditHandicap}
                                                            className="btn btn-secondary"
                                                            style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                        >
                                                            Cancel
                                                        </button>
                                                    </div>
                                                ) : (
                                                    <span
                                                        onClick={() => startEditHandicap(member)}
                                                        style={{ cursor: 'pointer', textDecoration: 'underline dotted' }}
                                                        title="Click to edit"
                                                    >
                                                        {member.provisionalHandicap?.toFixed(1) || '0.0'}
                                                    </span>
                                                )}
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

                {/* Invite Links Section */}
                <div className="card-glass" style={{ marginTop: 'var(--spacing-xl)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-lg)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)' }}>
                            <LinkIcon className="h-5 w-5" style={{ color: 'var(--color-accent)' }} />
                            <h3 style={{ margin: 0, color: 'var(--color-text)' }}>Invite Links</h3>
                        </div>
                        <button onClick={() => setShowInviteForm(!showInviteForm)} className="btn btn-primary" style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}>
                            {showInviteForm ? 'Cancel' : '+ Create Invite Link'}
                        </button>
                    </div>

                    <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)', fontSize: '0.875rem' }}>
                        Create invite links to allow new players to join your league. Share these links with potential members.
                    </p>

                    {showInviteForm && (
                        <div style={{ background: 'rgba(0, 0, 0, 0.2)', borderRadius: 'var(--radius-md)', padding: 'var(--spacing-lg)', marginBottom: 'var(--spacing-lg)' }}>
                            <form onSubmit={handleCreateInvite}>
                                <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)', marginBottom: 'var(--spacing-lg)' }}>
                                    <div className="form-group">
                                        <label className="form-label">Expires In (Days)</label>
                                        <input
                                            type="number"
                                            min="1"
                                            max="30"
                                            className="form-input"
                                            value={inviteFormData.expiresInDays}
                                            onChange={(e) => setInviteFormData({ ...inviteFormData, expiresInDays: parseInt(e.target.value) || 7 })}
                                        />
                                        <small className="text-gray-400" style={{ fontSize: '0.75rem', marginTop: '0.25rem', display: 'block' }}>
                                            1-30 days
                                        </small>
                                    </div>
                                    <div className="form-group">
                                        <label className="form-label">Max Uses</label>
                                        <input
                                            type="number"
                                            min="0"
                                            className="form-input"
                                            value={inviteFormData.maxUses}
                                            onChange={(e) => setInviteFormData({ ...inviteFormData, maxUses: parseInt(e.target.value) || 0 })}
                                        />
                                        <small className="text-gray-400" style={{ fontSize: '0.75rem', marginTop: '0.25rem', display: 'block' }}>
                                            0 = unlimited
                                        </small>
                                    </div>
                                </div>
                                <button type="submit" className="btn btn-success">
                                    Create Invite Link
                                </button>
                            </form>
                        </div>
                    )}

                    {invites.length === 0 ? (
                        <div style={{ textAlign: 'center', padding: 'var(--spacing-xl)', color: 'var(--color-text-muted)' }}>
                            <Users className="h-8 w-8 mx-auto mb-2" style={{ opacity: 0.5 }} />
                            <p>No invite links created yet.</p>
                            <p style={{ fontSize: '0.875rem' }}>Create an invite link to allow new players to join your league.</p>
                        </div>
                    ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
                            {invites.map((invite) => {
                                const { status, color } = getInviteStatus(invite)
                                const isActive = status === 'Active'

                                return (
                                    <div
                                        key={invite.id}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'space-between',
                                            padding: 'var(--spacing-md)',
                                            background: 'rgba(0, 0, 0, 0.2)',
                                            borderRadius: 'var(--radius-md)',
                                            opacity: isActive ? 1 : 0.6
                                        }}
                                    >
                                        <div style={{ flex: 1 }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-sm)', marginBottom: 'var(--spacing-xs)' }}>
                                                <code style={{
                                                    background: 'rgba(0, 0, 0, 0.3)',
                                                    padding: '0.25rem 0.5rem',
                                                    borderRadius: 'var(--radius-sm)',
                                                    fontSize: '0.75rem',
                                                    color: 'var(--color-text-secondary)'
                                                }}>
                                                    {window.location.origin}/invite/{invite.token.substring(0, INVITE_TOKEN_PREVIEW_LENGTH)}...
                                                </code>
                                                <span style={{ fontSize: '0.75rem', color }}>
                                                    {status}
                                                </span>
                                            </div>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)', fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                                <span style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                                                    <Clock className="h-3 w-3" />
                                                    Expires: {new Date(invite.expiresAt).toLocaleDateString()}
                                                </span>
                                                <span>
                                                    Uses: {invite.useCount}{invite.maxUses > 0 ? `/${invite.maxUses}` : ' (unlimited)'}
                                                </span>
                                            </div>
                                        </div>
                                        <div style={{ display: 'flex', gap: 'var(--spacing-sm)' }}>
                                            {isActive && (
                                                <button
                                                    onClick={() => copyInviteLink(invite.token)}
                                                    className="btn btn-secondary"
                                                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    title="Copy invite link"
                                                >
                                                    <Copy className="h-4 w-4" />
                                                    {copiedInvite === invite.token ? 'Copied!' : 'Copy'}
                                                </button>
                                            )}
                                            {isActive && (
                                                <button
                                                    onClick={() => revokeInvite(invite.id)}
                                                    className="btn btn-danger"
                                                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                    title="Revoke invite"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
