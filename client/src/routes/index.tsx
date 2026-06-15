import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useQuery } from '@tanstack/react-query'
import { fetchExperiments } from '@/lib/api'
import { LogOut, Plus, Activity } from 'lucide-react'
import { useState } from 'react'

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

  // Logged in: Show Dashboard
  return <Dashboard onLogout={handleLogout} />
}

function Dashboard({ onLogout }: { onLogout: () => void }) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['experiments'],
    queryFn: fetchExperiments,
  })

  return (
    <div className="w-full max-w-5xl space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight glow-text text-primary flex items-center gap-2">
            <Activity className="h-8 w-8" />
            Dashboard
          </h1>
          <p className="text-muted-foreground font-mono text-sm">Manage your backtesting experiments.</p>
        </div>
        <div className="flex items-center gap-4">
          <Link to="/create">
            <Button variant="glow" className="font-mono flex items-center gap-2">
              <Plus className="h-4 w-4" />
              New Experiment
            </Button>
          </Link>
          <Button variant="outline" onClick={onLogout} className="font-mono text-muted-foreground hover:text-foreground">
            <LogOut className="h-4 w-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>

      <Card className="border-primary/20 bg-black/40 backdrop-blur-md">
        <CardHeader>
          <CardTitle className="font-mono uppercase tracking-widest text-sm text-primary">Your Experiments</CardTitle>
          <CardDescription>A list of all backtesting runs you have initiated.</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center p-8">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
            </div>
          ) : error ? (
            <div className="text-red-500 font-mono text-sm p-4 border border-red-500/20 rounded bg-red-500/10">
              Error loading experiments. Please try logging in again.
            </div>
          ) : !data?.experiments || data.experiments.length === 0 ? (
            <div className="text-center p-12 border border-dashed border-border/50 rounded-lg">
              <p className="text-muted-foreground font-mono mb-4">No experiments found.</p>
              <Link to="/create">
                <Button variant="outline" className="font-mono border-primary/50 text-primary">
                  Create your first experiment
                </Button>
              </Link>
            </div>
          ) : (
            <div className="rounded-md border border-border/50 overflow-hidden">
              <Table>
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead className="font-mono uppercase text-xs">Name</TableHead>
                    <TableHead className="font-mono uppercase text-xs">Status</TableHead>
                    <TableHead className="font-mono uppercase text-xs">Created At</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.experiments.map((exp: { id: string; name: string; description: string; status: string; created_at: string }) => (
                    <TableRow key={exp.id} className="hover:bg-primary/5 cursor-pointer transition-colors">
                      <TableCell className="font-medium">
                        <div>{exp.name}</div>
                        <div className="text-xs text-muted-foreground font-mono mt-1">{exp.description}</div>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                          exp.status === 'PENDING' ? 'bg-yellow-500/20 text-yellow-500' :
                          exp.status === 'RUNNING' ? 'bg-blue-500/20 text-blue-500' :
                          exp.status === 'COMPLETED' ? 'bg-green-500/20 text-green-500' :
                          'bg-red-500/20 text-red-500'
                        }`}>
                          {exp.status}
                        </span>
                      </TableCell>
                      <TableCell className="font-mono text-sm text-muted-foreground">
                        {new Date(exp.created_at).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
