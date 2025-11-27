import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuth, SignInButton, SignedIn, SignedOut } from '@clerk/clerk-react'
import { Trophy, CheckCircle, XCircle, Clock, Users } from 'lucide-react'
import api from '../lib/api'
import type { InviteDetails } from '../types'

export default function JoinLeague() {
    const { token } = useParams<{ token: string }>()
    const navigate = useNavigate()
    const { getToken, isSignedIn } = useAuth()

    const [inviteDetails, setInviteDetails] = useState<InviteDetails | null>(null)
    const [loading, setLoading] = useState(true)
    const [accepting, setAccepting] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [success, setSuccess] = useState<string | null>(null)

    useEffect(() => {
        async function loadInvite() {
            if (!token) {
                setError('Invalid invite link')
                setLoading(false)
                return
            }

            try {
                api.setAuthTokenProvider(getToken)
                const details = await api.getInviteByToken(token)
                setInviteDetails(details)
            } catch (err) {
                const message = err instanceof Error ? err.message : 'Failed to load invite'
                if (message.includes('revoked')) {
                    setError('This invite link has been revoked')
                } else if (message.includes('expired')) {
                    setError('This invite link has expired')
                } else if (message.includes('maximum uses')) {
                    setError('This invite link has reached its maximum uses')
                } else if (message.includes('not found')) {
                    setError('This invite link is invalid or no longer exists')
                } else {
                    setError(message)
                }
            } finally {
                setLoading(false)
            }
        }

        if (isSignedIn) {
            loadInvite()
        } else {
            setLoading(false)
        }
    }, [token, getToken, isSignedIn])

    const handleAcceptInvite = async () => {
        if (!token) return

        setAccepting(true)
        setError(null)

        try {
            api.setAuthTokenProvider(getToken)
            const result = await api.acceptInvite(token)
            setSuccess(result.message)

            // Redirect to the league dashboard after a short delay
            setTimeout(() => {
                navigate(`/leagues/${inviteDetails?.league.id}/dashboard`)
            }, 2000)
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to join league')
        } finally {
            setAccepting(false)
        }
    }

    // Show loading state
    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--gradient-dark)' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    // Show sign-in prompt for unauthenticated users
    if (!isSignedIn) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                    <div style={{ maxWidth: '500px', margin: '0 auto' }}>
                        <div className="card-glass" style={{ textAlign: 'center' }}>
                            <div style={{
                                background: 'rgba(16, 185, 129, 0.2)',
                                borderRadius: 'var(--radius-full)',
                                padding: 'var(--spacing-lg)',
                                display: 'inline-flex',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <Users className="h-10 w-10" style={{ color: 'var(--color-accent)' }} />
                            </div>

                            <h2 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text)' }}>
                                You've Been Invited!
                            </h2>

                            <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-xl)' }}>
                                Sign in to accept this invitation and join the league.
                            </p>

                            <SignedOut>
                                <SignInButton mode="modal">
                                    <button className="btn btn-primary" style={{ width: '100%' }}>
                                        Sign In to Continue
                                    </button>
                                </SignInButton>
                            </SignedOut>

                            <p style={{ marginTop: 'var(--spacing-lg)', fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                                Don't have an account? You can create one when you sign in.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    // Show error state
    if (error && !inviteDetails) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                    <div style={{ maxWidth: '500px', margin: '0 auto' }}>
                        <div className="card-glass" style={{ textAlign: 'center' }}>
                            <div style={{
                                background: 'rgba(239, 68, 68, 0.2)',
                                borderRadius: 'var(--radius-full)',
                                padding: 'var(--spacing-lg)',
                                display: 'inline-flex',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <XCircle className="h-10 w-10" style={{ color: 'var(--color-danger)' }} />
                            </div>

                            <h2 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text)' }}>
                                Invalid Invite
                            </h2>

                            <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-xl)' }}>
                                {error}
                            </p>

                            <Link to="/" className="btn btn-primary" style={{ width: '100%' }}>
                                Go to Home
                            </Link>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    // Show success state
    if (success) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                    <div style={{ maxWidth: '500px', margin: '0 auto' }}>
                        <div className="card-glass" style={{ textAlign: 'center' }}>
                            <div style={{
                                background: 'rgba(16, 185, 129, 0.2)',
                                borderRadius: 'var(--radius-full)',
                                padding: 'var(--spacing-lg)',
                                display: 'inline-flex',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <CheckCircle className="h-10 w-10" style={{ color: 'var(--color-accent)' }} />
                            </div>

                            <h2 style={{ marginBottom: 'var(--spacing-md)', color: 'var(--color-text)' }}>
                                Welcome to {inviteDetails?.league.name}!
                            </h2>

                            <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-xl)' }}>
                                {success}
                            </p>

                            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>
                                Redirecting to the league dashboard...
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    // Show invite details and accept button
    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <div style={{ maxWidth: '500px', margin: '0 auto' }}>
                    <div className="card-glass">
                        <div style={{ textAlign: 'center', marginBottom: 'var(--spacing-xl)' }}>
                            <div style={{
                                background: 'rgba(16, 185, 129, 0.2)',
                                borderRadius: 'var(--radius-full)',
                                padding: 'var(--spacing-lg)',
                                display: 'inline-flex',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <Trophy className="h-10 w-10" style={{ color: 'var(--color-accent)' }} />
                            </div>

                            <h2 style={{ marginBottom: 'var(--spacing-sm)', color: 'var(--color-text)' }}>
                                Join {inviteDetails?.league.name}
                            </h2>

                            <p style={{ color: 'var(--color-text-secondary)' }}>
                                You've been invited to join this golf league
                            </p>
                        </div>

                        {inviteDetails?.league.description && (
                            <div style={{
                                background: 'rgba(0, 0, 0, 0.2)',
                                borderRadius: 'var(--radius-md)',
                                padding: 'var(--spacing-md)',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <p style={{ color: 'var(--color-text-secondary)', margin: 0, fontSize: '0.875rem' }}>
                                    {inviteDetails.league.description}
                                </p>
                            </div>
                        )}

                        <div style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 'var(--spacing-sm)',
                            color: 'var(--color-text-muted)',
                            fontSize: '0.875rem',
                            marginBottom: 'var(--spacing-xl)'
                        }}>
                            <Clock className="h-4 w-4" />
                            <span>
                                Invite expires {new Date(inviteDetails?.invite.expiresAt || '').toLocaleDateString()}
                            </span>
                        </div>

                        {error && (
                            <div className="alert alert-error" style={{ marginBottom: 'var(--spacing-lg)' }}>
                                {error}
                            </div>
                        )}

                        <SignedIn>
                            <button
                                onClick={handleAcceptInvite}
                                disabled={accepting}
                                className="btn btn-primary"
                                style={{ width: '100%' }}
                            >
                                {accepting ? 'Joining...' : 'Accept Invitation'}
                            </button>
                        </SignedIn>

                        <Link
                            to="/"
                            style={{
                                display: 'block',
                                textAlign: 'center',
                                marginTop: 'var(--spacing-md)',
                                color: 'var(--color-text-muted)',
                                fontSize: '0.875rem'
                            }}
                        >
                            Cancel
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    )
}
