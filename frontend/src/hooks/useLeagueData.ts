import { useState, useEffect, useCallback, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { useLeague } from '../contexts/LeagueContext'
import api from '../lib/api'
import type {
    Season,
    Match,
    Course,
    LeagueMemberWithPlayer,
    StandingsEntry,
    MatchDay,
    Score
} from '../types'

/**
 * Cache entry with timestamp for invalidation
 */
interface CacheEntry<T> {
    data: T
    timestamp: number
}

// Simple in-memory cache for API data
const cache = new Map<string, CacheEntry<unknown>>()
/**
 * Cache TTL set to 1 minute to balance freshness with reduced Firestore reads.
 * This duration works well for league data that changes infrequently during a session,
 * while still ensuring users see relatively recent data on page navigation.
 */
const CACHE_TTL = 60000

function getCachedData<T>(key: string): T | null {
    const entry = cache.get(key)
    if (entry && Date.now() - entry.timestamp < CACHE_TTL) {
        return entry.data as T
    }
    cache.delete(key)
    return null
}

function setCachedData<T>(key: string, data: T): void {
    cache.set(key, { data, timestamp: Date.now() })
}

/**
 * Custom hook for fetching league-specific data with caching and loading states.
 * Reduces boilerplate across components that need league data.
 */
export interface LeagueData {
    seasons: Season[]
    activeSeason: Season | null
    members: LeagueMemberWithPlayer[]
    courses: Course[]
    matches: Match[]
    matchDays: MatchDay[]
}

export interface UseLeagueDataResult extends LeagueData {
    loading: boolean
    error: string | null
    effectiveLeagueId: string | null
    refresh: () => Promise<void>
    // Helper functions that operate on loaded data
    getPlayerName: (playerId: string) => string
    getCourseName: (courseId: string) => string
    getCourse: (courseId: string) => Course | undefined
    getSeasonName: (seasonId: string) => string
    getMemberByPlayerId: (playerId: string) => LeagueMemberWithPlayer | undefined
}

interface UseLeagueDataOptions {
    /** Whether to load match data */
    loadMatches?: boolean
    /** Whether to load match days data */
    loadMatchDays?: boolean
    /** Whether to use cached data if available */
    useCache?: boolean
}

/**
 * Custom hook for loading common league data with configurable options.
 * Centralizes data fetching logic used across multiple components.
 */
export function useLeagueData(options: UseLeagueDataOptions = {}): UseLeagueDataResult {
    const { loadMatches = true, loadMatchDays = false, useCache = true } = options
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, isLoading: leagueLoading } = useLeague()
    
    const [seasons, setSeasons] = useState<Season[]>([])
    const [members, setMembers] = useState<LeagueMemberWithPlayer[]>([])
    const [courses, setCourses] = useState<Course[]>([])
    const [matches, setMatches] = useState<Match[]>([])
    const [matchDays, setMatchDays] = useState<MatchDay[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    
    const effectiveLeagueId = leagueId || currentLeague?.id || null
    
    // Use ref to track if we're already loading to prevent duplicate requests
    const loadingRef = useRef(false)

    const loadData = useCallback(async (bypassCache = false) => {
        if (!effectiveLeagueId || loadingRef.current) return
        
        const cacheKey = `league-data-${effectiveLeagueId}-${loadMatches}-${loadMatchDays}`
        
        // Check cache first
        if (useCache && !bypassCache) {
            const cached = getCachedData<LeagueData>(cacheKey)
            if (cached) {
                setSeasons(cached.seasons)
                setMembers(cached.members)
                setCourses(cached.courses)
                setMatches(cached.matches)
                setMatchDays(cached.matchDays)
                setLoading(false)
                return
            }
        }

        loadingRef.current = true
        setLoading(true)
        setError(null)

        try {
            // Build named promises for clarity
            const basePromises = {
                seasons: api.listSeasons(effectiveLeagueId),
                members: api.listLeagueMembers(effectiveLeagueId),
                courses: api.listCourses(effectiveLeagueId),
            }
            
            // Execute all base promises
            const [seasonsData, membersData, coursesData] = await Promise.all([
                basePromises.seasons,
                basePromises.members,
                basePromises.courses,
            ])
            
            // Conditionally fetch matches and match days
            let matchesData: Match[] = []
            let matchDaysData: MatchDay[] = []
            
            if (loadMatches || loadMatchDays) {
                const conditionalPromises: Promise<unknown>[] = []
                if (loadMatches) {
                    conditionalPromises.push(api.listMatches(effectiveLeagueId))
                }
                if (loadMatchDays) {
                    conditionalPromises.push(api.listMatchDays(effectiveLeagueId))
                }
                
                const conditionalResults = await Promise.all(conditionalPromises)
                
                let resultIndex = 0
                if (loadMatches) {
                    matchesData = conditionalResults[resultIndex++] as Match[]
                }
                if (loadMatchDays) {
                    matchDaysData = conditionalResults[resultIndex] as MatchDay[]
                }
            }

            setSeasons(seasonsData)
            setMembers(membersData)
            setCourses(coursesData)
            setMatches(matchesData)
            setMatchDays(matchDaysData)

            // Cache the data
            if (useCache) {
                setCachedData(cacheKey, {
                    seasons: seasonsData,
                    activeSeason: seasonsData.find(s => s.active) || null,
                    members: membersData,
                    courses: coursesData,
                    matches: matchesData,
                    matchDays: matchDaysData,
                })
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load league data')
        } finally {
            setLoading(false)
            loadingRef.current = false
        }
    }, [effectiveLeagueId, loadMatches, loadMatchDays, useCache])

    useEffect(() => {
        if (!leagueLoading && effectiveLeagueId) {
            loadData()
        }
    }, [loadData, leagueLoading, effectiveLeagueId])

    // Helper functions that use the loaded data
    const getPlayerName = useCallback((playerId: string): string => {
        const member = members.find(m => m.playerId === playerId)
        return member?.player?.name || 'Unknown'
    }, [members])

    const getCourseName = useCallback((courseId: string): string => {
        const course = courses.find(c => c.id === courseId)
        return course?.name || 'Unknown'
    }, [courses])

    const getCourse = useCallback((courseId: string): Course | undefined => {
        return courses.find(c => c.id === courseId)
    }, [courses])

    const getSeasonName = useCallback((seasonId: string): string => {
        const season = seasons.find(s => s.id === seasonId)
        return season?.name || 'Unknown'
    }, [seasons])

    const getMemberByPlayerId = useCallback((playerId: string): LeagueMemberWithPlayer | undefined => {
        return members.find(m => m.playerId === playerId)
    }, [members])

    const refresh = useCallback(async () => {
        await loadData(true)
    }, [loadData])

    const activeSeason = seasons.find(s => s.active) || null

    return {
        seasons,
        activeSeason,
        members,
        courses,
        matches,
        matchDays,
        loading: leagueLoading || loading,
        error,
        effectiveLeagueId,
        refresh,
        getPlayerName,
        getCourseName,
        getCourse,
        getSeasonName,
        getMemberByPlayerId,
    }
}

/**
 * Custom hook for fetching standings data
 */
export interface UseStandingsResult {
    standings: StandingsEntry[]
    loading: boolean
    error: string | null
    refresh: () => Promise<void>
}

export function useStandings(leagueId: string | null | undefined): UseStandingsResult {
    const [standings, setStandings] = useState<StandingsEntry[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const loadStandings = useCallback(async () => {
        if (!leagueId) {
            setLoading(false)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const data = await api.getStandings(leagueId)
            // Sort by total points descending
            const sorted = data.sort((a, b) => b.totalPoints - a.totalPoints)
            setStandings(sorted)
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load standings')
        } finally {
            setLoading(false)
        }
    }, [leagueId])

    useEffect(() => {
        loadStandings()
    }, [loadStandings])

    return {
        standings,
        loading,
        error,
        refresh: loadStandings,
    }
}

/**
 * Custom hook for fetching current user info
 */
export interface UseCurrentUserResult {
    player: { id: string; name: string; email: string } | null
    isLinked: boolean
    loading: boolean
}

export function useCurrentUser(): UseCurrentUserResult {
    const [player, setPlayer] = useState<{ id: string; name: string; email: string } | null>(null)
    const [isLinked, setIsLinked] = useState(false)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        async function loadCurrentUser() {
            try {
                const userInfo = await api.getCurrentUser()
                if (userInfo.linked && userInfo.player) {
                    setPlayer({
                        id: userInfo.player.id,
                        name: userInfo.player.name,
                        email: userInfo.player.email,
                    })
                    setIsLinked(true)
                }
            } catch {
                // User not linked yet, that's okay
            } finally {
                setLoading(false)
            }
        }
        loadCurrentUser()
    }, [])

    return { player, isLinked, loading }
}

/**
 * Custom hook for fetching match scores
 */
export function useMatchScores(
    leagueId: string | null | undefined,
    matchIds: string[]
): { 
    scoresMap: Map<string, Score[]>
    loading: boolean 
} {
    const [scoresMap, setScoresMap] = useState<Map<string, Score[]>>(new Map())
    const [loading, setLoading] = useState(false)
    
    // Track which matches we've already loaded
    const loadedMatchesRef = useRef<Set<string>>(new Set())

    useEffect(() => {
        if (!leagueId || matchIds.length === 0) return

        // Filter to only unloaded matches
        const unloadedMatchIds = matchIds.filter(id => !loadedMatchesRef.current.has(id))
        if (unloadedMatchIds.length === 0) return

        const loadScores = async () => {
            setLoading(true)
            const newScoresMap = new Map(scoresMap)
            const BATCH_SIZE = 5

            for (let i = 0; i < unloadedMatchIds.length; i += BATCH_SIZE) {
                const batch = unloadedMatchIds.slice(i, i + BATCH_SIZE)
                await Promise.all(
                    batch.map(async (matchId) => {
                        try {
                            const matchScores = await api.getMatchScores(leagueId, matchId)
                            newScoresMap.set(matchId, matchScores)
                            loadedMatchesRef.current.add(matchId)
                        } catch (err) {
                            console.warn(`Failed to load scores for match ${matchId}:`, err)
                        }
                    })
                )
            }

            setScoresMap(newScoresMap)
            setLoading(false)
        }

        loadScores()
    }, [leagueId, matchIds, scoresMap])

    return { scoresMap, loading }
}

/**
 * Invalidate all cached league data (call after mutations)
 */
export function invalidateLeagueCache(leagueId?: string): void {
    if (leagueId) {
        // Remove all cache entries for this league
        for (const key of cache.keys()) {
            if (key.includes(leagueId)) {
                cache.delete(key)
            }
        }
    } else {
        // Clear all cache
        cache.clear()
    }
}
