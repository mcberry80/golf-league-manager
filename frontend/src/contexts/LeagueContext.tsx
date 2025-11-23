import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
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

export function LeagueProvider({ children }: { children: React.ReactNode }) {
    const { getToken, userId } = useAuth();
    const [currentLeague, setCurrentLeague] = useState<League | null>(null);
    const [userRole, setUserRole] = useState<'admin' | 'player' | null>(null);
    const [userLeagues, setUserLeagues] = useState<LeagueMember[]>([]);
    const [isLoading, setIsLoading] = useState(true);

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
            const league = await api.getLeague(leagueId);
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
            if (!userId) {
                setIsLoading(false);
                return;
            }

            try {
                // Get current user info
                const userInfo = await api.getCurrentUser();

                // If user has a linked player, get their leagues
                if (userInfo.linked && userInfo.player) {
                    const leagues = await api.listLeagues();
                    const leagueMembers: LeagueMember[] = [];

                    // Fetch membership info for each league
                    for (const league of leagues) {
                        try {
                            const members = await api.listLeagueMembers(league.id);
                            const userMember = members.find(m => m.playerId === userInfo.player!.id);
                            if (userMember) {
                                leagueMembers.push(userMember);
                            }
                        } catch (err) {
                            console.error(`Failed to load members for league ${league.id}:`, err);
                        }
                    }

                    setUserLeagues(leagueMembers);

                    // Restore selected league from local storage if available
                    const savedLeagueId = localStorage.getItem('selectedLeagueId');
                    if (savedLeagueId && leagueMembers.some(m => m.leagueId === savedLeagueId)) {
                        // Call selectLeague inline to avoid circular dependency
                        const selectLeagueInline = async (leagueId: string, membersList: LeagueMember[]) => {
                            setIsLoading(true);
                            try {
                                const league = await api.getLeague(leagueId);
                                setCurrentLeague(league);
                                localStorage.setItem('selectedLeagueId', leagueId);
                                const member = membersList.find(l => l.leagueId === leagueId);
                                setUserRole(member?.role || null);
                            } catch (error) {
                                console.error('Failed to select league:', error);
                            } finally {
                                setIsLoading(false);
                            }
                        };
                        await selectLeagueInline(savedLeagueId, leagueMembers);
                    } else if (leagueMembers.length > 0) {
                        // Default to first league
                        const selectLeagueInline = async (leagueId: string, membersList: LeagueMember[]) => {
                            setIsLoading(true);
                            try {
                                const league = await api.getLeague(leagueId);
                                setCurrentLeague(league);
                                localStorage.setItem('selectedLeagueId', leagueId);
                                const member = membersList.find(l => l.leagueId === leagueId);
                                setUserRole(member?.role || null);
                            } catch (error) {
                                console.error('Failed to select league:', error);
                            } finally {
                                setIsLoading(false);
                            }
                        };
                        await selectLeagueInline(leagueMembers[0].leagueId, leagueMembers);
                    }
                }
            } catch (error) {
                console.error('Failed to load leagues:', error);
            } finally {
                setIsLoading(false);
            }
        };

        loadLeagues();
    }, [userId]);

    const refreshLeagues = async () => {
        if (!userId) return;

        try {
            const userInfo = await api.getCurrentUser();

            if (userInfo.linked && userInfo.player) {
                const leagues = await api.listLeagues();
                const leagueMembers: LeagueMember[] = [];

                // Fetch membership info for each league
                for (const league of leagues) {
                    try {
                        const members = await api.listLeagueMembers(league.id);
                        const userMember = members.find(m => m.playerId === userInfo.player!.id);
                        if (userMember) {
                            leagueMembers.push(userMember);
                        }
                    } catch (err) {
                        console.error(`Failed to load members for league ${league.id}:`, err);
                    }
                }

                setUserLeagues(leagueMembers);
            }
        } catch (error) {
            console.error('Failed to refresh leagues:', error);
        }
    };

    return (
        <LeagueContext.Provider value={{
            currentLeague,
            userRole,
            userLeagues,
            isLoading,
            selectLeague,
            refreshLeagues
        }}>
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
