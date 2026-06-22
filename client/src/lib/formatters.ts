export function formatDuration(start: string, end?: string) {
  const startDate = new Date(start).getTime();
  const endDate = end ? new Date(end).getTime() : Date.now();
  const ms = endDate - startDate;
  return `${ms} ms`;
}
