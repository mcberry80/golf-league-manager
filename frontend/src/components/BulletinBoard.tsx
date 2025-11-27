import { useState, useEffect, useCallback } from 'react'
import { MessageSquare, Send, Trash2, RefreshCw } from 'lucide-react'
import api from '../lib/api'
import { formatRelativeTime } from '../lib/utils'
import type { BulletinMessage } from '../types'

interface BulletinBoardProps {
    leagueId: string
    seasonId: string
    currentPlayerId?: string
    isAdmin?: boolean
}

export default function BulletinBoard({ leagueId, seasonId, currentPlayerId, isAdmin }: BulletinBoardProps) {
    const [messages, setMessages] = useState<BulletinMessage[]>([])
    const [newMessage, setNewMessage] = useState('')
    const [loading, setLoading] = useState(true)
    const [posting, setPosting] = useState(false)
    const [error, setError] = useState('')

    const loadMessages = useCallback(async () => {
        try {
            setLoading(true)
            setError('')
            const data = await api.listBulletinMessages(leagueId, seasonId, 50)
            setMessages(data)
        } catch (err) {
            // Check for access denied errors (HTTP 403)
            const errorMessage = err instanceof Error ? err.message : 'Failed to load messages'
            if (errorMessage.includes('403') || errorMessage.toLowerCase().includes('forbidden') || errorMessage.toLowerCase().includes('access denied')) {
                setError('You must be a season player to view the bulletin board.')
            } else {
                setError(errorMessage)
            }
        } finally {
            setLoading(false)
        }
    }, [leagueId, seasonId])

    useEffect(() => {
        loadMessages()
    }, [loadMessages])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        if (!newMessage.trim()) return

        try {
            setPosting(true)
            setError('')
            const message = await api.createBulletinMessage(leagueId, seasonId, newMessage.trim())
            setMessages([message, ...messages])
            setNewMessage('')
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to post message')
        } finally {
            setPosting(false)
        }
    }

    const handleDelete = async (messageId: string) => {
        if (!confirm('Are you sure you want to delete this message?')) return

        try {
            await api.deleteBulletinMessage(leagueId, messageId)
            setMessages(messages.filter(m => m.id !== messageId))
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to delete message')
        }
    }

    return (
        <div className="bulletin-board" style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            {/* Header */}
            <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center',
                marginBottom: 'var(--spacing-md)',
                flexShrink: 0
            }}>
                <h3 style={{ 
                    margin: 0, 
                    color: 'var(--color-text)', 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.5rem',
                    fontSize: '1.1rem'
                }}>
                    <MessageSquare className="w-5 h-5 text-blue-400" />
                    Bulletin Board
                </h3>
                <button
                    onClick={loadMessages}
                    disabled={loading}
                    className="btn btn-secondary"
                    style={{ padding: '0.5rem', minWidth: 'auto' }}
                    title="Refresh messages"
                >
                    <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                </button>
            </div>

            {error && (
                <div className="alert alert-error" style={{ marginBottom: 'var(--spacing-md)', fontSize: '0.875rem', flexShrink: 0 }}>
                    {error}
                </div>
            )}

            {/* Post form */}
            <form onSubmit={handleSubmit} style={{ marginBottom: 'var(--spacing-md)', flexShrink: 0 }}>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <input
                        type="text"
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        placeholder="Post a message, talk some trash..."
                        maxLength={1000}
                        className="form-input"
                        style={{ flex: 1, fontSize: '0.875rem' }}
                        disabled={posting}
                    />
                    <button
                        type="submit"
                        disabled={posting || !newMessage.trim()}
                        className="btn btn-primary"
                        style={{ padding: '0.5rem 1rem', minWidth: 'auto' }}
                    >
                        <Send className="w-4 h-4" />
                    </button>
                </div>
            </form>

            {/* Messages list */}
            <div style={{ 
                flex: 1, 
                overflowY: 'auto',
                minHeight: 0
            }}>
                {loading && messages.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 'var(--spacing-lg)' }}>
                        <div className="spinner" style={{ width: '24px', height: '24px', margin: '0 auto' }}></div>
                    </div>
                ) : messages.length === 0 ? (
                    <div style={{ 
                        textAlign: 'center', 
                        padding: 'var(--spacing-lg)',
                        color: 'var(--color-text-muted)'
                    }}>
                        <MessageSquare className="w-8 h-8 mx-auto mb-2 opacity-50" />
                        <p style={{ margin: 0 }}>No messages yet. Be the first to post!</p>
                    </div>
                ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                        {messages.map((message) => (
                            <div 
                                key={message.id}
                                style={{
                                    background: 'rgba(255, 255, 255, 0.03)',
                                    borderRadius: 'var(--radius-md)',
                                    padding: '0.75rem',
                                    border: '1px solid var(--color-border)',
                                }}
                            >
                                <div style={{ 
                                    display: 'flex', 
                                    justifyContent: 'space-between', 
                                    alignItems: 'flex-start',
                                    marginBottom: '0.25rem'
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                                        <span style={{ 
                                            fontWeight: 600, 
                                            color: 'var(--color-primary-light)',
                                            fontSize: '0.875rem'
                                        }}>
                                            {message.playerName}
                                        </span>
                                        <span style={{ 
                                            color: 'var(--color-text-muted)',
                                            fontSize: '0.75rem'
                                        }}>
                                            {formatRelativeTime(message.createdAt)}
                                        </span>
                                    </div>
                                    {(message.playerId === currentPlayerId || isAdmin) && (
                                        <button
                                            onClick={() => handleDelete(message.id)}
                                            className="btn"
                                            style={{ 
                                                padding: '0.25rem',
                                                background: 'transparent',
                                                border: 'none',
                                                color: 'var(--color-text-muted)',
                                                cursor: 'pointer',
                                                minWidth: 'auto'
                                            }}
                                            title="Delete message"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    )}
                                </div>
                                <p style={{ 
                                    margin: 0, 
                                    color: 'var(--color-text)',
                                    fontSize: '0.875rem',
                                    lineHeight: 1.5,
                                    wordBreak: 'break-word'
                                }}>
                                    {message.content}
                                </p>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
}
