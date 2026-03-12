import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  locale: {
    locale: 'ar',
    rtl: { ar: true }
  },
  theme: {
    defaultTheme: 'mohesr',
    themes: {
      mohesr: {
        dark: false,
        colors: {
          primary: '#1a237e',
          secondary: '#ffd600',
          accent: '#0d47a1',
          error: '#d32f2f',
          success: '#388e3c',
          warning: '#f57c00',
          info: '#1976d2',
          background: '#f5f5f5',
          surface: '#ffffff',
        }
      }
    }
  },
  defaults: {
    VTextField: { variant: 'outlined', density: 'comfortable' },
    VSelect: { variant: 'outlined', density: 'comfortable' },
    VTextarea: { variant: 'outlined', density: 'comfortable' },
    VBtn: { variant: 'elevated' },
    VCard: { elevation: 2 },
  }
})
