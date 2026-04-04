/// <reference types="vite/client" />

import 'axios'

declare module 'axios' {
  interface AxiosRequestConfig {
    /** When true, the shared API client does not show an error toast (caller handles UX). */
    skipErrorToast?: boolean
  }
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}