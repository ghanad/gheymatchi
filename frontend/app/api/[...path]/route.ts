const defaultBackendBaseURL = "http://localhost:8080";

type RouteContext = {
  params: Promise<{
    path: string[];
  }>;
};

function backendBaseURL() {
  return process.env.BACKEND_API_BASE_URL || defaultBackendBaseURL;
}

function backendURL(path: string[], search: string) {
  const url = new URL(`/api/${path.join("/")}`, backendBaseURL());
  url.search = search;
  return url;
}

function forwardHeaders(request: Request) {
  const headers = new Headers(request.headers);
  headers.delete("host");
  headers.delete("content-length");
  return headers;
}

async function proxy(request: Request, context: RouteContext) {
  const incomingURL = new URL(request.url);
  const params = await context.params;
  const bodyAllowed = request.method !== "GET" && request.method !== "HEAD";
  const response = await fetch(backendURL(params.path, incomingURL.search), {
    method: request.method,
    headers: forwardHeaders(request),
    body: bodyAllowed ? await request.arrayBuffer() : undefined,
    cache: "no-store"
  });

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: response.headers
  });
}

export async function GET(request: Request, context: RouteContext) {
  return proxy(request, context);
}

export async function POST(request: Request, context: RouteContext) {
  return proxy(request, context);
}

export async function PATCH(request: Request, context: RouteContext) {
  return proxy(request, context);
}

export async function DELETE(request: Request, context: RouteContext) {
  return proxy(request, context);
}
