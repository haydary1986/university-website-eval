import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', name: 'login', component: () => import('../views/LoginView.vue'), meta: { guest: true } },
  { path: '/change-password', name: 'change-password', component: () => import('../views/ChangePasswordView.vue'), meta: { auth: true } },
  {
    path: '/',
    component: () => import('../layouts/DefaultLayout.vue'),
    meta: { auth: true },
    children: [
      { path: 'dashboard', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
      { path: 'submissions', name: 'submissions', component: () => import('../views/SubmissionListView.vue') },
      { path: 'submissions/new', name: 'submission-new', component: () => import('../views/SubmissionFormView.vue'), meta: { role: 'university' } },
      { path: 'submissions/:id', name: 'submission-detail', component: () => import('../views/SubmissionDetailView.vue') },
      { path: 'submissions/:id/diff', name: 'submission-diff', component: () => import('../views/SubmissionDiffView.vue'), meta: { adminOnly: true } },
      { path: 'universities', name: 'universities', component: () => import('../views/UniversityListView.vue'), meta: { adminOnly: true } },
      { path: 'admin/review', name: 'admin-review', component: () => import('../views/AdminReviewView.vue'), meta: { adminOnly: true } },
      { path: 'admin/users', name: 'admin-users', component: () => import('../views/UserManagementView.vue'), meta: { superAdmin: true } },
      { path: 'admin/academic-years', name: 'academic-years', component: () => import('../views/AcademicYearView.vue'), meta: { superAdmin: true } },
      { path: 'admin/categories', name: 'admin-categories', component: () => import('../views/CategoryManagementView.vue'), meta: { superAdmin: true } },
      { path: 'admin/audit-logs', name: 'audit-logs', component: () => import('../views/AuditLogView.vue'), meta: { superAdmin: true } },
      { path: 'admin/security', name: 'security', component: () => import('../views/SecurityView.vue'), meta: { superAdmin: true } },
      { path: 'admin/settings', name: 'settings', component: () => import('../views/SettingsView.vue'), meta: { superAdmin: true } },
      { path: 'stats', name: 'stats', component: () => import('../views/StatsView.vue'), meta: { adminOnly: true } },
      { path: 'ai-analysis', name: 'ai-analysis', component: () => import('../views/AIAnalysisView.vue'), meta: { adminOnly: true } },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || 'null')

  if (to.meta.auth && !token) return next('/login')
  if (to.meta.guest && token) return next('/dashboard')

  // Force password change if required
  if (token && user?.must_change_password && to.name !== 'change-password') {
    return next('/change-password')
  }

  if (to.meta.superAdmin && user?.role !== 'super_admin') return next('/dashboard')
  if (to.meta.adminOnly && !['admin', 'super_admin'].includes(user?.role)) return next('/dashboard')
  next()
})

export default router
