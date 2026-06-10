import { createFileRoute, Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

export const Route = createFileRoute('/')({
  component: Index,
})

function Index() {
  return (
    <div className="flex flex-col items-center justify-center text-center space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
      <div className="space-y-4">
        <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight">
          Backtesting <span className="text-primary glow-text">Platform</span>
        </h1>
        <p className="text-muted-foreground text-lg md:text-xl max-w-[600px] font-mono">
          Login or register to access the backtesting dashboard.
        </p>
      </div>

      <div className="flex items-center gap-4">
        <Link to="/login">
          <Button variant="glow" size="lg" className="font-mono text-base px-8 h-12">
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
