export type ApiStatus = {
  status: string;
};

const defaultBaseURL = "http://localhost:8080";

export function apiBaseURL() {
  return process.env.BACKEND_API_BASE_URL || defaultBaseURL;
}

export async function fetchReadiness(): Promise<ApiStatus> {
  const response = await fetch(`${apiBaseURL()}/readyz`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`API readiness check failed with ${response.status}`);
  }

  return response.json() as Promise<ApiStatus>;
}
