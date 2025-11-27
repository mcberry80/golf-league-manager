import { ReactNode } from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'

// Constants
export const HOLE_NUMBERS = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const
export const EMPTY_STROKES_ARRAY = Array(9).fill(0) as number[]

// Colors for win/loss/tie
export const HOLE_RESULT_COLORS = {
    win: { bg: 'rgba(16, 185, 129, 0.3)', text: 'var(--color-accent)' },      // Green for win
    loss: { bg: 'rgba(239, 68, 68, 0.3)', text: 'var(--color-danger)' },      // Red for loss
    tie: { bg: 'rgba(156, 163, 175, 0.3)', text: 'var(--color-text-secondary)' }, // Gray for tie
    none: { bg: 'transparent', text: 'inherit' }
} as const

// Pre-defined styles for golf scoring symbols to avoid creating new objects on every render
const GOLF_SYMBOL_STYLES = {
    eagle: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        border: '2px solid var(--color-accent)',
        boxShadow: '0 0 0 2px var(--color-accent)',
        backgroundColor: 'rgba(16, 185, 129, 0.2)',
        fontWeight: 'bold'
    } as React.CSSProperties,
    birdie: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        border: '2px solid var(--color-accent)',
        fontWeight: 'bold'
    } as React.CSSProperties,
    par: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '22px',
        height: '22px',
        fontWeight: 'bold'
    } as React.CSSProperties,
    bogey: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '22px',
        height: '22px',
        border: '2px solid var(--color-warning)',
        fontWeight: 'bold'
    } as React.CSSProperties,
    doubleBogey: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '22px',
        height: '22px',
        border: '2px solid var(--color-danger)',
        boxShadow: '0 0 0 2px var(--color-danger)',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        fontWeight: 'bold'
    } as React.CSSProperties
} as const

// Golf scoring symbol helper - returns styled element for the score
export function getGolfScoreSymbol(gross: number, par: number): { style: React.CSSProperties, display: string } {
    const diff = gross - par
    
    if (diff <= -2) {
        // Eagle or better - double circle
        return { style: GOLF_SYMBOL_STYLES.eagle, display: gross.toString() }
    } else if (diff === -1) {
        // Birdie - single circle
        return { style: GOLF_SYMBOL_STYLES.birdie, display: gross.toString() }
    } else if (diff === 1) {
        // Bogey - single square
        return { style: GOLF_SYMBOL_STYLES.bogey, display: gross.toString() }
    } else if (diff >= 2) {
        // Double bogey or worse - double square
        return { style: GOLF_SYMBOL_STYLES.doubleBogey, display: gross.toString() }
    } else {
        // Par - plain
        return { style: GOLF_SYMBOL_STYLES.par, display: gross.toString() }
    }
}

// Score Row Component
export interface ScoreRowProps {
    label: string
    scores: number[]
    total: number | string
    color?: string
    bgColor?: string
    withBorder?: boolean
    // Enhanced properties for styling individual cells
    cellColors?: ('win' | 'loss' | 'tie' | 'none')[]
    // For golf scoring symbols
    showGolfSymbols?: boolean
    pars?: number[]
}

export function ScoreRow({ 
    label, 
    scores, 
    total, 
    color, 
    bgColor, 
    withBorder = false,
    cellColors,
    showGolfSymbols = false,
    pars
}: ScoreRowProps) {
    return (
        <tr style={{ 
            borderBottom: withBorder ? '1px solid var(--color-border)' : undefined, 
            background: bgColor 
        }}>
            <td style={{ padding: '0.5rem', color: color || 'var(--color-text-muted)' }}>{label}</td>
            {scores.map((score, i) => {
                const cellColor = cellColors?.[i] ? HOLE_RESULT_COLORS[cellColors[i]] : null
                const parValue = pars?.[i]
                const showSymbol = showGolfSymbols && parValue !== undefined
                const symbolInfo = showSymbol ? getGolfScoreSymbol(score, parValue) : null
                
                return (
                    <td 
                        key={i} 
                        style={{ 
                            padding: '0.5rem', 
                            textAlign: 'center', 
                            color: cellColor?.text || color,
                            backgroundColor: cellColor?.bg || 'transparent'
                        }}
                    >
                        {symbolInfo ? (
                            <span style={symbolInfo.style}>{symbolInfo.display}</span>
                        ) : (
                            score
                        )}
                    </td>
                )
            })}
            <td style={{ padding: '0.5rem', textAlign: 'center', fontWeight: 'bold', color }}>{total}</td>
        </tr>
    )
}

// Scorecard Table Component
export interface ScorecardTableProps {
    rows: Array<{
        label: string
        scores: number[]
        total: number | string
        color?: string
        bgColor?: string
        withBorder?: boolean
        // Enhanced properties for styling individual cells
        cellColors?: ('win' | 'loss' | 'tie' | 'none')[]
        // For golf scoring symbols
        showGolfSymbols?: boolean
        pars?: number[]
    }>
}

export function ScorecardTable({ rows }: ScorecardTableProps) {
    return (
        <div className="table-container" style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', fontSize: '0.8rem', borderCollapse: 'collapse' }}>
                <thead>
                    <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                        <th style={{ padding: '0.5rem', textAlign: 'left' }}>Hole</th>
                        {HOLE_NUMBERS.map(i => (
                            <th key={i} style={{ padding: '0.5rem', textAlign: 'center' }}>{i}</th>
                        ))}
                        <th style={{ padding: '0.5rem', textAlign: 'center', fontWeight: 'bold' }}>Total</th>
                    </tr>
                </thead>
                <tbody>
                    {rows.map((row, index) => (
                        <ScoreRow key={index} {...row} />
                    ))}
                </tbody>
            </table>
        </div>
    )
}

// Expandable Card Component
export interface ExpandableCardProps {
    isExpanded: boolean
    onToggle: () => void
    header: ReactNode
    rightContent: ReactNode
    children: ReactNode
}

export function ExpandableCard({ isExpanded, onToggle, header, rightContent, children }: ExpandableCardProps) {
    return (
        <div style={{ 
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-md)',
            overflow: 'hidden'
        }}>
            <button
                onClick={onToggle}
                style={{
                    width: '100%',
                    padding: 'var(--spacing-md)',
                    background: 'rgba(255, 255, 255, 0.02)',
                    border: 'none',
                    cursor: 'pointer',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    color: 'var(--color-text)'
                }}
            >
                {header}
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-lg)' }}>
                    {rightContent}
                    {isExpanded ? <ChevronUp className="w-5 h-5" /> : <ChevronDown className="w-5 h-5" />}
                </div>
            </button>
            {isExpanded && (
                <div style={{ 
                    padding: 'var(--spacing-md)',
                    background: 'rgba(0, 0, 0, 0.2)',
                    borderTop: '1px solid var(--color-border)'
                }}>
                    {children}
                </div>
            )}
        </div>
    )
}

// Stat Item Component
export interface StatItemProps {
    label: string
    value: string | number
    color?: string
}

export function StatItem({ label, value, color }: StatItemProps) {
    return (
        <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>{label}</p>
            <p style={{ fontWeight: '600', color }}>{value}</p>
        </div>
    )
}

// Absent Badge Component
export interface AbsentBadgeProps {
    small?: boolean
}

export function AbsentBadge({ small = false }: AbsentBadgeProps) {
    return (
        <span 
            style={{ 
                fontSize: small ? '0.65rem' : '0.75rem', 
                backgroundColor: 'var(--color-warning)',
                color: '#000',
                padding: small ? '0.1rem 0.3rem' : '0.15rem 0.4rem',
                borderRadius: '3px',
                fontWeight: '600',
                textTransform: 'uppercase',
                letterSpacing: '0.5px'
            }}
        >
            Absent
        </span>
    )
}
