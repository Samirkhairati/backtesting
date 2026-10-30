import { createFileRoute, redirect } from '@tanstack/react-router'
import { Dashboard } from '@/components/experiments/Dashboard'
import { useAuth } from '@/lib/auth'

export const Route = createFileRoute('/dashboard')({
  beforeLoad: () => {
    if (!useAuth.getState().token) {
      throw redirect({ to: '/' })
    }
  },
  component: Dashboard,
})
