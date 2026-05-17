import Link from "next/link";

export default function ProductsPage() {
  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Products</span>
        <h1>Products</h1>
      </section>
      <section className="placeholder">
        <p>Product listing and editing will be connected in the next frontend phase.</p>
        <div className="link-list">
          <Link className="link-row" href="/products/example">
            Product detail placeholder <span>Open</span>
          </Link>
        </div>
      </section>
    </>
  );
}
