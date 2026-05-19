CREATE TABLE users (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE products (
	id TEXT PRIMARY KEY,
	user_id TEXT,
	name TEXT NOT NULL,
	description TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE product_sources (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL,
	url TEXT NOT NULL,
	source_name TEXT,
	is_active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
	UNIQUE (product_id, url)
);

CREATE TABLE price_points (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL,
	product_source_id TEXT NOT NULL,
	price_irr BIGINT NOT NULL,
	usd_irr_rate_value_text TEXT,
	gold_gram_irr_rate_value_text TEXT,
	price_usd TEXT,
	price_gold_gram TEXT,
	captured_at TEXT NOT NULL,
	raw_payload TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
	FOREIGN KEY (product_source_id) REFERENCES product_sources(id) ON DELETE CASCADE
);

CREATE TABLE market_rates (
	id TEXT PRIMARY KEY,
	rate_type TEXT NOT NULL,
	unit TEXT NOT NULL,
	value_text TEXT NOT NULL,
	observed_at TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE alerts (
	id TEXT PRIMARY KEY,
	user_id TEXT,
	product_id TEXT,
	name TEXT NOT NULL,
	condition_type TEXT NOT NULL,
	target_unit TEXT NOT NULL DEFAULT 'IRR',
	threshold_value_text TEXT NOT NULL,
	is_active INTEGER NOT NULL DEFAULT 1,
	last_triggered_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE notifications (
	id TEXT PRIMARY KEY,
	alert_id TEXT,
	channel TEXT NOT NULL,
	recipient TEXT NOT NULL,
	status TEXT NOT NULL,
	attempt_count INTEGER NOT NULL DEFAULT 0,
	last_error TEXT,
	sent_at TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE SET NULL
);

CREATE TABLE crawl_runs (
	id TEXT PRIMARY KEY,
	source_id TEXT,
	status TEXT NOT NULL,
	started_at TEXT NOT NULL,
	finished_at TEXT,
	error_message TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (source_id) REFERENCES product_sources(id) ON DELETE SET NULL
);

CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	token_hash TEXT NOT NULL UNIQUE,
	created_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_products_user_id ON products(user_id);
CREATE INDEX idx_product_sources_product_id ON product_sources(product_id);
CREATE INDEX idx_price_points_product_captured_at ON price_points(product_id, captured_at);
CREATE INDEX idx_price_points_source_captured_at ON price_points(product_source_id, captured_at);
CREATE INDEX idx_market_rates_type_observed_at ON market_rates(rate_type, observed_at);
CREATE INDEX idx_alerts_product_id ON alerts(product_id);
CREATE INDEX idx_notifications_alert_id ON notifications(alert_id);
CREATE INDEX idx_crawl_runs_source_started_at ON crawl_runs(source_id, started_at);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
