import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { SeasonPlayerWithPlayer, LeagueMemberWithPlayer, Season } from '../../types'
import { LoadingSpinner, AccessDenied } from '../../components/Layout'

export default function SeasonPlayerManagement() {
    const { leagueId, seasonId } = useParams<{ leagueId: string; seasonId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [seasonPlayers, setSeasonPlayers] = useState<SeasonPlayerWithPlayer[]>([])
    const [leagueMembers, setLeagueMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [season, setSeason] = useState<Season | null>(null)
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        playerId: '',
        provisionalHandicap: 0,
    })
    const [editingHandicap, setEditingHandicap] = useState<string | null>(null)
    const [editHandicapValue, setEditHandicapValue] = useState<number>(0)

    const loadData = useCallback(async () => {
        if (!currentLeague || !seasonId) return

        try {
            const [seasonPlayersData, membersData, seasonData] = await Promise.all([
                api.listSeasonPlayers(currentLeague.id, seasonId),
                api.listLeagueMembers(currentLeague.id),
                api.getSeason(currentLeague.id, seasonId),
            ])
            setSeasonPlayers(seasonPlayersData)
            setLeagueMembers(membersData)
            setSeason(seasonData)
        } catch (error) {
            console.error('Failed to load data:', error)
        } finally {
            setLoading(false)
        }
    }, [currentLeague, seasonId])

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

        if (currentLeague && seasonId) {
            loadData()
        }
    }, [currentLeague, leagueLoading, navigate, leagueId, seasonId, loadData])

    // Get members not already in the season
    const availableMembers = leagueMembers.filter(
        m => !seasonPlayers.some(sp => sp.playerId === m.playerId)
    )

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague || !seasonId) return

        try {
            await api.addSeasonPlayer(currentLeague.id, seasonId, formData.playerId, formData.provisionalHandicap)
            setShowForm(false)
            setFormData({ playerId: '', provisionalHandicap: 0 })
            loadData()
        } catch (error) {
            alert('Failed to add player to season: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    function startEditHandicap(player: SeasonPlayerWithPlayer) {
        setEditingHandicap(player.playerId)
        setEditHandicapValue(player.provisionalHandicap || 0)
    }

    function cancelEditHandicap() {
        setEditingHandicap(null)
        setEditHandicapValue(0)
    }

    async function saveHandicap(player: SeasonPlayerWithPlayer) {
        if (!currentLeague || !seasonId) return

        try {
            await api.updateSeasonPlayer(currentLeague.id, seasonId, player.playerId, { provisionalHandicap: editHandicapValue })
            setEditingHandicap(null)
            setEditHandicapValue(0)
            loadData()
        } catch (error) {
            alert('Failed to update handicap: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    async function removePlayer(player: SeasonPlayerWithPlayer) {
        if (!currentLeague || !seasonId) return
        if (!confirm(`Are you sure you want to remove ${player.player?.name || player.player?.email} from this season?`)) return

        try {
            await api.removeSeasonPlayer(currentLeague.id, seasonId, player.playerId)
            loadData()
        } catch (error) {
            alert('Failed to remove player from season: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    // Pre-fill handicap when player is selected
    function handlePlayerSelect(playerId: string) {
        const member = leagueMembers.find(m => m.playerId === playerId)
        setFormData({
            playerId,
            provisionalHandicap: member?.provisionalHandicap || 0,
        })
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
                        <h1>Season Players</h1>
                        <p className="text-gray-400 mt-1">
                            {season?.name || 'Loading...'} - {currentLeague.name}
                        </p>
                    </div>
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary" disabled={availableMembers.length === 0}>
                        {showForm ? 'Cancel' : '+ Add Player to Season'}
                    </button>
                </div>

                {showForm && availableMembers.length > 0 && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Add Player to Season</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Player</label>
                                    <select
                                        className="form-select"
                                        value={formData.playerId}
                                        onChange={(e) => handlePlayerSelect(e.target.value)}
                                        required
                                    >
                                        <option value="">Select Player</option>
                                        {availableMembers.map(member => (
                                            <option key={member.playerId} value={member.playerId}>
                                                {member.player?.name || member.player?.email}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Provisional Handicap for Season</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        className="form-input"
                                        value={formData.provisionalHandicap}
                                        onChange={(e) => setFormData({ ...formData, provisionalHandicap: parseFloat(e.target.value) || 0 })}
                                        placeholder="12.5"
                                    />
                                    <small className="text-gray-400" style={{ fontSize: '0.75rem', marginTop: '0.25rem', display: 'block' }}>
                                        Pre-filled from league member handicap. Adjust if needed.
                                    </small>
                                </div>
                            </div>
                            <button type="submit" className="btn btn-success">
                                Add to Season
                            </button>
                        </form>
                    </div>
                )}

                {availableMembers.length === 0 && leagueMembers.length > 0 && !showForm && (
                    <div className="alert alert-info" style={{ marginBottom: 'var(--spacing-lg)' }}>
                        All league members have been added to this season.
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Season Roster ({seasonPlayers.length})</h3>
                    {seasonPlayers.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No players added to this season yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Email</th>
                                        <th>Season Handicap</th>
                                        <th>Added</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {seasonPlayers.map((player) => (
                                        <tr key={player.id}>
                                            <td style={{ fontWeight: '600' }}>{player.player?.name || 'Unknown'}</td>
                                            <td>{player.player?.email || 'Unknown'}</td>
                                            <td>
                                                {editingHandicap === player.playerId ? (
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
                                                            onClick={() => saveHandicap(player)}
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
                                                        onClick={() => startEditHandicap(player)}
                                                        style={{ cursor: 'pointer', textDecoration: 'underline dotted' }}
                                                        title="Click to edit"
                                                    >
                                                        {player.provisionalHandicap?.toFixed(1) || '0.0'}
                                                    </span>
                                                )}
                                            </td>
                                            <td>{new Date(player.addedAt).toLocaleDateString()}</td>
                                            <td>
                                                <button
                                                    onClick={() => removePlayer(player)}
                                                    className="btn btn-danger"
                                                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                                                >
                                                    Remove from Season
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
