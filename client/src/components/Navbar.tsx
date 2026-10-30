import { Link, useNavigate } from '@tanstack/react-router'
import { Activity, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/lib/auth'
import { GithubIcon, GITHUB_URL } from '@/components/GithubIcon'

export function Navbar() {
  const token = useAuth((s) => s.token)
  const logout = useAuth((s) => s.logout)
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate({ to: '/' })
  }

  return (
    <header className="sticky top-0 z-40 h-16 w-full border-b border-primary/20 bg-background/70 backdrop-blur-md">
      <nav
        aria-label="Main"
        className="mx-auto flex h-full max-w-7xl items-center justify-between gap-2 px-4 sm:px-6"
      >
        <div className="flex min-w-0 flex-1 items-center gap-3 overflow-hidden sm:gap-6">
          <Link
            to="/"
            className="flex shrink-0 items-center gap-2 font-mono text-base font-bold tracking-tight focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:text-lg"
          >
            <Activity className="h-5 w-5 text-primary" aria-hidden="true" />
            <span>
              Backtest<span className="text-primary">ing</span>
            </span>
          </Link>
          {token && (
            <Link
              to="/dashboard"
              className="px-1 py-1 font-mono text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              activeProps={{
                className: 'text-primary hover:text-primary border-b border-primary',
              }}
            >
              Dashboard
            </Link>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-1 sm:gap-2">
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="GitHub"
            className="inline-flex h-10 w-10 items-center justify-center text-muted-foreground transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <GithubIcon className="h-5 w-5" />
          </a>
          {token && (
            <Button
              variant="ghost"
              onClick={handleLogout}
              aria-label="Logout"
              className="px-2 font-mono text-muted-foreground hover:text-foreground sm:px-4"
            >
              <LogOut className="h-4 w-4 sm:mr-2" aria-hidden="true" />
              <span className="hidden sm:inline">Logout</span>
            </Button>
          )}
        </div>
      </nav>
    </header>
  )
}
