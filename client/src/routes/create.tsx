import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { MultiSelect } from '@/components/ui/multi-select'
import { ArrowLeft, Rocket, Calendar as CalendarIcon } from 'lucide-react'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { createExperiment, searchSymbols } from '@/lib/api'
import React from 'react'

export const Route = createFileRoute('/create')({
  component: CreateExperiment,
})

const formSchema = z.object({
  name: z.string().min(3, 'Name must be at least 3 characters.'),
  description: z.string().optional(),
  symbols: z.array(z.string()).min(1, 'Please select at least one symbol.'),
  strategies: z.array(z.string()).min(1, 'Please select at least one strategy.'),
  startDate: z.string().min(1, 'Start date is required.'),
  endDate: z.string().min(1, 'End date is required.'),
}).refine(data => {
  const start = new Date(data.startDate);
  const end = new Date(data.endDate);
  return start < end;
}, {
  message: "End date must be after start date.",
  path: ["endDate"]
}).refine(data => {
  const start = new Date(data.startDate);
  const end = new Date(data.endDate);
  const diffTime = Math.abs(end.getTime() - start.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)); 
  return diffDays >= 28;
}, {
  message: "Timeframe must be at least 1 month long.",
  path: ["endDate"]
})

// Hardcoded supported strategies for the MVP
const SUPPORTED_STRATEGIES = [
  { id: 'Buy and Hold', label: 'Buy and Hold' },
  { id: 'SMA Cross (50, 200)', label: 'SMA Cross (50, 200)' },
  { id: 'MACD (12, 26, 9)', label: 'MACD (12, 26, 9)' },
  { id: 'RSI Mean Reversion', label: 'RSI Mean Reversion' },
  { id: 'Bollinger Bands', label: 'Bollinger Bands' },
]

function CreateExperiment() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [searchTerm, setSearchTerm] = React.useState('')
  const [debouncedSearch, setDebouncedSearch] = React.useState('')

  React.useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchTerm), 300)
    return () => clearTimeout(timer)
  }, [searchTerm])

  // Fetch available symbols from the search API
  const { data: searchResults, isLoading: isLoadingSymbols } = useQuery({
    queryKey: ['symbols', debouncedSearch],
    queryFn: () => searchSymbols(debouncedSearch),
    enabled: debouncedSearch.length > 0,
  })

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      symbols: [],
      strategies: [],
      startDate: '2019-01-01',
      endDate: '2023-12-31',
    },
  })

  const mutation = useMutation({
    mutationFn: createExperiment,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experiments'] })
      navigate({ to: '/' })
    },
  })

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    // Transform form values into the CreateExperimentRequest JSON structure expected by backend
    const payload = {
      name: values.name,
      description: values.description || '',
      symbols: values.symbols, // Backend expects an array
      strategies: values.strategies.map((strat) => ({
        name: strat,
        parameters: {}, // Sending empty parameters for MVP as discussed
      })),
      timeframes: [
        {
          start: values.startDate + "T00:00:00Z", // Format as RFC3339 for DB if needed, or simple YYYY-MM-DD
          end: values.endDate + "T23:59:59Z",
        },
      ],
    }
    mutation.mutate(payload)
  }

  return (
    <div className="w-full max-w-2xl mx-auto space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex items-center gap-4 mb-8">
        <Link to="/">
          <Button variant="ghost" size="icon" className="rounded-full hover:bg-primary/10 hover:text-primary">
            <ArrowLeft className="h-5 w-5" />
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold tracking-tight glow-text text-primary flex items-center gap-2">
            Create Experiment
          </h1>
          <p className="text-muted-foreground font-mono text-sm">Configure your backtesting parameters.</p>
        </div>
      </div>

      <Card className="border-primary/20 bg-black/40 backdrop-blur-md">
        <CardHeader>
          <CardTitle className="font-mono uppercase tracking-widest text-sm text-primary">Configuration</CardTitle>
          <CardDescription>Select the asset, timeframe, and algorithmic strategy to simulate.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem className="md:col-span-2">
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Experiment Name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g. AAPL Moving Average Test" {...field} />
                      </FormControl>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem className="md:col-span-2">
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Description (Optional)</FormLabel>
                      <FormControl>
                        <Input placeholder="Testing long-term momentum..." {...field} />
                      </FormControl>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="symbols"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Asset Symbols</FormLabel>
                      <FormControl>
                        <MultiSelect
                          options={searchResults?.map((sym: { symbol: string, name: string }) => ({ label: `${sym.symbol} - ${sym.name}`, value: sym.symbol })) || []}
                          selected={field.value}
                          onChange={field.onChange}
                          onSearch={setSearchTerm}
                          placeholder={isLoadingSymbols ? "Searching..." : "Type to search stocks"}
                        />
                      </FormControl>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="strategies"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground">Strategies</FormLabel>
                      <FormControl>
                        <MultiSelect
                          options={SUPPORTED_STRATEGIES.map(s => ({ label: s.label, value: s.id }))}
                          selected={field.value}
                          onChange={field.onChange}
                          placeholder="Select algorithms"
                        />
                      </FormControl>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="startDate"
                  render={({ field }) => (
                    <FormItem className="flex flex-col justify-end">
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground mb-1">Start Date</FormLabel>
                      <Popover>
                        <PopoverTrigger asChild>
                          <FormControl>
                            <Button
                              variant={"outline"}
                              className={cn(
                                "w-full pl-3 text-left font-normal h-10",
                                !field.value && "text-muted-foreground"
                              )}
                            >
                              {field.value ? (
                                format(new Date(field.value + "T12:00:00Z"), "PPP")
                              ) : (
                                <span>Pick a date</span>
                              )}
                              <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                            </Button>
                          </FormControl>
                        </PopoverTrigger>
                        <PopoverContent className="w-auto p-0" align="start">
                          <Calendar
                            mode="single"
                            captionLayout="dropdown"
                            startMonth={new Date(1990, 0)}
                            endMonth={new Date()}
                            selected={field.value ? new Date(field.value + "T12:00:00Z") : undefined}
                            onSelect={(date: Date | undefined) => field.onChange(date ? format(date, 'yyyy-MM-dd') : '')}
                            disabled={(date: Date) =>
                              date > new Date() || date < new Date("1900-01-01")
                            }
                          />
                        </PopoverContent>
                      </Popover>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="endDate"
                  render={({ field }) => (
                    <FormItem className="flex flex-col justify-end">
                      <FormLabel className="font-mono uppercase text-xs text-muted-foreground mb-1">End Date</FormLabel>
                      <Popover>
                        <PopoverTrigger asChild>
                          <FormControl>
                            <Button
                              variant={"outline"}
                              className={cn(
                                "w-full pl-3 text-left font-normal h-10",
                                !field.value && "text-muted-foreground"
                              )}
                            >
                              {field.value ? (
                                format(new Date(field.value + "T12:00:00Z"), "PPP")
                              ) : (
                                <span>Pick a date</span>
                              )}
                              <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                            </Button>
                          </FormControl>
                        </PopoverTrigger>
                        <PopoverContent className="w-auto p-0" align="start">
                          <Calendar
                            mode="single"
                            captionLayout="dropdown"
                            startMonth={new Date(1990, 0)}
                            endMonth={new Date()}
                            selected={field.value ? new Date(field.value + "T12:00:00Z") : undefined}
                            onSelect={(date: Date | undefined) => field.onChange(date ? format(date, 'yyyy-MM-dd') : '')}
                            disabled={(date: Date) =>
                              date > new Date() || date < new Date("1900-01-01")
                            }
                          />
                        </PopoverContent>
                      </Popover>
                      <FormMessage className="font-mono text-xs" />
                    </FormItem>
                  )}
                />
              </div>

              {mutation.isError && (
                <div className="text-red-500 font-mono text-sm p-4 border border-red-500/20 rounded bg-red-500/10">
                  {mutation.error.message}
                </div>
              )}

              <Button 
                type="submit" 
                variant="glow" 
                disabled={mutation.isPending} 
                className="w-full font-mono tracking-widest relative mt-4 h-12"
              >
                {mutation.isPending ? (
                  <div className="flex items-center justify-center gap-2">
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent" />
                    LAUNCHING...
                  </div>
                ) : (
                  <div className="flex items-center justify-center gap-2">
                    <Rocket className="h-4 w-4" />
                    SUBMIT EXPERIMENT
                  </div>
                )}
              </Button>

            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}
