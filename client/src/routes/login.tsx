import { createFileRoute, Link } from '@tanstack/react-router'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

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

  function onSubmit(values: z.infer<typeof formSchema>) {
    // This is just a mockup, so we just log
    console.log(values)
  }

  return (
    <div className="w-full max-w-md animate-in fade-in zoom-in-95 duration-500">
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
                      <Input placeholder="admin" {...field} />
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
                      <Input type="password" placeholder="••••••••" {...field} />
                    </FormControl>
                    <FormMessage className="font-mono text-xs" />
                  </FormItem>
                )}
              />
              <div className="pt-2">
                <Button type="submit" variant="glow" className="w-full font-mono uppercase tracking-widest">
                  Login
                </Button>
              </div>
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
