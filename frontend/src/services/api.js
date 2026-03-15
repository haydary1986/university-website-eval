import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: { 'Content-Type': 'application/json' }
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

// Helper to unwrap { key: data } responses
function unwrap(promise, key) {
  return promise.then(res => {
    if (key && res.data && res.data[key] !== undefined) {
      res.data = res.data[key]
    }
    return res
  })
}

export default {
  // Auth
  login: (data) => api.post('/auth/login', data),
  getMe: () => unwrap(api.get('/auth/me'), 'user'),
  changePassword: (data) => api.post('/auth/change-password', data),
  logout: () => api.post('/auth/logout'),
  getActiveSessions: () => api.get('/auth/sessions'),

  // Audit Logs
  getAuditLogs: (params) => api.get('/admin/audit-logs', { params }),

  // Security
  getSecurityOverview: () => api.get('/admin/security/overview'),
  getLoginAttempts: (params) => api.get('/admin/security/login-attempts', { params }),
  blockUser: (id, data) => api.put(`/admin/security/users/${id}/block`, data),
  unblockUser: (id) => api.put(`/admin/security/users/${id}/unblock`),
  getBlockedIPs: () => api.get('/admin/security/blocked-ips'),
  blockIP: (data) => api.post('/admin/security/block-ip', data),
  unblockIP: (ip) => api.delete(`/admin/security/unblock-ip/${ip}`),
  getAllSessions: () => api.get('/admin/security/sessions'),
  terminateSession: (id) => api.delete(`/admin/security/sessions/${id}`),
  terminateUserSessions: (id) => api.delete(`/admin/security/users/${id}/sessions`),

  // Universities
  getUniversities: (params) => unwrap(api.get('/universities', { params }), 'universities'),
  getUniversity: (id) => unwrap(api.get(`/universities/${id}`), 'university'),
  updateUniversity: (id, data) => unwrap(api.put(`/universities/${id}`, data), 'university'),

  // Academic Years
  getAcademicYears: () => unwrap(api.get('/academic-years'), 'academic_years'),
  createAcademicYear: (data) => unwrap(api.post('/academic-years', data), 'academic_year'),
  updateAcademicYear: (id, data) => unwrap(api.put(`/academic-years/${id}`, data), 'academic_year'),

  // Categories
  getCategories: () => unwrap(api.get('/categories'), 'categories'),
  getCategory: (id) => unwrap(api.get(`/categories/${id}`), 'category'),
  createCategory: (data) => unwrap(api.post('/admin/categories', data), 'category'),
  updateCategory: (id, data) => unwrap(api.put(`/admin/categories/${id}`, data), 'category'),
  deleteCategory: (id) => api.delete(`/admin/categories/${id}`),
  createCriteria: (categoryId, data) => unwrap(api.post(`/admin/categories/${categoryId}/criteria`, data), 'criteria'),
  updateCriteria: (id, data) => unwrap(api.put(`/admin/criteria/${id}`, data), 'criteria'),
  deleteCriteria: (id) => api.delete(`/admin/criteria/${id}`),

  // Submissions
  getSubmissions: (params) => unwrap(api.get('/submissions', { params }), 'submissions'),
  getSubmission: (id) => unwrap(api.get(`/submissions/${id}`), 'submission'),
  createSubmission: (data) => unwrap(api.post('/submissions', data), 'submission'),
  updateSubmission: (id, data) => unwrap(api.put(`/submissions/${id}`, data), 'submission'),
  submitSubmission: (id) => api.post(`/submissions/${id}/submit`),
  getSubmissionDiff: (id, version) => api.get(`/submissions/${id}/diff/${version}`),

  // Admin
  getAdminSubmissions: (params) => unwrap(api.get('/admin/submissions', { params }), 'submissions'),
  getAdminSubmission: (id) => unwrap(api.get(`/admin/submissions/${id}`), 'submission'),
  reviewSubmission: (id, data) => api.post(`/admin/submissions/${id}/review`, data),
  approveSubmission: (id) => api.put(`/admin/submissions/${id}/approve`),
  rejectSubmission: (id, data) => api.put(`/admin/submissions/${id}/reject`, data),

  // Users
  getUsers: () => unwrap(api.get('/admin/users'), 'users'),
  createUser: (data) => unwrap(api.post('/admin/users', data), 'user'),
  updateUser: (id, data) => unwrap(api.put(`/admin/users/${id}`, data), 'user'),
  deleteUser: (id) => api.delete(`/admin/users/${id}`),
  assignCategories: (id, data) => api.put(`/admin/users/${id}/assign-categories`, data),

  // Stats
  getStatsOverview: (params) => api.get('/stats/overview', { params }),
  getStatsUniversities: (params) => unwrap(api.get('/stats/universities', { params }), 'rankings'),
  getStatsCategories: (params) => unwrap(api.get('/stats/categories', { params }), 'category_stats'),
  getCategoryRankings: (params) => api.get('/stats/category-rankings', { params }),
  getUniversityProfile: (id, params) => api.get(`/stats/university-profile/${id}`, { params }),
  getStatsComparison: (id, params) => api.get(`/stats/comparison/${id}`, { params }),

  // Settings
  getSettings: () => api.get('/admin/settings'),
  updateSettings: (data) => api.put('/admin/settings', data),
  testAI: (data) => api.post('/admin/settings/test-ai', data),
  getPublicSettings: () => axios.get((import.meta.env.VITE_API_URL || '/api') + '/settings/public'),

  // AI
  analyzeSubmission: (id, provider) => api.post(`/ai/analyze-submission/${id}`, { provider }),
  suggestImprovements: (id, provider) => api.post(`/ai/suggest-improvements/${id}`, { provider }),
  compareUniversities: (ids, provider) => api.post('/ai/compare-universities', { university_ids: ids, provider }),

  // Export
  exportRankings: (params) => api.get('/export/rankings', { params, responseType: 'blob' }),
  exportCategoryRankings: (params) => api.get('/export/category-rankings', { params, responseType: 'blob' }),
  exportSubmissions: (params) => api.get('/export/submissions', { params, responseType: 'blob' }),

  // File upload
  uploadFile: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post('/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  }
}
