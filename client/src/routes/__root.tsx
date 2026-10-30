import { lazy, Suspense } from 'react'
import { createRootRoute, Outlet } from '@tanstack/react-router'
import { Navbar } from '@/components/Navbar'

// Devtools are for development only; the dynamic import is dropped from production builds
const TanStackRouterDevtools = import.meta.env.DEV
  ? lazy(() =>
      import('@tanstack/router-devtools').then((m) => ({
        default: m.TanStackRouterDevtools,
      })),
    )
  : () => null

export const Route = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/30">
      <div className="flex min-h-screen flex-col relative overflow-hidden">
        {/* Subtle orange glow in the background */}
        <div className="pointer-events-none absolute -top-40 left-1/2 -z-10 h-[400px] w-[800px] -translate-x-1/2 rounded-full bg-primary/10 blur-[100px]" />

        <Navbar />
        <main className="relative z-10 flex min-h-[calc(100vh-4rem)] flex-1 flex-col p-4 sm:p-6">
          <Outlet />
        </main>
      </div>
      <Suspense fallback={null}>
        <TanStackRouterDevtools />
      </Suspense>
    </div>
  ),
})
