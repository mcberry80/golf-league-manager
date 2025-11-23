import { useState, useEffect } from 'react'
import { useAuth } from '@clerk/clerk-react'
import { Link } from 'react-router-dom'
import api from '../lib/api'
import type { Player } from '../types'

export default function LinkAccount() {
    const { getToken } = useAuth()
    const [email, setEmail] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')
    const [success, setSuccess] = useState(false)
    const [linkedPlayer, setLinkedPlayer] = useState<Player | null>(null)

    useEffect(() => {
        async function checkLinkStatus() {
            try {
                api.setAuthTokenProvider(getToken)
                const userInfo = await api.getCurrentUser()
                if (userInfo.linked && userInfo.player) {
                    setLinkedPlayer(userInfo.player)
                    setSuccess(true)
                }
            } catch {
                // Not linked yet, that's okay
            }
        }
        checkLinkStatus()
    }, [getToken])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)
        setError('')

        try {
            api.setAuthTokenProvider(getToken)
            const player = await api.linkPlayerAccount({ email })
            setLinkedPlayer(player)
            setSuccess(true)
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to link account')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <div style={{ maxWidth: '500px', margin: '0 auto' }}>
                    <Link to="/" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                        ← Back to Home
                    </Link>

                    <div className="card-glass">
                        <h2 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Link Your Account</h2>

                        {success && linkedPlayer ? (
                            <div>
                                <div className="alert alert-success">
                                    <strong>Success!</strong> Your account is linked to {linkedPlayer.name}
                                </div>
                                <div style={{ marginTop: 'var(--spacing-lg)' }}>
                                    <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-md)' }}>
                                        <strong>Player:</strong> {linkedPlayer.name}
                                    </p>
                                    <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-md)' }}>
                                        <strong>Email:</strong> {linkedPlayer.email}
                                    </p>
                                    <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)' }}>
                                        <strong>Status:</strong> {linkedPlayer.active ? 'Active' : 'Inactive'}
                                    </p>
                                    <Link to="/" className="btn btn-primary" style={{ width: '100%' }}>
                                        Go to Dashboard
                                    </Link>
                                </div>
                            </div>
                        ) : (
                            <form onSubmit={handleSubmit}>
                                <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-lg)' }}>
                                    Enter the email address that the league admin used when creating your player profile.
                                </p>

                                {error && (
                                    <div className="alert alert-error">
                                        {error}
                                    </div>
                                )}

                                <div className="form-group">
                                    <label className="form-label" htmlFor="email">
                                        Email Address
                                    </label>
                                    <input
                                        id="email"
                                        type="email"
                                        className="form-input"
                                        value={email}
                                        onChange={(e) => setEmail(e.target.value)}
                                        required
                                        placeholder="your.email@example.com"
                                    />
                                </div>

                                <button
                                    type="submit"
                                    className="btn btn-primary"
                                    disabled={loading}
                                    style={{ width: '100%' }}
                                >
                                    {loading ? 'Linking...' : 'Link Account'}
                                </button>
                            </form>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
