# Sort Trigger Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Put the sort icon and “Sort by:” label inside the shared sort dropdown trigger, and move the Recent Changes sort control into its filter section.

**Architecture:** Keep sorting behavior and option definitions unchanged. Update the shared `DataSortSelect` presentation once so all existing consumers receive the same filter-style trigger, then pass the existing Recent Changes sort state through `ChangesFilters` instead of rendering it beside the entries.

**Tech Stack:** React, Next.js, Radix Select, Tailwind utility classes, Vitest, ESLint, Next production build.

## Global Constraints

- Reuse the existing shared `DataSortSelect` and `FilterCard` patterns.
- Do not add dependencies or change sorting/query behavior.
- Preserve the existing design tokens and responsive layouts.
- Keep changes limited to the shared sort control and Recent Changes filter/list wiring.

---

### Task 1: Make the shared sort trigger self-contained

**Files:**
- Modify: `frontend/src/components/shared/data-sort-select.tsx`

**Interfaces:**
- Consumes: `value`, `options`, `onChange`, and optional `className` from existing callers.
- Produces: A `SelectTrigger` whose visible value contains the `ArrowDownUp` icon, `Sort by:` label, and selected option.

- [x] **Step 1: Replace the external label wrapper contents**

Keep the outer wrapper only for caller layout and render the icon and label inside `SelectTrigger`:

```tsx
<div className={`flex items-center gap-2 ${className || ""}`}>
  <Select value={value} onValueChange={onChange}>
    <SelectTrigger className="min-w-0 flex-1">
      <span className="flex min-w-0 items-center gap-2">
        <ArrowDownUp className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="text-muted-foreground">Sort by:</span>
        <SelectValue />
      </span>
    </SelectTrigger>
    <SelectContent>
      {options.map((option) => (
        <SelectItem key={option.value} value={option.value}>
          {option.label}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
</div>
```

- [x] **Step 2: Run the frontend build**

Run: `cd frontend && npm run build`

Expected: exit code 0 and all application routes compile.

### Task 2: Move Recent Changes sorting into filters

**Files:**
- Modify: `frontend/src/app/(dashboard)/changes/changes-filters.tsx`
- Modify: `frontend/src/app/(dashboard)/changes/changes-list.tsx`
- Modify: `frontend/src/app/(dashboard)/changes/changes-screen.tsx`

**Interfaces:**
- Consumes: Existing `CHANGES_SORT_OPTIONS`, `sort` state, and `setSort` pagination reset behavior.
- Produces: Recent Changes filter card with the same shared sort dropdown as other filter sections.

- [x] **Step 1: Move the option list to the filter module**

Export `CHANGES_SORT_OPTIONS` from `changes-filters.tsx`, add `sort` and `onSortChange` to `ChangesFiltersProps`, and render `DataSortSelect` as the third filter control in the existing responsive grid.

- [x] **Step 2: Pass sort state from the screen**

Add `const [sort, setSort] = useState("date_desc")` to `ChangesScreen`, pass `sort={sort}` and `onSortChange={(value) => setSort(value)}` to `ChangesFilters`, and pass `sort={sort}` into `ChangesList` so fetching remains controlled by the page.

- [x] **Step 3: Remove sort UI and state from the entries section**

Remove `CHANGES_SORT_OPTIONS`, `DataSortSelect`, and local `sort` state from `ChangesList`. Accept `sort` as a prop, use it in `buildDashboardChangeParams` and the effect dependency list, include `sort` in the list reset key so sorting returns to page one, and leave the entries toolbar responsible only for page size and count.

- [x] **Step 4: Run focused frontend validation**

Run: `cd frontend && npm test`

Expected: all existing Vitest files pass.

### Task 3: Review the finished diff

**Files:**
- Review: `frontend/src/components/shared/data-sort-select.tsx`
- Review: `frontend/src/app/(dashboard)/changes/changes-filters.tsx`
- Review: `frontend/src/app/(dashboard)/changes/changes-list.tsx`
- Review: `frontend/src/app/(dashboard)/changes/changes-screen.tsx`

- [x] **Step 1: Run lint and whitespace checks**

Run:

```bash
cd frontend && npm run lint
cd .. && git diff --check
```

Expected: ESLint exits successfully with no new errors, and `git diff --check` produces no output.

- [x] **Step 2: Confirm only requested UI files and the plan are changed**

Run: `git status --short`

Expected: the pre-existing sorting implementation changes remain untouched, with only the shared sort control, Recent Changes files, and this plan added by this follow-up.
