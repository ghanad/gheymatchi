"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import {
  createProductSource,
  deleteProductSource,
  fetchProduct,
  fetchProductPricePoints,
  fetchProductSources,
  PricePoint,
  Product,
  ProductSource,
  ProductSourceInput,
  updateProductSource
} from "../../../lib/api";
import { formatDateTime, formatDecimalText, formatIRR } from "../../../lib/format";

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
  const [pricePoints, setPricePoints] = useState<PricePoint[]>([]);
  const [form, setForm] = useState<ProductSourceInput>(emptySourceForm);
  const [editingSourceID, setEditingSourceID] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isPriceLoading, setIsPriceLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [priceError, setPriceError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function loadProductDetail() {
      try {
        setIsLoading(true);
        const [loadedProduct, loadedSources, loadedPricePoints] = await Promise.all([
          fetchProduct(productID),
          fetchProductSources(productID),
          fetchProductPricePoints(productID)
        ]);
        if (isMounted) {
          setProduct(loadedProduct);
          setSources(loadedSources);
          setPricePoints(loadedPricePoints);
          setError(null);
          setPriceError(null);
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

  async function refreshPriceHistory() {
    try {
      setIsPriceLoading(true);
      const loadedPricePoints = await fetchProductPricePoints(productID);
      setPricePoints(loadedPricePoints);
      setPriceError(null);
    } catch (err) {
      setPriceError(err instanceof Error ? err.message : "Could not load price history");
    } finally {
      setIsPriceLoading(false);
    }
  }

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

      {product ? (
        <section className="panel price-history">
          <div className="panel-title">
            <h2>Price history</h2>
            <button type="button" className="secondary-button" onClick={refreshPriceHistory} disabled={isPriceLoading}>
              {isPriceLoading ? "Refreshing..." : "Refresh"}
            </button>
          </div>
          {priceError ? <div className="form-error">{priceError}</div> : null}
          {isLoading ? <p>Loading price history...</p> : null}
          {!isLoading && pricePoints.length === 0 ? <p>No price history yet.</p> : null}
          {pricePoints.length > 0 ? (
            <>
              <PriceChart pricePoints={pricePoints} />
              <PriceTable pricePoints={pricePoints} sources={sources} />
            </>
          ) : null}
        </section>
      ) : null}
    </>
  );
}

function PriceChart({ pricePoints }: { pricePoints: PricePoint[] }) {
  const width = 720;
  const height = 220;
  const padding = 28;
  const minPrice = Math.min(...pricePoints.map((point) => point.price_irr));
  const maxPrice = Math.max(...pricePoints.map((point) => point.price_irr));
  const priceRange = Math.max(maxPrice - minPrice, 1);
  const timeValues = pricePoints.map((point) => new Date(point.captured_at).getTime());
  const minTime = Math.min(...timeValues);
  const maxTime = Math.max(...timeValues);
  const timeRange = Math.max(maxTime - minTime, 1);

  const coordinates = pricePoints
    .map((point) => {
      const capturedAt = new Date(point.captured_at).getTime();
      const x = padding + ((capturedAt - minTime) / timeRange) * (width - padding * 2);
      const y = height - padding - ((point.price_irr - minPrice) / priceRange) * (height - padding * 2);
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <div className="price-chart" aria-label="Price IRR over time">
      <svg viewBox={`0 0 ${width} ${height}`} role="img">
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} />
        <line x1={padding} y1={padding} x2={padding} y2={height - padding} />
        <polyline points={coordinates} />
        {pricePoints.map((point) => {
          const capturedAt = new Date(point.captured_at).getTime();
          const x = padding + ((capturedAt - minTime) / timeRange) * (width - padding * 2);
          const y = height - padding - ((point.price_irr - minPrice) / priceRange) * (height - padding * 2);
          return (
            <circle key={point.id} cx={x} cy={y} r="4">
              <title>{`${formatDateTime(point.captured_at)} - ${formatIRR(point.price_irr)}`}</title>
            </circle>
          );
        })}
      </svg>
      <div className="chart-scale">
        <span>{formatIRR(minPrice)}</span>
        <span>{formatIRR(maxPrice)}</span>
      </div>
    </div>
  );
}

function PriceTable({ pricePoints, sources }: { pricePoints: PricePoint[]; sources: ProductSource[] }) {
  const sourceNames = new Map(sources.map((source) => [source.id, source.source_name]));
  const recentPoints = [...pricePoints].reverse().slice(0, 20);

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Captured</th>
            <th>IRR price</th>
            <th>USD</th>
            <th>Gold gram</th>
            <th>Source</th>
          </tr>
        </thead>
        <tbody>
          {recentPoints.map((point) => (
            <tr key={point.id}>
              <td>{formatDateTime(point.captured_at)}</td>
              <td>{formatIRR(point.price_irr)}</td>
              <td>{point.price_usd ? formatDecimalText(point.price_usd, "USD") : "Not available"}</td>
              <td>{point.price_gold_gram ? formatDecimalText(point.price_gold_gram, "g") : "Not available"}</td>
              <td>{sourceNames.get(point.product_source_id) || "Unknown source"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
