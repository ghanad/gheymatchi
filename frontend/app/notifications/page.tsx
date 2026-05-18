import { fetchNotifications } from "../../lib/api";
import type { Notification } from "../../lib/api";
import { formatDateTime } from "../../lib/format";

function statusClass(status: string) {
  if (status === "sent") {
    return "status-badge status-ready";
  }
  if (status === "failed") {
    return "status-badge status-error";
  }
  return "status-badge status-warning";
}

export default async function NotificationsPage() {
  let notifications: Notification[] = [];
  let error: string | null = null;

  try {
    notifications = await fetchNotifications();
  } catch (err) {
    error = err instanceof Error ? err.message : "Could not load notifications";
  }

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Notifications</span>
        <h1>Notification log</h1>
      </section>

      <section className="panel notification-list">
        {error ? <div className="form-error">{error}</div> : null}
        {!error && notifications.length === 0 ? <p>No notifications have been created yet.</p> : null}
        {notifications.map((notification) => (
          <article className="notification-row" key={notification.id}>
            <div>
              <strong>{notification.channel.replace("_", " ")}</strong>
              <span>Recipient: {notification.recipient}</span>
              {notification.alert_id ? <span>Alert: {notification.alert_id}</span> : null}
            </div>
            <span className={statusClass(notification.status)}>{notification.status}</span>
            <div className="notification-time">
              <span>Created {formatDateTime(notification.created_at)}</span>
              {notification.sent_at ? <span>Sent {formatDateTime(notification.sent_at)}</span> : null}
            </div>
          </article>
        ))}
      </section>
    </>
  );
}
