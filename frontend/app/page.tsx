import Link from "next/link";
import { StatusBadge } from "../components/status-badge";
import { apiBaseURL, fetchReadiness } from "../lib/api";

async function getApiState() {
  try {
    const readiness = await fetchReadiness();
    return {
      label: readiness.status === "ready" ? "Ready" : readiness.status,
      message: `Connected to ${apiBaseURL()}`,
      status: readiness.status === "ready" ? ("ready" as const) : ("warning" as const)
    };
  } catch (error) {
    return {
      label: "Unavailable",
      message: error instanceof Error ? error.message : "API readiness check failed",
      status: "error" as const
    };
  }
}

export default async function DashboardPage() {
  const apiState = await getApiState();

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Dashboard</span>
        <h1>Price tracking workspace</h1>
        <p>Local MVP dashboard for products, source URLs, and stored price history.</p>
      </section>

      <section className="dashboard-grid" aria-label="Dashboard summary">
        <article className="panel">
          <div className="panel-title">
            <h2>API</h2>
            <StatusBadge label={apiState.label} status={apiState.status} />
          </div>
          <p>{apiState.message}</p>
        </article>

        <article className="panel">
          <div className="panel-title">
            <h2>Products</h2>
          </div>
          <div className="metric">0</div>
          <p>Connect product management in the next frontend phase.</p>
        </article>

        <article className="panel">
          <div className="panel-title">
            <h2>Alerts</h2>
          </div>
          <div className="metric">0</div>
          <p>Alert rules are not part of this phase.</p>
        </article>

        <article className="panel panel-wide">
          <div className="panel-title">
            <h2>Workspace</h2>
          </div>
          <div className="link-list">
            <Link className="link-row" href="/products">
              Products <span>Open</span>
            </Link>
            <Link className="link-row" href="/alerts">
              Alerts <span>Open</span>
            </Link>
            <Link className="link-row" href="/settings">
              Settings <span>Open</span>
            </Link>
          </div>
        </article>

        <article className="panel">
          <div className="panel-title">
            <h2>Database</h2>
          </div>
          <p>SQLite is used for the local MVP.</p>
        </article>
      </section>
    </>
  );
}
