import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Course } from '../../types'

export default function CourseManagement() {
    const { leagueId } = useParams<{ leagueId: string }>()
    const { currentLeague, userRole, isLoading: leagueLoading, selectLeague } = useLeague()
    const navigate = useNavigate()
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        par: 36,
        courseRating: 36.0,
        slopeRating: 113,
        holePars: [4, 4, 4, 4, 4, 4, 4, 4, 4],
        holeHandicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
    })

    const loadCourses = useCallback(async () => {
        if (!currentLeague) return

        try {
            const data = await api.listCourses(currentLeague.id)
            setCourses(data || [])
        } catch (error) {
            console.error('Failed to load courses:', error)
        } finally {
            setLoading(false)
        }
    }, [currentLeague])

    useEffect(() => {
        if (leagueId && (!currentLeague || currentLeague.id !== leagueId)) {
            selectLeague(leagueId)
        }
    }, [leagueId, currentLeague, selectLeague])

    useEffect(() => {
        if (!leagueLoading && !currentLeague && !leagueId) {
            navigate('/leagues')
            return
        }

        if (currentLeague) {
            loadCourses()
        }
    }, [currentLeague, leagueLoading, navigate, leagueId, loadCourses])

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.createCourse(currentLeague.id, formData)
            setShowForm(false)
            setFormData({
                name: '',
                par: 36,
                courseRating: 36.0,
                slopeRating: 113,
                holePars: [4, 4, 4, 4, 4, 4, 4, 4, 4],
                holeHandicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
            })
            loadCourses()
        } catch (error) {
            alert('Failed to create course: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!currentLeague || userRole !== 'admin') {
        return (
            <div className="container" style={{ paddingTop: 'var(--spacing-2xl)' }}>
                <div className="alert alert-error">
                    <strong>Access Denied:</strong> You must be an admin of {currentLeague?.name || 'this league'} to access this page.
                </div>
                <Link to="/" className="btn btn-secondary" style={{ marginTop: 'var(--spacing-lg)' }}>
                    Return Home
                </Link>
            </div>
        )
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to={`/leagues/${currentLeague.id}/admin`} style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Admin
                </Link>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <div>
                        <h1>Course Management</h1>
                        <p className="text-gray-400 mt-1">{currentLeague.name}</p>
                    </div>
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ Add Course'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Add New Course</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Course Name</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        required
                                        placeholder="Pine Valley"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Par</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={formData.par}
                                        onChange={(e) => setFormData({ ...formData, par: parseInt(e.target.value) })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Course Rating</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        className="form-input"
                                        value={formData.courseRating}
                                        onChange={(e) => setFormData({ ...formData, courseRating: parseFloat(e.target.value) })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Slope Rating</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={formData.slopeRating}
                                        onChange={(e) => setFormData({ ...formData, slopeRating: parseInt(e.target.value) })}
                                        required
                                    />
                                </div>
                            </div>

                            <div style={{ marginTop: 'var(--spacing-xl)', paddingTop: 'var(--spacing-xl)', borderTop: '1px solid var(--color-border)' }}>
                                <h4 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Hole Details (9 Holes)</h4>

                                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(9, 1fr)', gap: 'var(--spacing-md)', marginBottom: 'var(--spacing-lg)' }}>
                                    {[...Array(9)].map((_, i) => (
                                        <div key={i} style={{ textAlign: 'center' }}>
                                            <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: '600', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-xs)' }}>
                                                Hole {i + 1}
                                            </label>
                                        </div>
                                    ))}
                                </div>

                                <div style={{ marginBottom: 'var(--spacing-lg)' }}>
                                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text)', marginBottom: 'var(--spacing-sm)' }}>
                                        Par for Each Hole
                                    </label>
                                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(9, 1fr)', gap: 'var(--spacing-md)' }}>
                                        {formData.holePars.map((par, i) => (
                                            <input
                                                key={i}
                                                type="number"
                                                className="form-input"
                                                value={par}
                                                onChange={(e) => {
                                                    const newPars = [...formData.holePars];
                                                    newPars[i] = parseInt(e.target.value) || 3;
                                                    setFormData({ ...formData, holePars: newPars });
                                                }}
                                                min="3"
                                                max="5"
                                                required
                                                style={{ padding: '0.5rem', textAlign: 'center' }}
                                            />
                                        ))}
                                    </div>
                                </div>

                                <div>
                                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text)', marginBottom: 'var(--spacing-sm)' }}>
                                        Handicap Ranking (1=hardest, 9=easiest)
                                    </label>
                                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(9, 1fr)', gap: 'var(--spacing-md)' }}>
                                        {formData.holeHandicaps.map((handicap, i) => (
                                            <input
                                                key={i}
                                                type="number"
                                                className="form-input"
                                                value={handicap}
                                                onChange={(e) => {
                                                    const newHandicaps = [...formData.holeHandicaps];
                                                    newHandicaps[i] = parseInt(e.target.value) || 1;
                                                    setFormData({ ...formData, holeHandicaps: newHandicaps });
                                                }}
                                                min="1"
                                                max="9"
                                                required
                                                style={{ padding: '0.5rem', textAlign: 'center' }}
                                            />
                                        ))}
                                    </div>
                                </div>
                            </div>

                            <button type="submit" className="btn btn-success" style={{ marginTop: 'var(--spacing-xl)' }}>
                                Add Course
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Courses ({courses.length})</h3>
                    {courses.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No courses added yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Par</th>
                                        <th>Course Rating</th>
                                        <th>Slope Rating</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {courses.map((course) => (
                                        <tr key={course.id}>
                                            <td style={{ fontWeight: '600' }}>{course.name}</td>
                                            <td>{course.par}</td>
                                            <td>{course.courseRating.toFixed(1)}</td>
                                            <td>{course.slopeRating}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
