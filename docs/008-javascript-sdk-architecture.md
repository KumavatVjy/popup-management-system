# Phase 3 — JavaScript SDK Architecture & Public API Design

This document defines the architectural blueprint and public API specification for the **Popup Management JavaScript SDK** (`@popup-manager/sdk`). 

The SDK is designed to be an ultra-lightweight, framework-agnostic client library that enables browser-based applications to seamlessly fetch, filter, schedule, and render promotional and informational popups managed by the central Popup Management System.

---

## 1. SDK Goals

The SDK serves as the client-side integration layer across diverse digital properties, including:
- `definedge.com` (Corporate portal)
- `definedgesecurities.com` (Trading & securities platform)
- `gurukul.definedgesecurities.com` (Educational portal)
- Arbitrary client websites (WordPress, React, Vue, Angular, or static HTML)

### Core Architectural Principles
1. **Framework Independence**: Zero dependencies on React, Vue, jQuery, or WordPress. Operates directly on standard Web APIs.
2. **Host Site Safety**: Fails silently without impacting host website operations. Zero global style bleed or layout corruption.
3. **Strict Credential Separation**: Runs in untrusted client browsers. Never receives or transmits administrative JWTs, secrets, or privileged credentials.
4. **Performance & Lightweight Footprint**: Target bundle size `< 10 KB` (minified/gzipped). Asynchronous and non-blocking.
5. **Accessibility by Default**: Compliant with WCAG standards, featuring focus trapping, ARIA roles, and keyboard navigation (`Escape` dismissal).

---

## 2. Architecture & Layer Separation

To ensure modularity, testability, and long-term maintainability, the SDK is partitioned into two distinct conceptual layers:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        PopupManager (Public API)                       │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
          ┌─────────────────────────┴─────────────────────────┐
          ▼                                                   ▼
┌───────────────────────────────────┐       ┌───────────────────────────────────┐
│       Layer A: API Client         │       │     Layer B: Popup Runtime        │
├───────────────────────────────────┤       ├───────────────────────────────────┤
│ • Transport & Network (Fetch API) │       │ • Configuration & Lifecycle State │
│ • Public Identity (websiteKey)    │       │ • Client Eligibility Filtering    │
│ • Response Parsing & Validation   │       │ • Shadow DOM & CSS Containment    │
│ • Safe Error Normalization        │       │ • Viewport Positioning System     │
│ • Zero Secret Handling            │       │ • DOM Event Handling & Cleanup    │
│ • Request Timeouts & AbortSignal  │       │ • Public Event Emitter (Pub/Sub)  │
└───────────────────────────────────┘       └───────────────────────────────────┘
```

### Why This Separation is Critical
- **Separation of Concerns**: Transport and network mechanics (HTTP headers, query parameters, deserialization) remain decoupled from UI rendering, layout calculation, and DOM manipulation.
- **Independent Testability**: The API Client can be unit-tested in Node.js without DOM mocks; the Runtime can be verified using mock payloads without network dependencies.
- **Extensibility**: Alternative presentation modes (e.g., headless data-only consumption in single-page apps) can reuse the API Client without loading the UI renderer.

---

## 3. Layer A: API Client Layer

The **API Client Layer** is responsible for all outbound communication with the central Popup Management API.

### Responsibilities
- Constructing target URLs against the configured `apiBaseUrl`.
- Passing the public integration credential (`websiteKey`).
- Handling timeouts and request cancellations via standard `AbortController`.
- Validating the received HTTP status and JSON envelope.
- Transforming server responses into standardized client-side data structures.
- Normalizing network or HTTP errors into safe internal structures that never leak infrastructure details.

### Security Boundary
- The API Client operates exclusively against the **Public Delivery Endpoint**.
- It **never** interacts with administrative routes (`/api/v1/login`, `/api/v1/websites`, etc.) and does not manage or store JWT tokens.

---

## 4. Layer B: Popup Runtime Layer

The **Popup Runtime Layer** orchestrates client-side state, eligibility rules, and presentation.

### Responsibilities
- **Lifecycle Management**: Transitions through `UNINITIALIZED` → `CONFIGURED` → `LOADING` → `READY` → `ACTIVE` → `DESTROYED`.
- **Eligibility Evaluation**: Validates date scheduling (`start_time`, `end_time`) and status flags against client time.
- **DOM Encapsulation**: Mounts a container with Shadow DOM isolation to guarantee total CSS immunity from the host page.
- **Position Coordinate Mapping**: Renders modals, banners, and toast drawers according to backend positioning directives.
- **User Interaction**: Listens for dismiss buttons, backdrop taps, and the `Escape` key, triggering coordinated exit animations and event notifications.

---

## 5. Website Identification Strategy (`websiteKey` vs `website_id`)

### The Flaw of Using Database Primary Keys (`website_id`)
The current backend routes reference numeric primary keys (e.g., `website_id = 1, 2, 3`). Exposing raw database IDs in browser script tags introduces significant liabilities:
1. **Enumeration Attacks**: Sequential integers permit attackers or competitors to easily discover other tenant configurations by incrementing IDs.
2. **Database Coupling**: Primary keys tightly couple public website integrations with internal database schema mechanics.
3. **No Revocation / Rotation**: If a key is abused, changing an auto-incrementing database ID requires breaking database foreign keys across multiple tables.

### The Production Solution: `websiteKey`
The SDK designs around a dedicated public string identifier:
```javascript
PopupManager.init({
  websiteKey: "wg_live_9f83b2e1a74c"
});
```
- **Structure**: Prefixed, cryptographically random, non-sequential string (e.g., `wg_live_<hash>` or `pub_wk_<hash>`).
- **Scope**: Public credential granting read-only access strictly to active popups configured for that specific website.
- **Lifecycle**: Rotatable and deactivatable from the admin dashboard without altering database row IDs.

---

## 6. Public Delivery API Requirements (Backend Prerequisite)

### The Existing Backend Gap
The existing backend endpoint:
```http
GET /api/v1/popups/website/{website_id}
```
requires an **Admin JWT** in the `Authorization: Bearer <token>` header. 

Browser JavaScript executed by anonymous site visitors **must never** have access to an admin JWT. Therefore, the SDK cannot consume this endpoint directly.

### The Required Public Delivery Route
A dedicated, public, read-only delivery endpoint must be introduced to the backend before SDK runtime integration:

```http
GET /api/v1/public/popups?website_key={websiteKey}
```
*(Or via request header `X-Website-Key: {websiteKey}`)*

### Contract Requirements for the Public Delivery Endpoint
| Attribute | Specification |
| :--- | :--- |
| **Authentication** | None (Public read-only). Verified solely by `website_key`. |
| **CORS Policy** | Enabled for cross-origin browser requests (`Origin: *` or registered domain whitelist). |
| **Server-Side Filtering** | Only returns popups where `status = true` and `deleted_at IS NULL`. |
| **Sanitized Response DTO** | Excludes all internal fields (`created_by`, `deleted_at`, internal user emails, DB IDs). |
| **Payload Structure** | Returns `id`, `title`, `content`, `position`, `start_time`, `end_time`. |

```json
{
  "success": true,
  "message": "Popups fetched successfully",
  "data": [
    {
      "id": 12,
      "title": "Welcome Offer",
      "content": "<h2>Get 20% off</h2><p>Use code WELCOME20</p>",
      "position": "center",
      "start_time": "2026-04-01T00:00:00Z",
      "end_time": "2026-04-30T23:59:59Z"
    }
  ]
}
```

---

## 7. Authentication & Security Model

```
┌────────────────────────────────────────────────────────┐
│                   Security Boundary                    │
├──────────────────────────┬─────────────────────────────┤
│      ADMIN CONSOLE       │      CLIENT SDK RUNTIME     │
├──────────────────────────┼─────────────────────────────┤
│ • Private Environment    │ • Untrusted Browser Runtime │
│ • Administrator Login    │ • Anonymous Site Visitors   │
│ • Full CRUD Operations   │ • Read-Only Fetch           │
│ • Protected by JWT       │ • Identified by websiteKey  │
│ • Authorization Header   │ • Zero Privileged Secrets   │
└──────────────────────────┴─────────────────────────────┘
```

### Hard Rules
1. **Never** include the backend `JWT_SECRET`, database passwords, or admin passwords in client-side code.
2. **Never** initiate an admin login flow (`POST /api/v1/login`) from the browser SDK.
3. The browser SDK treats all client-side storage (memory, DOM) as accessible to the end user.

---

## 8. Popup Eligibility Model

The SDK must evaluate whether a retrieved popup is eligible for display based on status and schedule parameters.

### Eligibility Decision Table
| `status` | `start_time` | `end_time` | Eligibility Condition |
| :---: | :---: | :---: | :--- |
| `false` | Any | Any | **Ineligible** (Never displayed). |
| `true` | `null` | `null` | **Eligible** (Always active). |
| `true` | Provided | `null` | **Eligible** if `now >= start_time`. |
| `true` | `null` | Provided | **Eligible** if `now <= end_time`. |
| `true` | Provided | Provided | **Eligible** if `start_time <= now <= end_time`. |

### Server-Side vs. Client-Side Evaluation
- **Primary Defense (Server-Side)**: The public delivery API should filter out expired or future popups before transmitting the payload to save bandwidth and prevent pre-release content leaks.
- **Secondary Defense (Client-Side SDK)**: The SDK runtime independently re-validates the schedule against the local client clock. This handles edge cases where a user keeps a single-page session open across a scheduled cutoff time.

---

## 9. Position System

The SDK strictly respects the four frozen backend positions from Phase 1/Phase 2:

| Backend Position | Visual Behavior | CSS Layout Mapping |
| :--- | :--- | :--- |
| `center` | Centered modal dialog with semi-opaque backdrop overlay | Fixed viewport center: `top: 50%; left: 50%; transform: translate(-50%, -50%);` |
| `top-banner` | Full-width sticky header banner spanning the top edge | Fixed viewport top: `top: 0; left: 0; width: 100%;` |
| `bottom-left` | Floating toast/drawer card pinned to lower-left corner | Fixed corner: `bottom: 24px; left: 24px; max-width: 420px;` |
| `bottom-right` | Floating toast/drawer card pinned to lower-right corner | Fixed corner: `bottom: 24px; right: 24px; max-width: 420px;` |

*Note: No unapproved positions (`top-center`, `bottom-center`, `top-left`) will be introduced.*

---

## 10. Rendering Strategy & CSS Containment

When injecting HTML into third-party websites, style collisions are the primary failure point.

### Containment Strategy Comparison
| Strategy | Pros | Cons | Verdict |
| :--- | :--- | :--- | :---: |
| **Namespaced CSS** (`.popup-manager-*`) | Simple, universally supported | Vulnerable to aggressive global CSS (e.g., `* { all: unset; }` or typography rules) | Inadequate |
| **`<iframe>` Embed** | 100% total style isolation | Heavy, impossible to auto-size dynamically, breaks responsive modals | Inadequate |
| **Shadow DOM** (`attachShadow({ mode: 'open' })`) | Complete style encapsulation, dynamic height, accessible | Requires modern browser (all modern browsers support it) | **SELECTED** |

### Selected Architecture: Shadow DOM Containment
```
Host Page DOM (definedge.com)
  │
  └── <div id="popup-manager-host-root"> (Injected by SDK)
        │
        └── #shadow-root (open)
              │
              ├── <style>/* Encapsulated SDK Styles */</style>
              ├── <div class="pm-overlay"></div>
              └── <div class="pm-popup-card pm-position-center" role="dialog">
                    ├── <button class="pm-close-btn" aria-label="Close popup">&times;</button>
                    └── <div class="pm-content">...</div>
                  </div>
```

---

## 11. Multiple-Popup Strategy

The backend can return multiple active popups for a given website. Because the current API contract does not feature an explicit `priority` field, the SDK must handle this deterministically:

1. **Initial Rule (v1)**: **Single active popup at any time.**
2. **Selection Heuristic**: The SDK selects the **first eligible popup** returned by the server.
3. **Queueing Policy**: If multiple popups are eligible, only one is rendered. Once dismissed, subsequent popups are not forcibly popped up in the same session to avoid intrusive user experiences.
4. **Future Extension**: When the backend introduces a priority order or placement category, the runtime will adopt position-based concurrent display (e.g., one `top-banner` + one `bottom-right` toast simultaneously).

---

## 12. Lifecycle State Machine

```
[ UNINITIALIZED ]
       │  PopupManager.init(config)
       ▼
[ CONFIGURED ]
       │  autoLoad: true OR PopupManager.load()
       ▼
[ LOADING ]
       │  HTTP response received
       ▼
[ READY ] ── (No eligible popups) ──► [ IDLE ]
       │  autoShow: true OR PopupManager.show()
       ▼
[ ACTIVE / SHOWN ]
       │  User dismisses OR PopupManager.hide()
       ▼
[ HIDDEN ]
       │  PopupManager.destroy()
       ▼
[ DESTROYED ]
```

---

## 13. Public API Design

The SDK exposes a single global entry point: `window.PopupManager` (or default ESM export `PopupManager`).

```javascript
import PopupManager from '@popup-manager/sdk';

// 1. Initialization
PopupManager.init({
  websiteKey: 'wg_live_9f83b2e1a74c',
  apiBaseUrl: 'https://api.example.com',
  autoLoad: true,
  autoShow: true,
  debug: false
});

// 2. Event Listeners
PopupManager.on('popup:shown', (popup) => {
  console.log('Popup is now visible:', popup.id);
});

PopupManager.on('popup:closed', (popup) => {
  console.log('User closed popup:', popup.id);
});

// 3. Programmatic Controls
PopupManager.hide();
PopupManager.refresh();
PopupManager.destroy();
```

### Public Methods
| Method | Parameters | Return | Description |
| :--- | :--- | :--- | :--- |
| `init(config)` | `config: SDKConfig` | `void` | Configures credentials and runtime preferences. |
| `load()` | None | `Promise<Popup[]>` | Explicitly fetches popups from the API. |
| `show(popupId?)` | `popupId?: number` | `void` | Renders and displays an eligible popup. |
| `hide()` | None | `void` | Hides the currently active popup with animation. |
| `refresh()` | None | `Promise<void>` | Clears cached data, refetches, and re-evaluates display. |
| `destroy()` | None | `void` | Teardown: removes Shadow DOM, unbinds listeners, resets state. |
| `on(event, callback)`| `event: string, cb: Function` | `void` | Subscribes to runtime lifecycle events. |
| `off(event, callback)`| `event: string, cb: Function` | `void` | Unsubscribes from runtime lifecycle events. |

---

## 14. Configuration Schema (`SDKConfig`)

```typescript
interface SDKConfig {
  websiteKey: string;     // Required: Public website identifier
  apiBaseUrl?: string;    // Optional: Defaults to production API endpoint
  autoLoad?: boolean;     // Optional: Default true (fetches popups on init)
  autoShow?: boolean;     // Optional: Default true (renders immediately once loaded)
  debug?: boolean;        // Optional: Default false (enables diagnostic logging)
}
```

---

## 15. Event System

A lightweight pub/sub mechanism allows host pages to hook into popup interactions:

| Event Name | Payload | Emitted When |
| :--- | :--- | :--- |
| `popup:loaded` | `{ popups: Popup[] }` | Popups have been successfully fetched from the API. |
| `popup:shown` | `{ popup: Popup }` | A popup has completed its entrance animation and is visible. |
| `popup:hidden` | `{ popup: Popup }` | A popup was hidden programmatically via `PopupManager.hide()`. |
| `popup:closed` | `{ popup: Popup, reason: string }` | A popup was dismissed by the user (close button, backdrop, or Escape). |
| `popup:error` | `{ error: Error, context: string }` | An internal network, parsing, or DOM error occurred. |

---

## 16. Error Handling & Host Website Safety

### Safety Principles
1. **Never Crash the Host**: An unhandled exception in an auxiliary tool (like popups) must never break checkout flows, navigation, or critical site scripts.
2. **Graceful Degradation**: If network requests fail, timeout, or return 4xx/5xx status codes, the SDK logs internally (if `debug: true`), emits `popup:error`, and quietly ceases execution.
3. **Promise Rejection Containment**: All internal asynchronous promises are caught and handled.

---

## 17. Debug Mode

- **`debug: false` (Default)**: Silent operation. No `console.log` noise in host developer consoles.
- **`debug: true`**: Outputs structured diagnostics prefixed with `[PopupManager]`:
  ```
  [PopupManager] Initialized with websiteKey: wg_live_***
  [PopupManager] Fetching popups from: https://api.example.com/api/v1/public/popups
  [PopupManager] Found 2 popups; 1 eligible for display (Position: center)
  [PopupManager] Mounted Shadow DOM container
  ```
- **Security Check**: Even in debug mode, authorization secrets, passwords, or full headers are **never** logged.

---

## 18. Browser Support & Compatibility

- **Target Engines**: All modern evergreen browsers (Chrome, Edge, Safari, Firefox, iOS Safari, Android Chrome).
- **Language Level**: ECMAScript 2020 (`ES2020`).
- **Prerequisites**: Native support for `Fetch API`, `Promise`, `Custom Elements`, and `Shadow DOM`.
- **Zero Heavy Polyfills**: To maintain ultra-lightweight size, legacy IE11 support is explicitly excluded.

---

## 19. Package & Distribution Strategy

The SDK will be distributed in two formats:

```
popup-manager-sdk/
├── dist/
│   ├── popup-manager.esm.js     # ES Module for bundlers (React, Next.js, Vite, Webpack)
│   ├── popup-manager.umd.js     # UMD bundle for AMD/CommonJS
│   └── popup-manager.min.js     # IIFE bundle for direct script tags (WordPress, Webflow)
```

### Loading via CDN (WordPress / Static HTML)
```html
<script src="https://cdn.example.com/sdk/v1/popup-manager.min.js" async></script>
<script>
  window.addEventListener('load', function() {
    PopupManager.init({
      websiteKey: 'wg_live_9f83b2e1a74c'
    });
  });
</script>
```

### Loading via NPM (React / Next.js / Vue)
```bash
npm install @popup-manager/sdk
```

---

## 20. Versioning Strategy

- **Semantic Versioning (`MAJOR.MINOR.PATCH`)**:
  - `MAJOR`: Breaking changes to public API methods or event payload structures.
  - `MINOR`: New features (e.g., new position types, animation options) without breaking existing integrations.
  - `PATCH`: Bug fixes, CSS isolation improvements, accessibility patches.
- **Independence from API Version**: The SDK version (`v1.0.0`) operates independently from the backend REST version (`/api/v1`).

---

## 21. Caching Strategy

1. **Phase 1 Approach**: **In-memory cache per page lifecycle**.
   - Popups are fetched once on `init()` or `load()`.
   - Stored in runtime memory during the user's visit.
   - Cleared on page reload or via `PopupManager.refresh()`.
2. **Avoid LocalStorage in v1**: Storing popup payloads in `localStorage` creates risks of stale campaigns being displayed after an admin deactivates them in the dashboard.
3. **HTTP Cache**: The public backend delivery endpoint will supply standard HTTP caching headers (`ETag`, `Cache-Control: max-age=60`).

---

## 22. Cross-Origin Resource Sharing (CORS)

Browser clients will request popup configurations from a centralized API domain (e.g., `api.popupmanager.com` called from `definedge.com`).
- The public delivery API endpoint must configure CORS headers:
  ```http
  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, OPTIONS
  Access-Control-Allow-Headers: Content-Type, X-Website-Key
  ```
- Because the delivery endpoint is strictly read-only and unauthenticated, wildcard or verified domain access is completely secure.

---

## 23. Accessibility (a11y) Standards

1. **ARIA Roles**:
   - Modal popups render with `role="dialog"` and `aria-modal="true"`.
   - Banners and toasts render with `role="region"` or `role="alert"`.
2. **Keyboard Navigation**:
   - Pressing the `Escape` key immediately closes the active popup.
   - Tab focus is trapped within the active modal until dismissed.
   - Focus is returned to the previously active element upon closure.
3. **Screen Readers**:
   - `aria-labelledby` binds to the popup title element.
   - Dismiss buttons feature explicit `aria-label="Close popup"`.

---

## 24. Performance Optimization

1. **Non-Blocking Loading**: The SDK runs entirely asynchronously without blocking browser HTML parsing or the First Contentful Paint (FCP).
2. **No External Stylesheet Requests**: All CSS is bundled directly within the JS bundle and injected directly into the Shadow Root, eliminating secondary render-blocking HTTP roundtrips.
3. **Zero Layout Thrashing**: Container dimensions and positions are calculated using CSS transforms rather than continuous JavaScript window-resize listeners.

---

## 25. Future Analytics Architecture

While v1 focuses strictly on delivery and rendering, internal hooks are architected to support telemetry in subsequent phases:
- `trackImpression(popupId)`: Triggered when `popup:shown` fires.
- `trackClick(popupId, eventTarget)`: Triggered when an interactive element inside the popup is clicked.
- `trackDismiss(popupId, reason)`: Triggered when `popup:closed` fires.

---

## 26. Framework Integration Patterns

### WordPress
Enqueue via standard `functions.php`:
```php
function enqueue_popup_manager_sdk() {
    wp_enqueue_script(
        'popup-manager-sdk',
        'https://cdn.example.com/sdk/v1/popup-manager.min.js',
        array(),
        '1.0.0',
        true
    );
    wp_add_inline_script(
        'popup-manager-sdk',
        'PopupManager.init({ websiteKey: "wg_live_wordpress_site" });'
    );
}
add_action('wp_enqueue_scripts', 'enqueue_popup_manager_sdk');
```

### React / Next.js
Hook-based integration with clean teardown:
```tsx
import { useEffect } from 'react';
import PopupManager from '@popup-manager/sdk';

export function PopupProvider({ websiteKey }: { websiteKey: string }) {
  useEffect(() => {
    PopupManager.init({
      websiteKey,
      autoLoad: true,
      autoShow: true
    });

    return () => {
      PopupManager.destroy();
    };
  }, [websiteKey]);

  return null;
}
```

---

## 27. Proposed SDK Directory Structure

```
popup-manager-sdk/
├── src/
│   ├── index.ts                 # Public entry point & singleton export
│   ├── config.ts                # Configuration parsing and default options
│   ├── client/
│   │   ├── api-client.ts        # HTTP Fetch wrapper & response normalizer
│   │   └── types.ts             # Public delivery DTO interfaces
│   ├── runtime/
│   │   ├── manager.ts           # State machine and programmatic controller
│   │   ├── eligibility.ts       # Date and status rule evaluator
│   │   └── state.ts             # Internal reactive runtime state
│   ├── renderer/
│   │   ├── shadow-dom.ts        # Container injection & Shadow DOM wrapper
│   │   ├── templates.ts         # Modal, banner, and toast HTML builders
│   │   ├── styles.ts            # Encapsulated CSS style definitions
│   │   └── positions.ts         # Coordinates and viewport alignment
│   ├── events/
│   │   └── event-emitter.ts     # Internal event bus & public pub/sub
│   └── utils/
│       ├── logger.ts            # Safe debug console logger
│       └── dom.ts               # Focus trap and accessibility helpers
├── test/                        # Unit tests (Vitest / Jest)
├── rollup.config.mjs            # Multi-target bundler config (ESM, UMD, IIFE)
├── tsconfig.json                # TypeScript compilation config
├── package.json                 # Package metadata and build scripts
└── README.md                    # Integration guide
```

---

## 28. Open Backend Prerequisites

Before Phase 3 Lesson 2 (SDK Implementation) can proceed to live integration, the backend must implement the following architectural prerequisites:

1. **Website Public Key Management**:
   - Add a `website_key` column to the `websites` table (`VARCHAR(64)`, unique, indexed).
   - Generate secure, random public keys (e.g., `wg_live_<random_bytes>`) upon website registration.
2. **Public Delivery Endpoint**:
   - Implement `GET /api/v1/public/popups?website_key={key}`.
   - Must be unauthenticated (accessible without JWT).
   - Must return only sanitized public fields (`id`, `title`, `content`, `position`, `start_time`, `end_time`).
3. **Server-Side Active Filtering**:
   - The public endpoint must automatically query `WHERE status = true AND deleted_at IS NULL`.
   - Optionally filter by current timestamp (`start_time <= NOW() AND (end_time IS NULL OR end_time >= NOW())`).
4. **CORS Headers**:
   - Configure Gin CORS middleware on the public delivery route to allow cross-origin browser requests.
