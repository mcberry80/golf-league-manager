import { ReactNode } from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'

// Constants
export const HOLE_NUMBERS = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const
export const EMPTY_STROKES_ARRAY = Array(9).fill(0) as number[]

// Score Row Component
export interface ScoreRowProps {
    label: string
    scores: number[]
    total: number | string
    color?: string
    bgColor?: string
    withBorder?: boolean
}

export function ScoreRow({ label, scores, total, color, bgColor, withBorder = false }: ScoreRowProps) {
    return (
        <tr style={{ 
            borderBottom: withBorder ? '1px solid var(--color-border)' : undefined, 
            background: bgColor 
        }}>
            <td style={{ padding: '0.5rem', color: color || 'var(--color-text-muted)' }}>{label}</td>
            {scores.map((score, i) => (
                <td key={i} style={{ padding: '0.5rem', textAlign: 'center', color }}>{score}</td>
            ))}
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
