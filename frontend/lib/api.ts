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

export type ProductInput = {
  name: string;
  description?: string;
};

export type ProductSource = {
  id: string;
  product_id: string;
  url: string;
  source_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type ProductSourceInput = {
  url: string;
  source_name?: string;
  is_active?: boolean;
};

export type PricePoint = {
  id: string;
  product_id: string;
  product_source_id: string;
  price_irr: number;
  usd_irr_rate_value_text?: string;
  gold_gram_irr_rate_value_text?: string;
  price_usd?: string;
  price_gold_gram?: string;
  captured_at: string;
  raw_payload?: string;
  created_at: string;
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

export type Notification = {
  id: string;
  alert_id?: string;
  channel: "email" | "sms" | "dry_run";
  recipient: string;
  status: "pending" | "sent" | "failed";
  attempt_count: number;
  last_error?: string;
  sent_at?: string;
  created_at: string;
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

export async function fetchProduct(productID: string): Promise<Product> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Product request failed"));
  }

  return response.json() as Promise<Product>;
}

export async function createProduct(input: ProductInput): Promise<Product> {
  const response = await fetch(`${apiBaseURL()}/api/products`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Create product failed"));
  }

  return response.json() as Promise<Product>;
}

export async function updateProduct(productID: string, input: Partial<ProductInput>): Promise<Product> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Update product failed"));
  }

  return response.json() as Promise<Product>;
}

export async function deleteProduct(productID: string): Promise<void> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}`, {
    method: "DELETE"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Delete product failed"));
  }
}

export async function fetchProductSources(productID: string): Promise<ProductSource[]> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/sources`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Source request failed"));
  }

  const body = (await response.json()) as { sources: ProductSource[] };
  return body.sources;
}

export async function createProductSource(productID: string, input: ProductSourceInput): Promise<ProductSource> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/sources`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Create source failed"));
  }

  return response.json() as Promise<ProductSource>;
}

export async function updateProductSource(
  productID: string,
  sourceID: string,
  input: Partial<ProductSourceInput>
): Promise<ProductSource> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/sources/${sourceID}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Update source failed"));
  }

  return response.json() as Promise<ProductSource>;
}

export async function deleteProductSource(productID: string, sourceID: string): Promise<void> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/sources/${sourceID}`, {
    method: "DELETE"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Delete source failed"));
  }
}

export async function fetchProductPricePoints(productID: string): Promise<PricePoint[]> {
  const response = await fetch(`${apiBaseURL()}/api/products/${productID}/price-points`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Price history request failed"));
  }

  const body = (await response.json()) as { price_points: PricePoint[] };
  return body.price_points;
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

export async function fetchNotifications(): Promise<Notification[]> {
  const response = await fetch(`${apiBaseURL()}/api/notifications`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, "Notification request failed"));
  }

  const body = (await response.json()) as { notifications: Notification[] };
  return body.notifications;
}

async function errorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const body = (await response.json()) as { error?: { message?: string } };
    return body.error?.message || `${fallback} with ${response.status}`;
  } catch {
    return `${fallback} with ${response.status}`;
  }
}
