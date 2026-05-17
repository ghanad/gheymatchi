CREATE TABLE price_points_new (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL,
	product_source_id TEXT NOT NULL,
	price_irr INTEGER NOT NULL,
	captured_at TEXT NOT NULL,
	raw_payload TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
	FOREIGN KEY (product_source_id) REFERENCES product_sources(id) ON DELETE CASCADE
);

INSERT INTO price_points_new (id, product_id, product_source_id, price_irr, captured_at, created_at)
SELECT price_points.id, price_points.product_id, price_points.source_id, price_points.amount_minor, price_points.captured_at, price_points.created_at
FROM price_points
JOIN product_sources ON product_sources.id = price_points.source_id
WHERE price_points.amount_minor > 0;

DROP TABLE price_points;
ALTER TABLE price_points_new RENAME TO price_points;

CREATE INDEX idx_price_points_product_captured_at ON price_points(product_id, captured_at);
CREATE INDEX idx_price_points_source_captured_at ON price_points(product_source_id, captured_at);
