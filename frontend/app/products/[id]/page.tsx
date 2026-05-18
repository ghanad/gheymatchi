"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import {
  createProductSource,
  deleteProductSource,
  fetchProduct,
  fetchProductSources,
  Product,
  ProductSource,
  ProductSourceInput,
  updateProductSource
} from "../../../lib/api";

const emptySourceForm: ProductSourceInput = {
  url: "",
  source_name: "",
  is_active: true
};

export default function ProductDetailPage() {
  const params = useParams<{ id: string }>();
  const productID = params.id;
  const [product, setProduct] = useState<Product | null>(null);
  const [sources, setSources] = useState<ProductSource[]>([]);
  const [form, setForm] = useState<ProductSourceInput>(emptySourceForm);
  const [editingSourceID, setEditingSourceID] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function loadProductDetail() {
      try {
        setIsLoading(true);
        const [loadedProduct, loadedSources] = await Promise.all([fetchProduct(productID), fetchProductSources(productID)]);
        if (isMounted) {
          setProduct(loadedProduct);
          setSources(loadedSources);
          setError(null);
        }
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : "Could not load product");
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadProductDetail();
    return () => {
      isMounted = false;
    };
  }, [productID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSaving(true);
      if (editingSourceID) {
        const updated = await updateProductSource(productID, editingSourceID, form);
        setSources((current) => current.map((source) => (source.id === updated.id ? updated : source)));
      } else {
        const created = await createProductSource(productID, form);
        setSources((current) => [created, ...current]);
      }
      resetForm();
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save source");
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDelete(sourceID: string) {
    try {
      await deleteProductSource(productID, sourceID);
      setSources((current) => current.filter((source) => source.id !== sourceID));
      if (editingSourceID === sourceID) {
        resetForm();
      }
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete source");
    }
  }

  async function toggleSource(source: ProductSource) {
    try {
      const updated = await updateProductSource(productID, source.id, { is_active: !source.is_active });
      setSources((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update source");
    }
  }

  function startEditing(source: ProductSource) {
    setEditingSourceID(source.id);
    setForm({
      url: source.url,
      source_name: source.source_name,
      is_active: source.is_active
    });
  }

  function resetForm() {
    setEditingSourceID(null);
    setForm(emptySourceForm);
  }

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Product</span>
        <h1>{product?.name || "Product detail"}</h1>
        {product?.description ? <p>{product.description}</p> : null}
      </section>

      <section className="panel product-workspace">
        <Link className="back-link" href="/products">
          Back to products
        </Link>
        {error ? <div className="form-error">{error}</div> : null}
        {isLoading ? <p>Loading product...</p> : null}

        {product ? (
          <div className="product-grid">
            <form className="product-form" onSubmit={handleSubmit}>
              <h2>{editingSourceID ? "Edit source" : "New source"}</h2>
              <label className="field">
                <span>Source URL</span>
                <input
                  value={form.url}
                  onChange={(event) => setForm({ ...form, url: event.target.value })}
                  placeholder="https://example.com/product"
                  required
                />
              </label>
              <label className="field">
                <span>Source name</span>
                <input
                  value={form.source_name || ""}
                  onChange={(event) => setForm({ ...form, source_name: event.target.value })}
                  placeholder="digikala"
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
                  {editingSourceID ? "Save" : "Add source"}
                </button>
                {editingSourceID ? (
                  <button type="button" className="secondary-button" onClick={resetForm}>
                    Cancel
                  </button>
                ) : null}
              </div>
            </form>

            <div className="product-list">
              <h2>Source URLs</h2>
              {sources.length === 0 ? <p>No sources for this product.</p> : null}
              {sources.map((source) => (
                <article className="product-row source-row" key={source.id}>
                  <div>
                    <strong>{source.source_name}</strong>
                    <a href={source.url} target="_blank" rel="noreferrer">
                      {source.url}
                    </a>
                  </div>
                  <span className={source.is_active ? "status-badge status-ready" : "status-badge status-warning"}>
                    {source.is_active ? "Active" : "Paused"}
                  </span>
                  <div className="row-actions">
                    <button type="button" className="secondary-button" onClick={() => toggleSource(source)}>
                      {source.is_active ? "Deactivate" : "Activate"}
                    </button>
                    <button type="button" className="secondary-button" onClick={() => startEditing(source)}>
                      Edit
                    </button>
                    <button type="button" className="danger-button" onClick={() => handleDelete(source.id)}>
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
