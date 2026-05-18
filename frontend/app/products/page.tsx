"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { createProduct, deleteProduct, fetchProducts, Product, ProductInput, updateProduct } from "../../lib/api";

const emptyForm: ProductInput = {
  name: "",
  description: ""
};

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [form, setForm] = useState<ProductInput>(emptyForm);
  const [editingProductID, setEditingProductID] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function loadProducts() {
      try {
        setIsLoading(true);
        const loadedProducts = await fetchProducts();
        if (isMounted) {
          setProducts(loadedProducts);
          setError(null);
        }
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

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSaving(true);
      if (editingProductID) {
        const updated = await updateProduct(editingProductID, form);
        setProducts((current) => current.map((product) => (product.id === updated.id ? updated : product)));
      } else {
        const created = await createProduct(form);
        setProducts((current) => [created, ...current]);
      }
      resetForm();
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save product");
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDelete(productID: string) {
    try {
      await deleteProduct(productID);
      setProducts((current) => current.filter((product) => product.id !== productID));
      if (editingProductID === productID) {
        resetForm();
      }
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete product");
    }
  }

  function startEditing(product: Product) {
    setEditingProductID(product.id);
    setForm({
      name: product.name,
      description: product.description || ""
    });
  }

  function resetForm() {
    setEditingProductID(null);
    setForm(emptyForm);
  }

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Products</span>
        <h1>Products</h1>
      </section>

      <section className="panel product-workspace">
        {error ? <div className="form-error">{error}</div> : null}

        <div className="product-grid">
          <form className="product-form" onSubmit={handleSubmit}>
            <h2>{editingProductID ? "Edit product" : "New product"}</h2>
            <label className="field">
              <span>Name</span>
              <input
                value={form.name}
                onChange={(event) => setForm({ ...form, name: event.target.value })}
                placeholder="iPhone 16"
                required
              />
            </label>
            <label className="field">
              <span>Description</span>
              <textarea
                value={form.description || ""}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
                placeholder="Optional notes"
              />
            </label>
            <div className="button-row">
              <button type="submit" disabled={isSaving}>
                {editingProductID ? "Save" : "Create"}
              </button>
              {editingProductID ? (
                <button type="button" className="secondary-button" onClick={resetForm}>
                  Cancel
                </button>
              ) : null}
            </div>
          </form>

          <div className="product-list">
            <h2>Tracked products</h2>
            {isLoading ? <p>Loading products...</p> : null}
            {!isLoading && products.length === 0 ? <p>No products yet.</p> : null}
            {products.map((product) => (
              <article className="product-row" key={product.id}>
                <div>
                  <strong>{product.name}</strong>
                  {product.description ? <span>{product.description}</span> : <span>No description</span>}
                </div>
                <div className="row-actions">
                  <Link className="secondary-link-button" href={`/products/${product.id}`}>
                    Sources
                  </Link>
                  <button type="button" className="secondary-button" onClick={() => startEditing(product)}>
                    Edit
                  </button>
                  <button type="button" className="danger-button" onClick={() => handleDelete(product.id)}>
                    Delete
                  </button>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
