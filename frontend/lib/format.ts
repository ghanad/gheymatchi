export function formatIRR(value: number): string {
  return `${new Intl.NumberFormat("en-US").format(value)} IRR`;
}

export function formatDecimalText(value: string, suffix: string): string {
  const numericValue = Number(value);
  if (!Number.isFinite(numericValue)) {
    return `${value} ${suffix}`;
  }
  return `${new Intl.NumberFormat("en-US", {
    maximumFractionDigits: 8
  }).format(numericValue)} ${suffix}`;
}

export function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}
