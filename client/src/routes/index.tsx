import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { useState } from 'react'
import { Dashboard } from '@/components/experiments/Dashboard'

export const Route = createFileRoute('/')({
  component: Index,
})

function Index() {
  const [token, setToken] = useState<string | null>(() => typeof window !== 'undefined' ? localStorage.getItem('token') : null)
  const navigate = useNavigate()

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
    navigate({ to: '/' })
  }

  // If not logged in, show Landing Page
  if (!token) {
    return (
      <div className="flex flex-col items-center justify-center text-center space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
        <div className="space-y-4">
          <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight">
            Backtesting <span className="text-primary">Platform</span>
          </h1>
          <p className="text-muted-foreground text-lg md:text-xl max-w-[600px] font-mono">
            Login or register to access the backtesting dashboard.
          </p>
        </div>

        <div className="flex items-center gap-4">
          <Link to="/login">
            <Button variant="default" size="lg" className="font-mono text-base px-8 h-12">
              Login
            </Button>
          </Link>
          <Link to="/register">
            <Button variant="outline" size="lg" className="font-mono text-base px-8 h-12 border-primary/50 text-primary hover:bg-primary/10">
              Register
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  // Logged in: Show Dashboard
  return <Dashboard onLogout={handleLogout} />
}
