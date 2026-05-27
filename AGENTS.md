# Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

# Agent Initialization Directive

This repository uses the following consistency protocols for agentic coding work. Follow both unless a more specific instruction in the session or a repo-local file explicitly overrides them.

## Universal Backend Consistency Protocol

**TARGET AUDIENCE:** Agentic AI Coding Assistants
**PURPOSE:** To enforce strict, zero-deviation architectural, API, and structural consistency across all backend services (Node.js, Go, Python, Java, etc.) within this repository.

### 1. Core Directives (The "Consumer" Rule)
When generating, editing, or refactoring backend code, you are a **CONSUMER** of the existing architecture, never an **INVENTOR**.
* **DO NOT** introduce new structural paradigms, ORMs, frameworks, or architectural patterns on a whim.
* Ensure data flow, API responses, and database interactions share the exact same logical footprint across all endpoints and microservices.
* **If a pattern already exists in the codebase for a specific backend task, you MUST reuse it.**

### 2. Architectural Layers & Separation of Concerns
You must strictly separate route handling from business logic and database access. Never write "fat controllers."

* **Layer 1 - Route/Controller (Delivery):** This layer only handles incoming HTTP/gRPC requests, extracts parameters/bodies, calls the Service layer, and returns the response. **NO business logic or database queries here.**
* **Layer 2 - Service (Business Logic):** This layer contains all the core logic, calculations, and orchestrations. It knows nothing about HTTP request/response objects.
* **Layer 3 - Repository/Data Access:** This layer handles direct interactions with the database, caching layer (Redis), or external APIs.

### 3. API & Payload Standardization
All API endpoints must communicate using a uniform contract.

* **Response Wrappers:** Always wrap API responses in the globally established format. (e.g., `{ success: true, data: {...}, meta: {...} }`). Do not return raw arrays or unformatted strings unless explicitly required by a webhook or third-party integration.
* **Naming Conventions:** Strictly adhere to the established casing format for JSON payloads (e.g., always `snake_case` or always `camelCase`). Do not mix them.
* **REST/GraphQL/RPC:** Follow the existing endpoint design. Do not invent custom verbs or non-standard URL paths (e.g., use `POST /users`, do not use `POST /create_user`).

### 4. Database & State Mutations
Database interactions must be predictable, safe, and performant.

* **ORM/Query Builders:** Use the globally established database tool (e.g., Prisma, GORM, SQLAlchemy, raw SQL with pgx). Do not introduce a new query method.
* **Transactions:** When performing multiple related mutations (e.g., creating a User and an associated Account), you MUST wrap the operations in a database transaction to ensure atomicity.
* **No Hard Deletes:** Unless explicitly instructed, use "soft deletes" (e.g., setting a `deleted_at` timestamp) if that is the established pattern for the schema.
* **N+1 Queries:** Always actively prevent N+1 query problems by using the established eager loading, joins, or dataloader patterns.

### 5. Security & Validation
Never trust client input. Security layers must be applied consistently.

* **Input Validation:** ALL incoming data (Body, Query, Params) MUST be validated before reaching the Service layer using the established validation tool (e.g., Zod, Go Struct Tags, Pydantic).
* **Authentication/Authorization:** Apply the existing Auth middleware to protected routes. Do not write custom token verification logic inside the controller.
* **Secrets Management:** NEVER hardcode secrets, API keys, or database URIs. Always use environment variables or the established secrets manager.

### 6. Error Handling & Logging
Errors must be handled centrally and predictably.

* **No Silent Fails:** Never use empty `catch` blocks or ignore returned errors (e.g., in Go, always handle `if err != nil`).
* **Structured Errors:** Throw or return globally defined Error classes/types (e.g., `AppError(404, "User not found")`) rather than generic language errors.
* **Centralized Middleware/Interceptor:** Let the global error-handling middleware format the HTTP error response. Do not manually construct `res.status(500).send(...)` in every controller.
* **Logging:** Use the established logging library (e.g., Winston, Zap, Logrus). Do not use `console.log` or `print()` in production-bound code.

### 7. Dependency & Package Constraints (Guardrails)
* **STOP before adding packages:** Do not add new libraries (e.g., a new HTTP client, a new date parser) if an equivalent already exists in the dependency file (`package.json`, `go.mod`, `requirements.txt`).
* **Propose First:** If a feature genuinely requires a new backend dependency, explicitly ask the developer for approval before installation.

## Universal Frontend Consistency Protocol

**TARGET AUDIENCE:** Agentic AI Coding Assistants
**PURPOSE:** To enforce strict, zero-deviation architectural and visual consistency across the entire frontend ecosystem (Web, Mobile, Desktop) within this repository.

### 1. Core Directives (The "Consumer" Rule)
When generating, editing, or refactoring frontend code, you are a **CONSUMER** of the existing architecture, never an **INVENTOR**.
* **DO NOT** introduce new structural paradigms, state management libraries, or data-fetching patterns on a whim.
* Ensure code behavior and UI elements share the exact same logical and structural footprint across all screens.
* **If a pattern already exists in the codebase for a specific task, you MUST reuse it.**

### 2. The UI & Styling Layer
All platforms use a centralized token system. You are strictly forbidden from hardcoding raw visual values.

* **NO hardcoded Hex/RGB codes or arbitrary dimensions** (e.g., `#FF5733`, `w-[324px]`).
* **Web (Tailwind):** Use predefined styling classes from `tailwind.config.ts`.
* **Mobile/Desktop:** Use the global theme provider.
* **Component Reuse:** Never instantiate raw HTML or native elements if a system abstraction exists in the `/ui` or `/components` directory. If a component lacks a feature, update the base component (via a backward-compatible prop) rather than building a custom variant on the page level.

### 3. Data Fetching & Mutations
Data flows must remain predictable and uniform. Do not mix paradigms.

* **Server-Side Fetching (Next.js/SSR):** Use Server Components for all initial read operations.
* **Client-Side Fetching:** Use the globally established client-fetching library (e.g., SWR or React Query) for dynamic client data. Do not use raw `useEffect` with `fetch` for data hydration unless explicitly required by the established pattern.
* **Mutations:** Use established Server Actions or defined API routes. Always handle the loading state during a mutation and implement optimistic UI updates where applicable.
* **No Direct DB Access in Client:** Never write database queries (e.g., PostgreSQL, Prisma, Drizzle) inside client-side components.

### 4. State Management Architecture
State must be stored at the lowest possible tier required for its function.

1. **URL / Search Params (Tier 1):** Use the URL for shareable, persistent state (e.g., pagination `?page=2`, search filters, active tabs). Do not use local state for this.
2. **Local Component State (Tier 2):** Use `useState` or `useReducer` strictly for isolated, transient UI state (e.g., "is this specific dropdown open?").
3. **Global State (Tier 3):** Use the established global store (e.g., Zustand, Context API) ONLY for state that must be accessed across disparate parts of the app (e.g., User Session, Shopping Cart, Theme).
* **PROHIBITED:** Do not introduce Redux, MobX, or additional state libraries unless specifically directed.

### 5. Type Safety & Data Models
Strict typing is mandatory.

* **No `any`:** You are forbidden from using the `any` type or `@ts-ignore` comments to bypass type checking.
* **Centralized Models:** Use the globally defined interfaces and types (usually located in a `/types` or `/models` directory) for all database schemas and API payloads. Do not redefine types locally within a component.
* **Validation:** Use the established validation library (e.g., Zod, Yup) for all form inputs and API payload verification.

### 6. Error Handling & Loading States
You must account for all edge cases gracefully.

* **Loading:** Use the globally defined `<Skeleton>` loader or spinner. Do not write localized text like `"Loading..."`.
* **Errors:** Wrap operations in standard `try/catch` blocks. For UI presentation, use the global `<ErrorBoundary>` or `<ErrorFallback>` components.
* **Notifications:** For user-facing feedback (success/fail), trigger the global Toast notification system. Do not use native `alert()`.

### 7. Dependency & Package Constraints (Guardrails)
* **STOP before adding packages:** Do not add new npm packages (e.g., `axios`, `moment`, `lodash`) to solve a problem if native APIs (`fetch`, `Intl`, modern JS methods) or an already installed equivalent exists in `package.json`.
* **Propose First:** If a feature cannot be built without a new dependency, explicitly ask the developer for approval before running `npm install`.

<!-- context7 -->
Use Context7 MCP to fetch current documentation whenever the user asks about a library, framework, SDK, API, CLI tool, or cloud service -- even well-known ones like React, Next.js, Prisma, Express, Tailwind, Django, or Spring Boot. This includes API syntax, configuration, version migration, library-specific debugging, setup instructions, and CLI tool usage. Use even when you think you know the answer -- your training data may not reflect recent changes. Prefer this over web search for library docs.

Do not use for: refactoring, writing scripts from scratch, debugging business logic, code review, or general programming concepts.

## Steps

1. Always start with `resolve-library-id` using the library name and the user's question, unless the user provides an exact library ID in `/org/project` format
2. Pick the best match (ID format: `/org/project`) by: exact name match, description relevance, code snippet count, source reputation (High/Medium preferred), and benchmark score (higher is better). If results don't look right, try alternate names or queries (e.g., "next.js" not "nextjs", or rephrase the question). Use version-specific IDs when the user mentions a version
3. `query-docs` with the selected library ID and the user's full question (not single words)
4. Answer using the fetched docs
<!-- context7 -->
