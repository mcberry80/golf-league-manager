import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type { Player, Round, HandicapRecord } from '../types'

export default function Players() {
    const { currentLeague, isLoading: leagueLoading } = useLeague()
    const [player, setPlayer] = useState<Player | null>(null)
    const [handicap, setHandicap] = useState<HandicapRecord | null>(null)
    const [rounds, setRounds] = useState<Round[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        async function loadPlayerData() {
            try {
                const userInfo = await api.getCurrentUser()

                if (userInfo.linked && userInfo.player) {
                    setPlayer(userInfo.player)

                    if (currentLeague) {
                        // Load handicap for current league
                        try {
                            const handicapData = await api.getPlayerHandicap(currentLeague.id, userInfo.player.id)
                            setHandicap(handicapData)
                        } catch (err) {
                            console.log('No handicap yet')
                            setHandicap(null)
                        }

                        // Load rounds for current league
                        try {
                            const roundsData = await api.getPlayerRounds(currentLeague.id, userInfo.player.id)
                            setRounds(roundsData)
                        } catch (err) {
                            console.log('No rounds yet')
                            setRounds([])
                        }
                    }
                }
            } catch (error) {
                console.error('Failed to load player data:', error)
            } finally {
                setLoading(false)
            }
        }

        if (!leagueLoading) {
            loadPlayerData()
        }
    }, [currentLeague, leagueLoading])

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!player) {
        return (
            <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                <div className="alert alert-warning">
                    <strong>Not Linked:</strong> Please link your account to a player profile first.
                </div>
                <Link to="/link-account" className="btn btn-primary" style={{ marginTop: 'var(--spacing-lg)' }}>
                    Link Account
                </Link>
            </div>
        )
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Home
                </Link>

                <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h1>My Profile</h1>
                    {currentLeague && <p className="text-gray-400 mt-1">{currentLeague.name}</p>}
                </div>

                {/* Player Info */}
                <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                    <h2 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>{player.name}</h2>
                    <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                        <div>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Email</p>
                            <p style={{ color: 'var(--color-text)' }}>{player.email}</p>
                        </div>
                        <div>
                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Status</p>
                            <span className={`badge ${player.active ? 'badge-success' : 'badge-danger'}`}>
                                {player.active ? 'Active' : 'Inactive'}
                            </span>
                        </div>
                    </div>
                </div>

                {!currentLeague ? (
                    <div className="alert alert-info">
                        Select a league to view your handicap and round history.
                    </div>
                ) : (
                    <>
                        {/* Handicap */}
                        {handicap && (
                            <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                                <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Current Handicap</h3>
                                <div className="grid grid-cols-3" style={{ gap: 'var(--spacing-lg)' }}>
                                    <div>
                                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>League Handicap</p>
                                        <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-primary)' }}>
                                            {handicap.league_handicap.toFixed(1)}
                                        </p>
                                    </div>
                                    <div>
                                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Course Handicap</p>
                                        <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-text)' }}>
                                            {handicap.course_handicap.toFixed(1)}
                                        </p>
                                    </div>
                                    <div>
                                        <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: 'var(--spacing-xs)' }}>Playing Handicap</p>
                                        <p style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-text)' }}>
                                            {handicap.playing_handicap}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* Recent Rounds */}
                        <div className="card-glass">
                            <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Recent Rounds</h3>
                            {rounds.length === 0 ? (
                                <p style={{ color: 'var(--color-text-muted)' }}>No rounds recorded yet for this league.</p>
                            ) : (
                                <div className="table-container">
                                    <table className="table">
                                        <thead>
                                            <tr>
                                                <th>Date</th>
                                                <th>Gross Score</th>
                                                <th>Adjusted Score</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {rounds.map((round) => (
                                                <tr key={round.id}>
                                                    <td>{new Date(round.date).toLocaleDateString()}</td>
                                                    <td>{round.total_gross}</td>
                                                    <td>{round.total_adjusted}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
