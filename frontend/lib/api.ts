export type ApiStatus = {
  status: string;
};

export type Product = {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
};

export type Alert = {
  id: string;
  product_id: string;
  name: string;
  condition_type: "BELOW" | "ABOVE";
  target_unit: "IRR" | "USD" | "GOLD_GRAM";
  threshold_value_text: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type AlertInput = {
  name: string;
  condition_type: Alert["condition_type"];
  target_unit: Alert["target_unit"];
  threshold_value_text: string;
  is_active?: boolean;
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

export async function fetchProducts(): Promise<Product[]> {
  const response = await fetch(`${apiBaseURL()}/api/products`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Product request failed with ${response.status}`);
  }

  const body = (await response.json()) as { products: Product[] };
  return body.products;
}

export async function fetchProductAlerts(productID: string): Promise<Alert[]> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/alerts`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Alert request failed with ${response.status}`);
  }

  const body = (await response.json()) as { alerts: Alert[] };
  return body.alerts;
}

export async function createProductAlert(productID: string, input: AlertInput): Promise<Alert> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/alerts`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Create alert failed"));
  }

  return response.json() as Promise<Alert>;
}

export async function updateProductAlert(productID: string, alertID: string, input: Partial<AlertInput>): Promise<Alert> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/alerts/${alertID}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Update alert failed"));
  }

  return response.json() as Promise<Alert>;
}

export async function deleteProductAlert(productID: string, alertID: string): Promise<void> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/alerts/${alertID}`, {
    method: "DELETE"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Delete alert failed"));
  }
}

async function errorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const body = (await response.json()) as { error?: { message?: string } };
    return body.error?.message || `${fallback} with ${response.status}`;
  } catch {
    return `${fallback} with ${response.status}`;
  }
}
