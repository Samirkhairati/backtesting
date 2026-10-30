import type { Experiment, ExperimentEvent, ExperimentsResponse, SymbolInfo } from "./types";
import { recordServerTime } from "./useServerClock";

// Same-origin by default: nginx (prod) and the Vite dev server both proxy /api to the API.
// Set VITE_API_URL at build time to point at a separately hosted API.
const API_URL = import.meta.env.VITE_API_URL ?? "/api/v1";

function getHeaders() {
  const token = localStorage.getItem("token");
  return {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function fetchExperiments(): Promise<ExperimentsResponse> {
  const res = await fetch(`${API_URL}/experiments`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error("Unauthorized");
    throw new Error("Failed to fetch experiments");
  }
  const data = (await res.json()) as ExperimentsResponse;
  recordServerTime(data.server_time);
  return { ...data, experiments: data.experiments ?? [] };
}

export async function fetchExperimentById(id: string): Promise<Experiment> {
  const res = await fetch(`${API_URL}/experiments/${id}`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error("Unauthorized");
    throw new Error("Failed to fetch experiment");
  }
  const data = (await res.json()) as Experiment;
  recordServerTime(data.server_time);
  return data;
}

export async function fetchSymbols(): Promise<SymbolInfo[]> {
  const res = await fetch(`${API_URL}/data/symbols`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    throw new Error("Failed to fetch symbols");
  }
  const data = await res.json();
  return data.symbols;
}

export async function searchSymbols(query: string): Promise<SymbolInfo[]> {
  if (!query) return [];
  const res = await fetch(`${API_URL}/data/symbols/search?q=${encodeURIComponent(query)}`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    throw new Error("Failed to search symbols");
  }
  const data = await res.json();
  return data.symbols;
}

export async function createExperiment(payload: Record<string, unknown>): Promise<unknown> {
  const res = await fetch(`${API_URL}/experiments`, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Failed to create experiment");
  }
  return res.json();
}

export async function cancelExperiment(id: string): Promise<unknown> {
  const res = await fetch(`${API_URL}/experiments/${id}/cancel`, {
    method: "POST",
    headers: getHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || "Failed to cancel experiment");
  }
  return res.json();
}

export async function login(payload: Record<string, unknown>): Promise<{ token?: string }> {
  const res = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || "Login failed");
  }
  return res.json();
}

export async function register(payload: Record<string, unknown>): Promise<{ token?: string }> {
  const res = await fetch(`${API_URL}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || "Registration failed");
  }
  return res.json();
}

export function subscribeToExperimentEvents(
  experimentId: string,
  onEvent: (eventType: string, data: ExperimentEvent) => void,
  signal?: AbortSignal
) {
  const token = localStorage.getItem("token");
  // EventSource cannot send headers, so the token goes in the query string.
  const url = `${API_URL}/experiments/${experimentId}/events${token ? `?token=${encodeURIComponent(token)}` : ""}`;

  const eventSource = new EventSource(url);

  eventSource.onmessage = (event) => {
    try {
      const parsed = JSON.parse(event.data) as ExperimentEvent;
      onEvent(parsed.event_type, parsed);
    } catch (err) {
      console.error("Failed to parse SSE message", err);
    }
  };

  // Do not close on error: EventSource reconnects on its own. Polling is the source of truth anyway.

  const close = () => eventSource.close();
  if (signal) {
    if (signal.aborted) close();
    else signal.addEventListener("abort", close, { once: true });
  }

  return close;
}
