export function planStatusColor(status: string): string {
  switch ((status || "").toLowerCase()) {
    case "scheduled":
      return "#5ce0b2";
    case "planned":
      return "#3db8ff";
    case "canceled":
      return "#ff6b7a";
    case "draft":
      return "#f0c35a";
    default:
      return "#9aadc8";
  }
}
