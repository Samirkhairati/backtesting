import { useState, useEffect } from 'react'

export function SmoothProgress({ stage, percentage }: { stage: string, percentage: number }) {
  const [displayPercent, setDisplayPercent] = useState(0);

  useEffect(() => {
    let startTimestamp: number;
    const startValue = displayPercent;
    const duration = 1000; // 1 second

    const step = (timestamp: number) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / duration, 1);
      // easeOutQuart
      const ease = 1 - Math.pow(1 - progress, 4);
      setDisplayPercent(Math.round(startValue + (percentage - startValue) * ease));
      
      if (progress < 1) {
        requestAnimationFrame(step);
      }
    };
    
    requestAnimationFrame(step);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [percentage]);

  return (
    <div className="w-full text-xs font-mono">
      <div className="flex justify-between mb-1 text-muted-foreground">
        <span>{stage}</span>
        <span className="tabular-nums">{displayPercent}%</span>
      </div>
      <div className="h-1.5 w-full bg-secondary rounded-full overflow-hidden">
        <div 
          className="h-full bg-primary transition-all duration-1000 ease-out"
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  )
}
