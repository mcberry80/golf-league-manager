import { SignedIn, SignedOut, SignInButton, UserButton } from '@clerk/clerk-react'
import { Link, useNavigate } from 'react-router-dom'
import { Trophy } from 'lucide-react'
import type { League } from '../types'

interface PageHeaderProps {
    currentLeague: League | null
    showSignIn?: boolean
}

export default function PageHeader({ currentLeague, showSignIn = true }: PageHeaderProps) {
    const navigate = useNavigate()

    return (
        <header className="border-b" style={{ borderColor: 'var(--color-border)', background: 'rgba(30, 41, 59, 0.8)', backdropFilter: 'blur(10px)' }}>
            <div className="container" style={{ padding: 'var(--spacing-lg)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div className="flex items-center gap-4">
                        <Link to="/" style={{ textDecoration: 'none' }}>
                            <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>
                                ⛳ Golf League
                            </h2>
                        </Link>
                        <SignedIn>
                            {currentLeague && (
                                <div className="hidden md:flex items-center text-gray-400 text-sm border-l border-gray-700 pl-4 ml-4">
                                    <Trophy className="w-4 h-4 mr-2 text-emerald-500" />
                                    <span className="text-gray-200 font-medium">{currentLeague.name}</span>
                                    <button
                                        onClick={() => navigate('/leagues')}
                                        className="btn btn-sm btn-outline ml-2"
                                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.5rem', height: 'auto' }}
                                    >
                                        Switch
                                    </button>
                                </div>
                            )}
                        </SignedIn>
                    </div>
                    <div className="flex items-center gap-4">
                        {showSignIn && (
                            <SignedOut>
                                <SignInButton mode="modal">
                                    <button className="btn btn-primary">
                                        Sign In
                                    </button>
                                </SignInButton>
                            </SignedOut>
                        )}
                        <SignedIn>
                            <button
                                onClick={() => navigate('/leagues')}
                                className="btn btn-sm btn-outline md:hidden"
                            >
                                Leagues
                            </button>
                            <UserButton afterSignOutUrl="/" />
                        </SignedIn>
                    </div>
                </div>
            </div>
        </header>
    )
}
