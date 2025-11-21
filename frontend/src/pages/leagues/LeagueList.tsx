import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../../contexts/LeagueContext';
import { api } from '../../lib/api';
import { League, LeagueMember } from '../../types';
import { Plus, Trophy, Users, ArrowRight } from 'lucide-react';

export default function LeagueList() {
    const navigate = useNavigate();
    const { selectLeague, userLeagues } = useLeague();
    const [leagues, setLeagues] = useState<League[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const loadLeagues = async () => {
            try {
                const allLeagues = await api.listLeagues();
                setLeagues(allLeagues);
            } catch (error) {
                console.error('Failed to load leagues:', error);
            } finally {
                setIsLoading(false);
            }
        };

        loadLeagues();
    }, []);

    const handleSelectLeague = (leagueId: string) => {
        selectLeague(leagueId);
        navigate('/');
    };

    if (isLoading) {
        return (
            <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
                <div className="flex justify-center items-center min-h-screen">
                    <div className="spinner"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)', paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
            <div className="container animate-fade-in">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-2xl)' }}>
                    <div>
                        <h1>Your Leagues</h1>
                        <p style={{ marginTop: 'var(--spacing-sm)', color: 'var(--color-text-secondary)' }}>
                            Select a league to manage or view
                        </p>
                    </div>
                    <button
                        onClick={() => navigate('/leagues/create')}
                        className="btn btn-primary"
                    >
                        <Plus className="h-5 w-5 mr-2" />
                        Create New League
                    </button>
                </div>

                {leagues.length === 0 ? (
                    <div className="card-glass" style={{ textAlign: 'center', padding: 'var(--spacing-2xl)' }}>
                        <Trophy className="mx-auto h-12 w-12" style={{ color: 'var(--color-text-muted)', marginBottom: 'var(--spacing-md)' }} />
                        <h3 style={{ marginTop: 'var(--spacing-md)', color: 'var(--color-text)' }}>No leagues found</h3>
                        <p style={{ marginTop: 'var(--spacing-sm)', color: 'var(--color-text-secondary)' }}>
                            Get started by creating a new league.
                        </p>
                        <div style={{ marginTop: 'var(--spacing-xl)' }}>
                            <button
                                onClick={() => navigate('/leagues/create')}
                                className="btn btn-primary"
                            >
                                <Plus className="h-5 w-5 mr-2" />
                                Create League
                            </button>
                        </div>
                    </div>
                ) : (
                    <div className="grid" style={{ gap: 'var(--spacing-xl)' }}>
                        <style>{`
                            @media (min-width: 1024px) {
                                .league-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
                            }
                            @media (min-width: 640px) and (max-width: 1023px) {
                                .league-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
                            }
                            @media (max-width: 639px) {
                                .league-grid { grid-template-columns: repeat(1, minmax(0, 1fr)); }
                            }
                        `}</style>
                        <div className="league-grid" style={{ display: 'grid', gap: 'var(--spacing-xl)' }}>
                        {leagues.map((league) => {
                            const member = userLeagues.find((l: LeagueMember) => l.league_id === league.id);
                            const role = member?.role || 'none';

                            return (
                                <div
                                    key={league.id}
                                    className="card"
                                    style={{ cursor: 'pointer', textDecoration: 'none', color: 'inherit' }}
                                    onClick={() => handleSelectLeague(league.id)}
                                >
                                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--spacing-lg)' }}>
                                        <div style={{ 
                                            background: 'rgba(16, 185, 129, 0.2)', 
                                            borderRadius: 'var(--radius-full)', 
                                            padding: 'var(--spacing-md)' 
                                        }}>
                                            <Trophy className="h-6 w-6" style={{ color: 'var(--color-accent)' }} />
                                        </div>
                                        {role === 'admin' && (
                                            <span className="badge badge-primary">
                                                Admin
                                            </span>
                                        )}
                                    </div>
                                    <h3 style={{ 
                                        marginBottom: 'var(--spacing-sm)', 
                                        color: 'var(--color-text)',
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {league.name}
                                    </h3>
                                    <p style={{ 
                                        marginTop: 'var(--spacing-sm)', 
                                        color: 'var(--color-text-muted)', 
                                        fontSize: '0.875rem',
                                        minHeight: '2.5rem',
                                        overflow: 'hidden',
                                        display: '-webkit-box',
                                        WebkitLineClamp: 2,
                                        WebkitBoxOrient: 'vertical'
                                    }}>
                                        {league.description || 'No description provided'}
                                    </p>
                                    <div style={{ 
                                        marginTop: 'var(--spacing-lg)', 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        justifyContent: 'space-between',
                                        fontSize: '0.875rem',
                                        color: 'var(--color-text-muted)',
                                        paddingTop: 'var(--spacing-md)',
                                        borderTop: '1px solid var(--color-border)'
                                    }}>
                                        <div style={{ display: 'flex', alignItems: 'center' }}>
                                            <Users className="flex-shrink-0 mr-1.5 h-4 w-4" />
                                            <span>View League</span>
                                        </div>
                                        <ArrowRight className="h-4 w-4" />
                                    </div>
                                    <div style={{ 
                                        marginTop: 'var(--spacing-md)', 
                                        fontSize: '0.75rem', 
                                        color: 'var(--color-text-muted)' 
                                    }}>
                                        Created {new Date(league.created_at).toLocaleDateString()}
                                    </div>
                                </div>
                            );
                        })}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
