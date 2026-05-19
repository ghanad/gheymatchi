const defaultBackendBaseURL = "http://localhost:8080";

function backendReadyURL() {
  return new URL("/readyz", process.env.BACKEND_API_BASE_URL || defaultBackendBaseURL);
}

export async function GET() {
  const response = await fetch(backendReadyURL(), {
    cache: "no-store"
  });

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: response.headers
  });
}
