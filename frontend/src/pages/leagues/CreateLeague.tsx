import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../../contexts/LeagueContext';
import { api } from '../../lib/api';
import { Trophy, ArrowLeft } from 'lucide-react';

export default function CreateLeague() {
    const navigate = useNavigate();
    const { selectLeague, refreshLeagues } = useLeague();
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);
        setError(null);

        try {
            const league = await api.createLeague({
                name,
                description,
            });

            // Refresh user's leagues to include the new one
            await refreshLeagues();

            // Select the new league and navigate to dashboard
            selectLeague(league.id);
            navigate('/');
        } catch (err) {
            console.error('Failed to create league:', err);
            setError('Failed to create league. Please try again.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)', paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
            <div className="container animate-fade-in">
                <div style={{ maxWidth: '600px', margin: '0 auto' }}>
                    <div style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <button
                            onClick={() => navigate('/leagues')}
                            className="btn btn-secondary"
                            style={{ display: 'inline-flex', alignItems: 'center' }}
                        >
                            <ArrowLeft className="h-4 w-4 mr-2" />
                            Back to Leagues
                        </button>
                    </div>

                    <div className="card">
                        <div style={{ textAlign: 'center', marginBottom: 'var(--spacing-2xl)' }}>
                            <div style={{ 
                                display: 'inline-flex', 
                                padding: 'var(--spacing-lg)', 
                                background: 'rgba(16, 185, 129, 0.2)', 
                                borderRadius: 'var(--radius-full)',
                                marginBottom: 'var(--spacing-lg)'
                            }}>
                                <Trophy className="h-8 w-8" style={{ color: 'var(--color-accent)' }} />
                            </div>
                            <h1 style={{ marginBottom: 'var(--spacing-md)' }}>
                                Create a New League
                            </h1>
                            <p style={{ color: 'var(--color-text-secondary)' }}>
                                Start a new golf league and invite players
                            </p>
                        </div>

                        <form onSubmit={handleSubmit}>
                            {error && (
                                <div className="alert alert-error" style={{ marginBottom: 'var(--spacing-lg)' }}>
                                    <strong>Error:</strong> {error}
                                </div>
                            )}

                            <div className="form-group">
                                <label htmlFor="name" className="form-label">
                                    League Name
                                </label>
                                <input
                                    id="name"
                                    name="name"
                                    type="text"
                                    required
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    className="form-input"
                                    placeholder="e.g. Sunday Morning Golfers"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="description" className="form-label">
                                    Description
                                </label>
                                <textarea
                                    id="description"
                                    name="description"
                                    rows={3}
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    className="form-textarea"
                                    placeholder="Brief description of your league..."
                                />
                            </div>

                            <div style={{ marginTop: 'var(--spacing-xl)' }}>
                                <button
                                    type="submit"
                                    disabled={isSubmitting}
                                    className="btn btn-primary"
                                    style={{ width: '100%' }}
                                >
                                    {isSubmitting ? 'Creating...' : 'Create League'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
}
