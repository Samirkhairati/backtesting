import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Terminal, Database } from 'lucide-react'

import { login, register } from '@/lib/api'
import { useAuth } from '@/lib/auth'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

export type AuthMode = 'login' | 'register'

const loginSchema = z.object({
  username: z.string().min(3, {
    message: 'Username must be at least 3 characters.',
  }),
  password: z.string().min(1, {
    message: 'Password is required.',
  }),
})

const registerSchema = z
  .object({
    username: z.string().min(3, {
      message: 'Username must be at least 3 characters.',
    }),
    password: z.string().min(1, {
      message: 'Password is required.',
    }),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  })

const labelClass = 'font-mono uppercase text-xs text-muted-foreground'

function Spinner({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-2">
      <div className="h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent" />
      {label}
    </div>
  )
}

function ErrorMessage({ message }: { message: string }) {
  if (!message) return null
  return (
    <div role="alert" className="mt-2 text-center font-mono text-xs text-red-500">
      {message}
    </div>
  )
}

function useAuthSubmit() {
  const navigate = useNavigate()
  const setToken = useAuth((s) => s.setToken)
  const [isLoading, setIsLoading] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  const run = async (action: () => Promise<{ token?: string }>, noTokenMessage?: string) => {
    setIsLoading(true)
    setErrorMsg('')
    try {
      const data = await action()
      if (data.token) {
        setToken(data.token)
        navigate({ to: '/dashboard' })
      } else if (noTokenMessage) {
        setErrorMsg(noTokenMessage)
      }
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : 'An unexpected error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return { isLoading, errorMsg, run }
}

function LoginForm() {
  const form = useForm<z.infer<typeof loginSchema>>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '' },
  })
  const { isLoading, errorMsg, run } = useAuthSubmit()

  const onSubmit = (values: z.infer<typeof loginSchema>) =>
    run(
      () => login({ username: values.username, password: values.password }),
      'Login failed: No token received',
    )

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel className={labelClass}>Username</FormLabel>
              <FormControl>
                <Input placeholder="admin" autoComplete="username" disabled={isLoading} {...field} />
              </FormControl>
              <FormMessage className="font-mono text-xs" />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel className={labelClass}>Password</FormLabel>
              <FormControl>
                <Input type="password" placeholder="••••••••" autoComplete="current-password" disabled={isLoading} {...field} />
              </FormControl>
              <FormMessage className="font-mono text-xs" />
            </FormItem>
          )}
        />
        <div className="pt-2">
          <Button type="submit" disabled={isLoading} className="w-full font-mono uppercase tracking-widest">
            {isLoading ? <Spinner label="Authenticating..." /> : 'Login'}
          </Button>
        </div>
        <ErrorMessage message={errorMsg} />
      </form>
    </Form>
  )
}

function RegisterForm() {
  const form = useForm<z.infer<typeof registerSchema>>({
    resolver: zodResolver(registerSchema),
    defaultValues: { username: '', password: '', confirmPassword: '' },
  })
  const { isLoading, errorMsg, run } = useAuthSubmit()

  const onSubmit = (values: z.infer<typeof registerSchema>) =>
    run(async () => {
      // First register the user, then automatically log them in
      await register({ username: values.username, password: values.password })
      return login({ username: values.username, password: values.password })
    })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel className={labelClass}>Username</FormLabel>
              <FormControl>
                <Input placeholder="admin" autoComplete="username" disabled={isLoading} {...field} />
              </FormControl>
              <FormMessage className="font-mono text-xs" />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel className={labelClass}>Password</FormLabel>
              <FormControl>
                <Input type="password" placeholder="••••••••" autoComplete="new-password" disabled={isLoading} {...field} />
              </FormControl>
              <FormMessage className="font-mono text-xs" />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="confirmPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel className={labelClass}>Confirm Password</FormLabel>
              <FormControl>
                <Input type="password" placeholder="••••••••" autoComplete="new-password" disabled={isLoading} {...field} />
              </FormControl>
              <FormMessage className="font-mono text-xs" />
            </FormItem>
          )}
        />
        <div className="pt-2">
          <Button type="submit" disabled={isLoading} className="w-full font-mono uppercase tracking-widest">
            {isLoading ? <Spinner label="Creating..." /> : 'Create Account'}
          </Button>
        </div>
        <ErrorMessage message={errorMsg} />
      </form>
    </Form>
  )
}

export function AuthCard({
  mode,
  onModeChange,
}: {
  mode: AuthMode
  onModeChange: (mode: AuthMode) => void
}) {
  const modes: { id: AuthMode; label: string }[] = [
    { id: 'login', label: 'Login' },
    { id: 'register', label: 'Register' },
  ]
  const Icon = mode === 'login' ? Terminal : Database

  return (
    <Card className="w-full border-primary/20 bg-black/40 backdrop-blur-md">
      <CardHeader className="space-y-4">
        <div role="tablist" aria-label="Authentication mode" className="grid grid-cols-2 border border-primary/20">
          {modes.map((m) => (
            <button
              key={m.id}
              type="button"
              role="tab"
              id={`auth-tab-${m.id}`}
              aria-selected={mode === m.id}
              aria-controls="auth-panel"
              onClick={() => onModeChange(m.id)}
              className={cn(
                'h-10 font-mono text-sm uppercase tracking-widest transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset',
                mode === m.id
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-primary/10 hover:text-foreground',
              )}
            >
              {m.label}
            </button>
          ))}
        </div>
        <div className="space-y-1 pt-2">
          <div className="flex items-center gap-2">
            <Icon className="h-6 w-6 text-primary" aria-hidden="true" />
            <CardTitle className="text-2xl font-bold tracking-tight">
              {mode === 'login' ? 'Welcome back' : 'Create account'}
            </CardTitle>
          </div>
          <CardDescription className="font-mono text-xs">
            {mode === 'login' ? 'Enter your credentials to continue.' : 'Sign up to start running backtests.'}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent id="auth-panel" role="tabpanel" aria-labelledby={`auth-tab-${mode}`}>
        {mode === 'login' ? <LoginForm key="login" /> : <RegisterForm key="register" />}
      </CardContent>
      <CardFooter className="justify-center border-t border-border/50 pt-6">
        <p className="font-mono text-sm text-muted-foreground">
          {mode === 'login' ? "Don't have an account? " : 'Already have an account? '}
          <button
            type="button"
            onClick={() => onModeChange(mode === 'login' ? 'register' : 'login')}
            className="text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            {mode === 'login' ? 'Register here' : 'Login here'}
          </button>
        </p>
      </CardFooter>
    </Card>
  )
}
