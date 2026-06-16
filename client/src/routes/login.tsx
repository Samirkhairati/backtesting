import { createFileRoute, Link } from '@tanstack/react-router'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useState } from 'react'
import { login } from '@/lib/api'
import { useNavigate } from '@tanstack/react-router'

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
import { Terminal } from 'lucide-react'

export const Route = createFileRoute('/login')({
  component: Login,
})

const formSchema = z.object({
  username: z.string().min(3, {
    message: 'Username must be at least 3 characters.',
  }),
  password: z.string().min(6, {
    message: 'Password must be at least 6 characters.',
  }),
})

function Login() {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  const navigate = useNavigate()
  const [isLoading, setIsLoading] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    setIsLoading(true)
    setErrorMsg('')
    try {
      const data = await login({ username: values.username, password: values.password })
      if (data.token) {
        localStorage.setItem('token', data.token)
        navigate({ to: '/' })
      } else {
        setErrorMsg('Login failed: No token received')
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        setErrorMsg(err.message)
      } else {
        setErrorMsg('An unexpected error occurred')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] p-4 animate-in fade-in zoom-in-95 duration-700">
      <Card className="border-primary/20">
        <CardHeader className="space-y-1">
          <div className="flex items-center gap-2 mb-2">
            <Terminal className="h-6 w-6 text-primary glow-text" />
            <CardTitle className="text-2xl font-bold tracking-tight">Login</CardTitle>
          </div>
          <CardDescription className="font-mono text-xs">
            Enter your credentials to continue.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Username</FormLabel>
                    <FormControl>
                      <Input placeholder="admin" disabled={isLoading} {...field} />
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
                    <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="••••••••" disabled={isLoading} {...field} />
                    </FormControl>
                    <FormMessage className="font-mono text-xs" />
                  </FormItem>
                )}
              />
              <div className="pt-2">
                <Button type="submit" variant="glow" disabled={isLoading} className="w-full font-mono uppercase tracking-widest relative">
                  {isLoading ? (
                    <div className="flex items-center gap-2">
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent" />
                      Authenticating...
                    </div>
                  ) : (
                    "Login"
                  )}
                </Button>
              </div>
              {errorMsg && (
                <div className="text-red-500 font-mono text-xs text-center mt-2">
                  {errorMsg}
                </div>
              )}
            </form>
          </Form>
        </CardContent>
        <CardFooter className="flex justify-center border-t border-border/50 pt-6">
          <div className="text-sm text-muted-foreground font-mono">
            Don't have an account?{' '}
            <Link to="/register" className="text-primary hover:underline underline-offset-4">
              Register here
            </Link>
          </div>
        </CardFooter>
      </Card>
    </div>
  )
}
