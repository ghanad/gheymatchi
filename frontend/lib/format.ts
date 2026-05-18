export function formatIRR(value: number): string {
  return `${new Intl.NumberFormat("en-US").format(value)} IRR`;
}

export function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}
