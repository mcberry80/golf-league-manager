/**
 * Shared utility functions for the Golf League Manager frontend.
 * Centralizes common formatting and helper functions to avoid duplication.
 */

/**
 * Format a date string for display (short format: "Jan 1")
 */
export function formatDateShort(dateString: string): string {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric'
    })
}

/**
 * Format a date string for display (medium format: "Jan 1, 2024")
 */
export function formatDate(dateString: string): string {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    })
}

/**
 * Format a date string with weekday (full format: "Mon, Jan 1, 2024")
 */
export function formatDateWithWeekday(dateString: string): string {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { 
        weekday: 'short',
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    })
}

/**
 * Format a date for date inputs (YYYY-MM-DD) using UTC to avoid timezone issues
 */
export function formatDateOnly(dateString: string): string {
    const date = new Date(dateString)
    const year = date.getUTCFullYear()
    const month = date.getUTCMonth() + 1
    const day = date.getUTCDate()
    return `${month}/${day}/${year}`
}

/**
 * Format a relative time string (e.g., "5m ago", "2h ago", "3d ago")
 */
export function formatRelativeTime(dateString: string): string {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    
    return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    })
}

/**
 * Create a lookup map from an array by a key property
 */
export function createLookupMap<T extends Record<string, unknown>>(
    items: T[],
    keyProperty: keyof T
): Map<string, T> {
    const map = new Map<string, T>()
    for (const item of items) {
        const key = item[keyProperty] as string
        if (key) {
            map.set(key, item)
        }
    }
    return map
}
