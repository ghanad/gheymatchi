"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import {
  createProduct,
  createProductSource,
  deleteProduct,
  fetchProducts,
  Product,
  ProductInput,
  updateProduct
} from "../../lib/api";

type ProductForm = ProductInput & {
  source_name: "digikala";
  source_url: string;
};

const emptyForm: ProductForm = {
  name: "",
  description: "",
  source_name: "digikala",
  source_url: ""
};

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [form, setForm] = useState<ProductForm>(emptyForm);
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
        const updated = await updateProduct(editingProductID, {
          name: form.name,
          description: form.description
        });
        setProducts((current) => current.map((product) => (product.id === updated.id ? updated : product)));
      } else {
        if (!isDigikalaProductURL(form.source_url)) {
          throw new Error("Digikala product URL must be from digikala.com and include a dkp product id");
        }
        const created = await createProduct({
          name: form.name,
          description: form.description
        });
        await createProductSource(created.id, {
          source_name: form.source_name,
          url: form.source_url,
          is_active: true
        });
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
      description: product.description || "",
      source_name: "digikala",
      source_url: ""
    });
  }

  function resetForm() {
    setEditingProductID(null);
    setForm({ ...emptyForm });
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
            {!editingProductID ? (
              <>
                <label className="field">
                  <span>Source site</span>
                  <select
                    value={form.source_name}
                    onChange={(event) => setForm({ ...form, source_name: event.target.value as "digikala" })}
                    required
                  >
                    <option value="digikala">Digikala</option>
                  </select>
                </label>
                <label className="field">
                  <span>Digikala product URL</span>
                  <input
                    value={form.source_url}
                    onChange={(event) => setForm({ ...form, source_url: event.target.value })}
                    placeholder="https://www.digikala.com/product/dkp-123456/"
                    required
                  />
                </label>
              </>
            ) : null}
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

function isDigikalaProductURL(value: string) {
  try {
    const url = new URL(value.trim());
    const host = url.hostname.toLowerCase();
    return (
      (host === "digikala.com" || host === "www.digikala.com") &&
      url.pathname
        .split("/")
        .some((part) => /^dkp-\d+$/i.test(part))
    );
  } catch {
    return false;
  }
}
