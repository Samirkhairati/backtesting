import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { ArrowRight, Cpu, Database, LineChart, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AuthCard, type AuthMode } from '@/components/AuthCard'
import { EquityCurve } from '@/components/EquityCurve'
import { GithubIcon, GITHUB_URL } from '@/components/GithubIcon'
import { useAuth } from '@/lib/auth'

export const Route = createFileRoute('/')({
  validateSearch: (search: Record<string, unknown>): { mode?: AuthMode } => ({
    mode: search.mode === 'register' ? 'register' : search.mode === 'login' ? 'login' : undefined,
  }),
  component: Index,
})

const FEATURES = [
  { icon: Database, text: 'Historical data ingestion with smart caching' },
  { icon: Cpu, text: 'Parallel distributed backtesting workers' },
  { icon: LineChart, text: 'Equity curves, Sharpe, drawdown and more' },
]

function Index() {
  const token = useAuth((s) => s.token)
  const { mode = 'login' } = Route.useSearch()
  const navigate = useNavigate({ from: '/' })

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-1 flex-col">
      <div className="grid flex-1 items-center gap-10 py-8 lg:grid-cols-2 lg:gap-16 lg:py-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
        <section className="space-y-8">
          <div className="space-y-5">
            <p className="font-mono text-xs uppercase tracking-[0.3em] text-primary">
              Algorithmic backtesting
            </p>
            <h1 className="text-4xl font-extrabold leading-tight tracking-tight sm:text-5xl lg:text-6xl">
              Backtest strategies across the <span className="text-primary glow-text">whole market.</span>
            </h1>
            <p className="max-w-xl font-mono text-sm leading-relaxed text-muted-foreground sm:text-base">
              A distributed engine that ingests historical data, runs your strategies in parallel across
              NASDAQ and S&amp;P symbols, and aggregates the results into one clear report.
            </p>
          </div>

          <ul className="space-y-3">
            {FEATURES.map(({ icon: Icon, text }) => (
              <li key={text} className="flex items-center gap-3 font-mono text-sm">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center border border-primary/30 bg-primary/10 text-primary">
                  <Icon className="h-4 w-4" aria-hidden="true" />
                </span>
                {text}
              </li>
            ))}
          </ul>

          <div className="relative max-w-xl border border-primary/20 bg-black/40 p-4 backdrop-blur-md">
            <div className="mb-2 flex items-center justify-between font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              <span>Portfolio equity</span>
              <span className="text-primary">+ illustrative</span>
            </div>
            <EquityCurve className="h-32 w-full sm:h-40" />
          </div>
        </section>

        <section className="mx-auto w-full max-w-md lg:max-w-none lg:justify-self-end lg:w-[26rem]">
          {token ? (
            <Card className="border-primary/20 bg-black/40 backdrop-blur-md">
              <CardHeader className="space-y-1">
                <CardTitle className="text-2xl font-bold tracking-tight">Welcome back</CardTitle>
                <CardDescription className="font-mono text-xs">
                  You are signed in. Pick up where you left off.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <Button asChild className="w-full font-mono uppercase tracking-widest">
                  <Link to="/dashboard">
                    Go to Dashboard
                    <ArrowRight className="ml-2 h-4 w-4" aria-hidden="true" />
                  </Link>
                </Button>
                <Button asChild variant="outline" className="w-full font-mono uppercase tracking-widest">
                  <Link to="/create">
                    <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                    New Experiment
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ) : (
            <AuthCard
              mode={mode}
              onModeChange={(m) => navigate({ search: { mode: m }, replace: true })}
            />
          )}
        </section>
      </div>

      <footer className="flex items-center justify-center gap-2 border-t border-primary/10 py-6 font-mono text-xs text-muted-foreground">
        <span>Built by Samir Khairati</span>
        <a
          href={GITHUB_URL}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="GitHub"
          className="inline-flex h-8 w-8 items-center justify-center transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <GithubIcon className="h-4 w-4" />
        </a>
      </footer>
    </div>
  )
}
