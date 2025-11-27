import React, { createContext, useContext, useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { League, LeagueMember } from '../types';
import { api } from '../lib/api';
import { useAuth } from '@clerk/clerk-react';

interface LeagueContextType {
    currentLeague: League | null;
    userRole: 'admin' | 'player' | null;
    userLeagues: LeagueMember[];
    isLoading: boolean;
    selectLeague: (leagueId: string, membersList?: LeagueMember[]) => void;
    refreshLeagues: () => Promise<void>;
}

const LeagueContext = createContext<LeagueContextType | undefined>(undefined);

// Helper to batch fetch league members in parallel with a limit
const BATCH_SIZE = 5;

async function fetchLeagueMembersInBatches(
    leagues: League[],
    playerId: string
): Promise<LeagueMember[]> {
    const leagueMembers: LeagueMember[] = [];
    
    // Process in batches to avoid overwhelming the server
    for (let i = 0; i < leagues.length; i += BATCH_SIZE) {
        const batch = leagues.slice(i, i + BATCH_SIZE);
        const batchResults = await Promise.all(
            batch.map(async (league) => {
                try {
                    const members = await api.listLeagueMembers(league.id);
                    return members.find(m => m.playerId === playerId) || null;
                } catch (err) {
                    console.error(`Failed to load members for league ${league.id}:`, err);
                    return null;
                }
            })
        );
        
        // Add non-null results to leagueMembers
        for (const result of batchResults) {
            if (result) {
                leagueMembers.push(result);
            }
        }
    }
    
    return leagueMembers;
}

export function LeagueProvider({ children }: { children: React.ReactNode }) {
    const { getToken, userId } = useAuth();
    const [currentLeague, setCurrentLeague] = useState<League | null>(null);
    const [userRole, setUserRole] = useState<'admin' | 'player' | null>(null);
    const [userLeagues, setUserLeagues] = useState<LeagueMember[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    
    // Ref to prevent duplicate loading on re-renders
    const loadingRef = useRef(false);
    // Cache leagues to avoid re-fetching
    const leaguesCacheRef = useRef<Map<string, League>>(new Map());

    // Set up auth token provider on mount
    useEffect(() => {
        if (userId && getToken) {
            // Set dynamic token provider that fetches fresh token on each request
            api.setAuthTokenProvider(getToken);
        }
    }, [userId, getToken]);

    const selectLeague = useCallback(async (leagueId: string, membersList?: LeagueMember[]) => {
        setIsLoading(true);
        try {
            // Check cache first
            let league = leaguesCacheRef.current.get(leagueId);
            if (!league) {
                league = await api.getLeague(leagueId);
                leaguesCacheRef.current.set(leagueId, league);
            }
            
            setCurrentLeague(league);
            localStorage.setItem('selectedLeagueId', leagueId);

            // Set user role for this league
            let currentMembers = membersList || userLeagues;

            // If we don't have members list and userLeagues is empty, fetch it
            if (currentMembers.length === 0) {
                try {
                    const userInfo = await api.getCurrentUser();
                    if (userInfo.linked && userInfo.player) {
                        const members = await api.listLeagueMembers(leagueId);
                        const userMember = members.find(m => m.playerId === userInfo.player!.id);
                        if (userMember) {
                            currentMembers = [userMember];
                        }
                    }
                } catch (err) {
                    console.error('Failed to fetch membership info:', err);
                }
            }

            const member = currentMembers.find(l => l.leagueId === leagueId);
            setUserRole(member?.role || null);
        } catch (error) {
            console.error('Failed to select league:', error);
        } finally {
            setIsLoading(false);
        }
    }, [userLeagues]);

    // Load user's leagues on mount or auth change
    useEffect(() => {
        const loadLeagues = async () => {
            if (!userId || loadingRef.current) {
                setIsLoading(false);
                return;
            }
            
            loadingRef.current = true;

            try {
                // Get current user info
                const userInfo = await api.getCurrentUser();

                // If user has a linked player, get their leagues
                if (userInfo.linked && userInfo.player) {
                    const leagues = await api.listLeagues();
                    
                    // Cache all leagues
                    for (const league of leagues) {
                        leaguesCacheRef.current.set(league.id, league);
                    }
                    
                    // Fetch membership info for all leagues in parallel batches
                    const leagueMembers = await fetchLeagueMembersInBatches(leagues, userInfo.player.id);

                    setUserLeagues(leagueMembers);

                    // Restore selected league from local storage if available
                    const savedLeagueId = localStorage.getItem('selectedLeagueId');
                    if (savedLeagueId && leagueMembers.some(m => m.leagueId === savedLeagueId)) {
                        // Use cached league
                        const league = leaguesCacheRef.current.get(savedLeagueId);
                        if (league) {
                            setCurrentLeague(league);
                            const member = leagueMembers.find(l => l.leagueId === savedLeagueId);
                            setUserRole(member?.role || null);
                        }
                    } else if (leagueMembers.length > 0) {
                        // Default to first league, use cached league
                        const league = leaguesCacheRef.current.get(leagueMembers[0].leagueId);
                        if (league) {
                            setCurrentLeague(league);
                            localStorage.setItem('selectedLeagueId', leagueMembers[0].leagueId);
                            setUserRole(leagueMembers[0].role || null);
                        }
                    }
                }
            } catch (error) {
                console.error('Failed to load leagues:', error);
            } finally {
                setIsLoading(false);
                loadingRef.current = false;
            }
        };

        loadLeagues();
    }, [userId]);

    const refreshLeagues = useCallback(async () => {
        if (!userId) return;

        try {
            const userInfo = await api.getCurrentUser();

            if (userInfo.linked && userInfo.player) {
                const leagues = await api.listLeagues();
                
                // Update cache
                for (const league of leagues) {
                    leaguesCacheRef.current.set(league.id, league);
                }
                
                // Fetch membership info for all leagues in parallel batches
                const leagueMembers = await fetchLeagueMembersInBatches(leagues, userInfo.player.id);

                setUserLeagues(leagueMembers);
            }
        } catch (error) {
            console.error('Failed to refresh leagues:', error);
        }
    }, [userId]);

    // Memoize context value to prevent unnecessary re-renders
    const contextValue = useMemo(() => ({
        currentLeague,
        userRole,
        userLeagues,
        isLoading,
        selectLeague,
        refreshLeagues
    }), [currentLeague, userRole, userLeagues, isLoading, selectLeague, refreshLeagues]);

    return (
        <LeagueContext.Provider value={contextValue}>
            {children}
        </LeagueContext.Provider>
    );
}

export function useLeague() {
    const context = useContext(LeagueContext);
    if (context === undefined) {
        throw new Error('useLeague must be used within a LeagueProvider');
    }
    return context;
}
