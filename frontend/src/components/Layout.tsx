import { ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { SignedIn, SignedOut, SignInButton, UserButton } from '@clerk/clerk-react'
import { Trophy } from 'lucide-react'
import { useLeague } from '../contexts/LeagueContext'

/**
 * Loading spinner component - centered in viewport
 */
export function LoadingSpinner() {
    return (
        <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <div className="spinner"></div>
        </div>
    )
}

/**
 * Inline loading spinner for use within components
 */
export function InlineSpinner({ size = 24 }: { size?: number }) {
    return (
        <div className="spinner" style={{ width: size, height: size, margin: '0 auto' }}></div>
    )
}

interface PageHeaderProps {
    /** Show the league context in the header */
    showLeagueContext?: boolean
    /** Custom right side content */
    rightContent?: ReactNode
}

/**
 * Shared page header component with authentication and league context
 */
export function PageHeader({ showLeagueContext = true, rightContent }: PageHeaderProps) {
    const navigate = useNavigate()
    const { currentLeague } = useLeague()

    return (
        <header className="border-b" style={{ borderColor: 'var(--color-border)', background: 'rgba(30, 41, 59, 0.8)', backdropFilter: 'blur(10px)' }}>
            <div className="container" style={{ padding: 'var(--spacing-md) var(--spacing-lg)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div className="flex items-center gap-4">
                        <Link to="/" style={{ textDecoration: 'none' }}>
                            <h2 style={{ margin: 0, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text', fontSize: '1.25rem' }}>
                                ⛳ Golf League
                            </h2>
                        </Link>
                        <SignedIn>
                            {showLeagueContext && currentLeague && (
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
                    <div>
                        <SignedOut>
                            <SignInButton mode="modal">
                                <button className="btn btn-primary">Sign In</button>
                            </SignInButton>
                        </SignedOut>
                        <SignedIn>
                            {rightContent || (
                                <div className="flex items-center gap-4">
                                    <button
                                        onClick={() => navigate('/leagues')}
                                        className="btn btn-sm btn-outline md:hidden"
                                    >
                                        Leagues
                                    </button>
                                    <UserButton afterSignOutUrl="/" />
                                </div>
                            )}
                        </SignedIn>
                    </div>
                </div>
            </div>
        </header>
    )
}

interface PageLayoutProps {
    /** Page content */
    children: ReactNode
    /** Whether to show the header */
    showHeader?: boolean
    /** Whether to show league context in the header */
    showLeagueContext?: boolean
    /** Custom right side content for the header */
    headerRightContent?: ReactNode
    /** Whether to use the animate-fade-in class on the container */
    animate?: boolean
    /** Custom padding for the container */
    containerPadding?: string
}

/**
 * Shared page layout component with common styling
 */
export function PageLayout({
    children,
    showHeader = true,
    showLeagueContext = true,
    headerRightContent,
    animate = true,
    containerPadding = 'var(--spacing-2xl)',
}: PageLayoutProps) {
    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            {showHeader && (
                <PageHeader 
                    showLeagueContext={showLeagueContext} 
                    rightContent={headerRightContent} 
                />
            )}
            <main 
                className={`container ${animate ? 'animate-fade-in' : ''}`}
                style={{ 
                    paddingTop: containerPadding, 
                    paddingBottom: containerPadding 
                }}
            >
                {children}
            </main>
        </div>
    )
}

interface AdminPageLayoutProps {
    /** Page content */
    children: ReactNode
    /** Page title */
    title: string
    /** Page subtitle (typically the league name) */
    subtitle?: string
    /** Back link URL (defaults to admin dashboard) */
    backTo?: string
    /** Back link text */
    backText?: string
    /** Action button content */
    actionButton?: ReactNode
}

/**
 * Shared layout for admin pages with consistent back link and title structure
 */
export function AdminPageLayout({
    children,
    title,
    subtitle,
    backTo,
    backText = '← Back to Admin',
    actionButton,
}: AdminPageLayoutProps) {
    const { currentLeague } = useLeague()
    const defaultBackTo = currentLeague ? `/leagues/${currentLeague.id}/admin` : '/leagues'

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link 
                    to={backTo || defaultBackTo} 
                    style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}
                >
                    {backText}
                </Link>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <div>
                        <h1>{title}</h1>
                        {subtitle && <p className="text-gray-400 mt-1">{subtitle}</p>}
                    </div>
                    {actionButton}
                </div>

                {children}
            </div>
        </div>
    )
}

interface ErrorAlertProps {
    /** Error message to display */
    message: string
    /** Optional additional details */
    details?: string[]
}

/**
 * Reusable error alert component
 */
export function ErrorAlert({ message, details }: ErrorAlertProps) {
    return (
        <div className="alert alert-error">
            <strong>{message}</strong>
            {details && details.length > 0 && (
                <ul style={{ marginTop: '0.5rem', marginBottom: 0, paddingLeft: '1.5rem' }}>
                    {details.map((detail, i) => (
                        <li key={i} style={{ fontSize: '0.9rem' }}>{detail}</li>
                    ))}
                </ul>
            )}
        </div>
    )
}

interface AccessDeniedProps {
    /** Message to display */
    message?: string
    /** League name for context */
    leagueName?: string
}

/**
 * Access denied component for unauthorized access
 */
export function AccessDenied({ message, leagueName }: AccessDeniedProps) {
    return (
        <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
            <div className="alert alert-error">
                <strong>Access Denied:</strong> {message || `You must be an admin of ${leagueName || 'this league'} to access this page.`}
            </div>
            <Link to="/" className="btn btn-secondary" style={{ marginTop: 'var(--spacing-lg)' }}>
                Return Home
            </Link>
        </div>
    )
}

interface NoLeagueSelectedProps {
    /** Message to display */
    message?: string
    /** Call to action text */
    ctaText?: string
    /** Navigate to this path when CTA is clicked */
    ctaPath?: string
}

/**
 * Component shown when no league is selected
 */
export function NoLeagueSelected({ 
    message = 'Select a league to view this content.',
    ctaText = 'Select League',
    ctaPath = '/leagues'
}: NoLeagueSelectedProps) {
    const navigate = useNavigate()
    
    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <div className="text-center">
                    <Trophy className="w-16 h-16 text-emerald-500 mx-auto mb-4" />
                    <h1 className="text-2xl font-bold text-white mb-2">No League Selected</h1>
                    <p className="text-gray-400 mb-6">{message}</p>
                    <button
                        onClick={() => navigate(ctaPath)}
                        className="btn btn-primary"
                    >
                        {ctaText}
                    </button>
                </div>
            </div>
        </div>
    )
}
