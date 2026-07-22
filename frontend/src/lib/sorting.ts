export type SortDirection = "asc" | "desc";

export interface SortParams {
  sort_by: string;
  sort_direction: SortDirection;
}

export function parseSortValue(value: string): SortParams {
  const separatorIndex = value.lastIndexOf("_");
  const sortBy = separatorIndex > 0 ? value.slice(0, separatorIndex) : value;
  const direction = separatorIndex > 0 ? value.slice(separatorIndex + 1) : "asc";

  return {
    sort_by: sortBy,
    sort_direction: direction === "desc" ? "desc" : "asc",
  };
}
