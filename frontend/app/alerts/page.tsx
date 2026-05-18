"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  AlertInput,
  createProductAlert,
  deleteProductAlert,
  fetchProductAlerts,
  fetchProducts,
  Product,
  updateProductAlert
} from "../../lib/api";

const defaultForm: AlertInput = {
  name: "",
  condition_type: "BELOW",
  target_unit: "IRR",
  threshold_value_text: "",
  is_active: true
};

export default function AlertsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProductID, setSelectedProductID] = useState("");
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [form, setForm] = useState<AlertInput>(defaultForm);
  const [editingAlertID, setEditingAlertID] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectedProduct = useMemo(
    () => products.find((product) => product.id === selectedProductID),
    [products, selectedProductID]
  );

  useEffect(() => {
    let isMounted = true;

    async function loadProducts() {
      try {
        setIsLoading(true);
        const loadedProducts = await fetchProducts();
        if (!isMounted) {
          return;
        }
        setProducts(loadedProducts);
        setSelectedProductID((current) => current || loadedProducts[0]?.id || "");
        setError(null);
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : "Could not load products");
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadProducts();
    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    let isMounted = true;

    async function loadAlerts() {
      if (!selectedProductID) {
        setAlerts([]);
        return;
      }
      try {
        const loadedAlerts = await fetchProductAlerts(selectedProductID);
        if (isMounted) {
          setAlerts(loadedAlerts);
          setError(null);
        }
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : "Could not load alerts");
        }
      }
    }

    loadAlerts();
    return () => {
      isMounted = false;
    };
  }, [selectedProductID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedProductID) {
      return;
    }

    try {
      setIsSaving(true);
      if (editingAlertID) {
        const updated = await updateProductAlert(selectedProductID, editingAlertID, form);
        setAlerts((current) => current.map((alert) => (alert.id === updated.id ? updated : alert)));
      } else {
        const created = await createProductAlert(selectedProductID, form);
        setAlerts((current) => [created, ...current]);
      }
      resetForm();
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save alert");
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDelete(alertID: string) {
    if (!selectedProductID) {
      return;
    }
    try {
      await deleteProductAlert(selectedProductID, alertID);
      setAlerts((current) => current.filter((alert) => alert.id !== alertID));
      if (editingAlertID === alertID) {
        resetForm();
      }
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete alert");
    }
  }

  function startEditing(alert: Alert) {
    setEditingAlertID(alert.id);
    setForm({
      name: alert.name,
      condition_type: alert.condition_type,
      target_unit: alert.target_unit,
      threshold_value_text: alert.threshold_value_text,
      is_active: alert.is_active
    });
  }

  function resetForm() {
    setEditingAlertID(null);
    setForm(defaultForm);
  }

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Alerts</span>
        <h1>Alert rules</h1>
      </section>

      <section className="panel alert-workspace">
        {error ? <div className="form-error">{error}</div> : null}

        <label className="field">
          <span>Product</span>
          <select
            value={selectedProductID}
            onChange={(event) => {
              setSelectedProductID(event.target.value);
              resetForm();
            }}
            disabled={isLoading || products.length === 0}
          >
            {products.map((product) => (
              <option key={product.id} value={product.id}>
                {product.name}
              </option>
            ))}
          </select>
        </label>

        {products.length === 0 && !isLoading ? (
          <p>No products exist yet. Create a product before adding alert rules.</p>
        ) : null}

        {selectedProduct ? (
          <div className="alert-grid">
            <form className="alert-form" onSubmit={handleSubmit}>
              <h2>{editingAlertID ? "Edit alert" : "New alert"}</h2>
              <label className="field">
                <span>Name</span>
                <input
                  value={form.name}
                  onChange={(event) => setForm({ ...form, name: event.target.value })}
                  placeholder="Target price"
                  required
                />
              </label>
              <div className="field-row">
                <label className="field">
                  <span>Condition</span>
                  <select
                    value={form.condition_type}
                    onChange={(event) =>
                      setForm({ ...form, condition_type: event.target.value as AlertInput["condition_type"] })
                    }
                  >
                    <option value="BELOW">Below</option>
                    <option value="ABOVE">Above</option>
                  </select>
                </label>
                <label className="field">
                  <span>Unit</span>
                  <select
                    value={form.target_unit}
                    onChange={(event) => setForm({ ...form, target_unit: event.target.value as AlertInput["target_unit"] })}
                  >
                    <option value="IRR">IRR</option>
                    <option value="USD">USD</option>
                    <option value="GOLD_GRAM">Gold gram</option>
                  </select>
                </label>
              </div>
              <label className="field">
                <span>Target value</span>
                <input
                  value={form.threshold_value_text}
                  onChange={(event) => setForm({ ...form, threshold_value_text: event.target.value })}
                  inputMode="decimal"
                  placeholder="85000000"
                  required
                />
              </label>
              <label className="check-field">
                <input
                  type="checkbox"
                  checked={form.is_active ?? true}
                  onChange={(event) => setForm({ ...form, is_active: event.target.checked })}
                />
                Active
              </label>
              <div className="button-row">
                <button type="submit" disabled={isSaving}>
                  {editingAlertID ? "Save" : "Create"}
                </button>
                {editingAlertID ? (
                  <button type="button" className="secondary-button" onClick={resetForm}>
                    Cancel
                  </button>
                ) : null}
              </div>
            </form>

            <div className="alert-list">
              <h2>{selectedProduct.name}</h2>
              {alerts.length === 0 ? <p>No alert rules for this product.</p> : null}
              {alerts.map((alert) => (
                <article className="alert-row" key={alert.id}>
                  <div>
                    <strong>{alert.name}</strong>
                    <span>
                      {alert.condition_type.toLowerCase()} {alert.threshold_value_text} {alert.target_unit}
                    </span>
                  </div>
                  <span className={alert.is_active ? "status-badge status-ready" : "status-badge status-warning"}>
                    {alert.is_active ? "Active" : "Paused"}
                  </span>
                  <div className="row-actions">
                    <button type="button" className="secondary-button" onClick={() => startEditing(alert)}>
                      Edit
                    </button>
                    <button type="button" className="danger-button" onClick={() => handleDelete(alert.id)}>
                      Delete
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </div>
        ) : null}
      </section>
    </>
  );
}
