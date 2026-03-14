import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const token = ref(localStorage.getItem('token') || '')
  const mustChangePassword = ref(false)

  const isAuthenticated = computed(() => !!token.value)
  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')
  const isAdmin = computed(() => user.value?.role === 'admin' || user.value?.role === 'super_admin')
  const isUniversity = computed(() => user.value?.role === 'university')
  const userRole = computed(() => user.value?.role || '')

  async function login(username, password) {
    const res = await api.login({ username, password })
    token.value = res.data.token
    user.value = res.data.user
    mustChangePassword.value = res.data.must_change_password || false
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    return res.data
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchMe() {
    try {
      const res = await api.getMe()
      user.value = res.data
      localStorage.setItem('user', JSON.stringify(res.data))
    } catch {
      logout()
    }
  }

  return { user, token, mustChangePassword, isAuthenticated, isSuperAdmin, isAdmin, isUniversity, userRole, login, logout, fetchMe }
})
