type StatusBadgeProps = {
  status: "ready" | "warning" | "error";
  label: string;
};

export function StatusBadge({ status, label }: StatusBadgeProps) {
  return <span className={`status-badge status-${status}`}>{label}</span>;
}
