import { createFileRoute, Link } from '@tanstack/react-router'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useState } from 'react'

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
import { Database } from 'lucide-react'

export const Route = createFileRoute('/register')({
  component: Register,
})

const formSchema = z.object({
  username: z.string().min(3, {
    message: 'Username must be at least 3 characters.',
  }),
  password: z.string().min(6, {
    message: 'Password must be at least 6 characters.',
  }),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

function Register() {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      password: '',
      confirmPassword: '',
    },
  })

  const [isLoading, setIsLoading] = useState(false)

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    setIsLoading(true)
    // Simulate API call for now
    setTimeout(() => {
      console.log(values)
      setIsLoading(false)
      // Wait for API to return token, then store and redirect
      // router.navigate({ to: '/' })
    }, 1500)
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] p-4 animate-in fade-in zoom-in-95 duration-700">
      <Card className="border-primary/20">
        <CardHeader className="space-y-1">
          <div className="flex items-center gap-2 mb-2">
            <Database className="h-6 w-6 text-primary glow-text" />
            <CardTitle className="text-2xl font-bold tracking-tight">Register</CardTitle>
          </div>
          <CardDescription className="font-mono text-xs">
            Create a new account.
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
              <FormField
                control={form.control}
                name="confirmPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Confirm Password</FormLabel>
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
                      Creating...
                    </div>
                  ) : (
                    "Create Account"
                  )}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
        <CardFooter className="flex justify-center border-t border-border/50 pt-6">
          <div className="text-sm text-muted-foreground font-mono">
            Already have an account?{' '}
            <Link to="/login" className="text-primary hover:underline underline-offset-4">
              Login here
            </Link>
          </div>
        </CardFooter>
      </Card>
    </div>
  )
}
