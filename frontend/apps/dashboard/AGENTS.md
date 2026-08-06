# AGENTS.md: Email Ingestion Gateway Dashboard Context

## 1. Project Mission & Overview
The **Email Ingestion Gateway Dashboard** (`frontend/apps/dashboard`) is a single-page application (SPA) designed to manage tenants, view inbound email logs, configure webhook endpoints, manage S3 attachment storage credentials, and monitor SMTP gateway activity.

## 2. Tech Stack & Core Dependencies
* **Framework:** **Vue 3** (Composition API with `<script setup>`)
* **Build Tool & Dev Server:** **Vite 8**
* **UI Component Library:** **Vuetify 4** (`vuetify`, `vite-plugin-vuetify`, `@mdi/font`, `@fontsource/roboto`, `@fontsource/roboto-mono`)
* **Routing:** **Vue 3 File-Based Router** (`vue-router` v5 using `vue-router/vite` / `vue-router/auto-routes`)
* **State Management:** **Pinia 3** (`pinia`)
* **Styling & CSS Utility:** **UnoCSS** with Vuetify preset (`unocss-preset-vuetify`, SASS/SCSS)
* **Code Quality & Linting:** **ESLint** (`eslint-config-vuetify`)
* **Package Manager:** **`pnpm`**

## 3. Mandatory Router Rule: File-Based Routing
> [!IMPORTANT]
> **Strict Requirement:** Vue 3 Router MUST use **file-based routing**. All application pages and routes are defined strictly by creating Vue single-file components (SFCs) within the `src/pages/` directory.

### File-Based Routing Conventions (`src/pages/`)
Routes are automatically synthesized by Vite via `vue-router/vite` and imported in `src/router/index.js` via `import { routes } from 'vue-router/auto-routes'`.

* **Index Route (`/`):** `src/pages/index.vue`
* **Static Routes (`/webhooks`):** `src/pages/webhooks/index.vue` or `src/pages/webhooks.vue`
* **Dynamic Parameter Routes (`/webhooks/:id`):** `src/pages/webhooks/[id].vue`
* **Nested Routes (`/settings/profile`):** `src/pages/settings/profile.vue`
* **Catch-All / 404 Route (`/:pathMatch(.*)*`):** `src/pages/[...all].vue`

**DO NOT** manually construct static route objects or edit array definitions in `src/router/index.js`. Any new page or path MUST be added by creating the appropriate SFC inside `src/pages/`.

## 4. Directory Mental Model
* **`src/pages/`**: File-based router views (Pages). Each `.vue` file maps directly to a URL route.
* **`src/components/`**: Reusable component UI primitives and domain widgets (NOT routed pages).
* **`src/stores/`**: Pinia state management stores (e.g., app settings, tenant state, user auth).
* **`src/plugins/`**: Plugin bootstrap logic (`vuetify.js`, `index.js`). Registers Vuetify, Pinia, and the Router.
* **`src/router/`**: Router initialization file (`index.js`). Imports auto-generated routes from `vue-router/auto-routes`.
* **`src/styles/`**: Global SASS styling files (`main.scss`, `settings.scss`).
* **`src/assets/`**: Static image and design assets.
* **`public/`**: Public static assets served at the root domain.

## 5. Development Standards & Best Practices
* **Component Paradigm:** Always use Vue 3 **Composition API** with `<script setup>`. **DO NOT** use the Options API.
* **API Prefix & Dual Principal Model (ADR-0002 & ADR-0004):** The Dashboard SPA calls the Human/User Dashboard API using the **`/app/v1/...`** base path (specifically targeting `/app/v1/applications/:app_id/...`), authenticated via `UserAuth` OIDC Bearer tokens. **DO NOT** use `/api/v1/...` (which is reserved for M2M tenant API integrations).
* **Styling & Layout:** Utilize Vuetify grid system (`v-container`, `v-row`, `v-col`, `v-card`, etc.) and utility classes from UnoCSS / Vuetify presets. Maintain dark/light theme consistency.
* **Package Management:** Use `pnpm` exclusively for managing dependencies and executing project scripts.
* **State Management:** Use Pinia composable stores (`defineStore`) located in `src/stores/` for managing global state (e.g., active application ID, webhooks, auth JWTs).

## 6. Operational Commands
Execute all commands from `frontend/apps/dashboard` (or via monorepo scripts using pnpm):

* **Install Dependencies:**
  ```bash
  pnpm install
  ```
* **Start Development Server:**
  ```bash
  pnpm dev
  ```
* **Production Build:**
  ```bash
  pnpm build
  ```
* **Preview Build:**
  ```bash
  pnpm preview
  ```
* **Linting & Fixing:**
  ```bash
  pnpm lint
  pnpm lint:fix
  ```

## 7. Hard Constraints & Anti-Patterns for AI Agents
* **DO NOT** define routes manually in `src/router/index.js`. Always use the **`src/pages/`** directory for file-based routing.
* **DO NOT** use `/api/v1/...` endpoints for dashboard API requests. The Dashboard SPA MUST target **`/app/v1/...`** (specifically `/app/v1/applications/:app_id/...`) per ADR-0002 & ADR-0004.
* **DO NOT** use Options API (`export default { data() { ... } }`). Always write idiomatic Vue 3 `<script setup>`.
* **DO NOT** use `npm` or `yarn`. Always use **`pnpm`**.
* **DO NOT** place page-level route components inside `src/components/`. Keep pages in `src/pages/` and reusable UI components in `src/components/`.
* **DO NOT** bypass Vuetify design tokens with ad-hoc hardcoded styling when standard Vuetify color/theme props exist.
