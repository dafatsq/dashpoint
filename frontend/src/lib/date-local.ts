const jakartaDateFormatter = new Intl.DateTimeFormat("en-CA", {
  timeZone: "Asia/Jakarta",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

export function getJakartaDateString(date = new Date()): string {
  return jakartaDateFormatter.format(date);
}
