type ProductDetailPageProps = {
  params: Promise<{
    id: string;
  }>;
};

export default async function ProductDetailPage({ params }: ProductDetailPageProps) {
  const { id } = await params;

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Product</span>
        <h1>Product detail</h1>
      </section>
      <section className="placeholder">
        <p>Product ID: {id}</p>
      </section>
    </>
  );
}
