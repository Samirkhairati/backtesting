import { createRootRoute, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/router-devtools'

export const Route = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/30">
      <div className="flex min-h-screen flex-col relative overflow-hidden">
        {/* Subtle orange glow in the background */}
        <div className="pointer-events-none absolute -top-40 left-1/2 -z-10 h-[400px] w-[800px] -translate-x-1/2 rounded-full bg-primary/10 blur-[100px]" />
        
        <main className="flex-1 flex flex-col items-center justify-center relative z-10 p-6">
          <Outlet />
        </main>
      </div>
      <TanStackRouterDevtools />
    </div>
  ),
})
