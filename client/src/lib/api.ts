const API_URL = "http://localhost:8080/api/v1";

function getHeaders() {
  const token = localStorage.getItem("token");
  return {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function fetchExperiments() {
  const res = await fetch(`${API_URL}/experiments`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error("Unauthorized");
    throw new Error("Failed to fetch experiments");
  }
  return res.json();
}

export async function fetchSymbols() {
  const res = await fetch(`${API_URL}/data/symbols`, {
    headers: getHeaders(),
  });
  if (!res.ok) {
    throw new Error("Failed to fetch symbols");
  }
  const data = await res.json();
  return data.symbols;
}

export async function createExperiment(payload: Record<string, unknown>) {
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

export async function login(payload: Record<string, unknown>) {
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

export async function register(payload: Record<string, unknown>) {
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
