# Popup Management System — API Documentation

This document describes the current REST API contract for the **Popup Management System** (`popup-manager-api`). It is generated from the live codebase and reflects the exact routes, DTOs, parameters, and error responses currently implemented.

---

## 1. API Overview

The API provides management capabilities for:
- **Authentication & User Profile**: Administrative login and session identity verification.
- **Websites**: Managing registered client domains, target platforms, and operational status.
- **Popups**: Configuring banners/modals, positioning, schedule windows, and associations with registered websites.

All application endpoints are versioned under:
```
/api/v1
```

A root informational endpoint is available at `GET /`:
- **Response**: `{"success": true, "message": "Popup Management API", "version": "v1"}`

---

## 2. Base URL

Default local development base URL:
```
http://localhost:8082/api/v1
```

*(Port is configured via the `APP_PORT` environment variable.)*

---

## 3. Authentication

The API uses **JSON Web Tokens (JWT)** for securing protected routes.

### Mechanism
- Tokens are issued upon successful authentication at `POST /api/v1/login`.
- Tokens are signed with HMAC-SHA256 (`HS256`) and contain the claims: `user_id`, `email`, `role`, and expiration `exp`.
- Expiration time is configured via `JWT_EXPIRE_HOURS`.

### Header Requirement
All protected endpoints require the HTTP `Authorization` request header:
```http
Authorization: Bearer <JWT_TOKEN>
```

### Access Matrix
| Access Level | Endpoints |
| :--- | :--- |
| **Public** | `GET /`, `POST /api/v1/login` |
| **Protected** | `GET /api/v1/profile`, all `/api/v1/websites` routes, all `/api/v1/popups` routes |

---

## 4. Common Response Format

Every endpoint returns a unified JSON envelope defined in `internal/common/response.go`.

### Standard Success (Action / Mutation)
```json
{
  "success": true,
  "message": "Website created successfully"
}
```
*(Note: When `data` is nil or empty, the `data` key is omitted.)*

### Standard Success with Data (Single Resource / Object)
```json
{
  "success": true,
  "message": "Website fetched successfully",
  "data": {
    "id": 1,
    "website_name": "Documentation Portal",
    "domain": "docs.example.com",
    "platform": "React",
    "status": true
  }
}
```

### Success with Pagination (List Queries)
```json
{
  "success": true,
  "message": "Websites fetched successfully",
  "data": [
    {
      "id": 1,
      "website_name": "Documentation Portal",
      "domain": "docs.example.com",
      "platform": "React",
      "status": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

## 5. Common Error Format

The API provides standardized error responses depending on whether the error is a general failure, HTTP contract error, or business validation failure.

### Generic Error (400, 401, 404, 409, 500)
Returned for missing authentication headers, not found entities, binding errors, or conflicts:
```json
{
  "success": false,
  "message": "resource not found"
}
```

### Field Validation Error (422 Unprocessable Entity)
Returned when domain or payload validation fails in business service layers (e.g., empty names, invalid platform, invalid position, invalid schedule time):
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "platform": "invalid platform",
    "domain": "invalid domain format"
  }
}
```

### Specific Error Statuses
- **400 Bad Request**: Malformed JSON syntax, Gin binding constraint violation, or non-numeric ID in path.
  ```json
  {
    "success": false,
    "message": "Invalid website ID"
  }
  ```
- **401 Unauthorized**: Missing/malformed `Authorization` header, invalid/expired token, or incorrect login credentials.
  ```json
  {
    "success": false,
    "message": "Authorization header is required"
  }
  ```
- **404 Not Found**: Resource does not exist or target ID not found during retrieval, update, or deletion.
  ```json
  {
    "success": false,
    "message": "resource not found"
  }
  ```
- **409 Conflict**: Unique constraint violation (such as duplicate website domain).
  ```json
  {
    "success": false,
    "message": "website domain already exists"
  }
  ```
- **422 Unprocessable Entity**: Semantic business validation failure, includes the `errors` map.
- **500 Internal Server Error**: Unexpected server-side failure.

---

## 6. Pagination

List endpoints support query parameters for pagination:

| Parameter | Type | Default | Constraints | Description |
| :--- | :--- | :--- | :--- | :--- |
| `page` | integer | `1` | Must be `> 0`; invalid values fall back to `1` | The page number to retrieve. |
| `limit` | integer | `10` | Must be `> 0`; maximum allowed is `100` | The number of records per page. |

### Pagination Envelope Details
- `page`: Current page number.
- `limit`: Applied limit.
- `total`: Total matching records count in the database.
- `total_pages`: Total number of pages (`ceil(total / limit)`).

### Example
```http
GET /api/v1/websites?page=2&limit=20
```

---

## 7. Search

List endpoints support substring pattern matching using the `search` query parameter.

| Parameter | Type | Description |
| :--- | :--- | :--- |
| `search` | string | Matches against designated module fields using SQL `LIKE %search%`. |

### Module Search Target Fields
- **Websites**: Matches against `website_name` OR `domain`.
- **Popups**: Matches against `title` OR `content`.

Search parameters work seamlessly in combination with pagination and sorting:
```http
GET /api/v1/popups?search=welcome&page=1&limit=10
```

---

## 8. Sorting

List endpoints support sorting with strict module-specific allowlists to prevent SQL injection and ensure predictable sorting.

| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `sort` | string | `id` | The field name to sort by (case-insensitive). |
| `order` | string | `desc` | Sort direction: `asc` or `desc` (case-insensitive). Defaults to `desc`. |

### Module Sort Allowlists

#### Websites Module
| Accepted Query Name | Database Column Mapped |
| :--- | :--- |
| `id` | `id` |
| `name` | `website_name` |
| `website_name` | `website_name` |
| `domain` | `domain` |
| `platform` | `platform` |
| `status` | `status` |
| `created_at` | `created_at` |

*Fallback: Any sort field not in this table defaults to `id`.*

#### Popups Module
| Accepted Query Name | Database Column Mapped |
| :--- | :--- |
| `id` | `id` |
| `title` | `title` |
| `position` | `position` |
| `status` | `status` |
| `start_time` | `start_time` |
| `end_time` | `end_time` |
| `created_at` | `created_at` |

*Fallback: Any sort field not in this table defaults to `id`.*

### Example
```http
GET /api/v1/websites?sort=domain&order=asc
```

---

## 9. Authentication Endpoints

### 9.1 User Login
Authenticates an administrator and returns a JWT access token.

- **HTTP Method**: `POST`
- **URL**: `/api/v1/login`
- **Authentication**: None (Public)
- **Headers**: `Content-Type: application/json`

#### Request Body
```json
{
  "email": "admin@example.com",
  "password": "your-password"
}
```

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `email` | string | Yes | Administrator email address. |
| `password` | string | Yes | Plaintext password. |

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "Administrator",
      "email": "admin@example.com",
      "role": "Super Admin"
    }
  }
}
```

#### Possible Errors
- **400 Bad Request**: Invalid JSON request payload.
- **401 Unauthorized**: Invalid credentials (`"invalid credentials"`).
- **500 Internal Server Error**: Failure during token generation.

---

### 9.2 Current User Profile
Returns claims and details of the currently authenticated administrator.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/profile`
- **Authentication**: Required (`Bearer <token>`)

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Profile fetched successfully",
  "data": {
    "user_id": 1,
    "email": "admin@example.com",
    "role": "Super Admin"
  }
}
```

#### Possible Errors
- **401 Unauthorized**: Missing token, invalid bearer format, or expired token.

---

## 10. Website Endpoints

All website endpoints require authentication.

### Supported Platforms
The `platform` field is validated against the following allowed values:
- `WordPress`
- `React`
- `Python`
- `Laravel`
- `NodeJS`
- `HTML`
- `Java`
- `VueJS`
- `Go`
- `Other`

### Domain Validation
The `domain` field must match standard domain syntax: `^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`.

---

### 10.1 Create Website
Registers a new client website.

- **HTTP Method**: `POST`
- **URL**: `/api/v1/websites`
- **Authentication**: Required

#### Request Body
```json
{
  "website_name": "Acme Store",
  "domain": "acme.example.com",
  "platform": "React"
}
```

| Field | Type | Rules | Description |
| :--- | :--- | :--- | :--- |
| `website_name` | string | Required, min 2, max 150 chars | Name of the website. |
| `domain` | string | Required, valid domain syntax, unique | Fully qualified domain name. |
| `platform` | string | Required, must be in supported list | Host platform framework. |

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Website created successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Missing required binding fields.
- **409 Conflict**: Domain already registered (`"website domain already exists"`).
- **422 Unprocessable Entity**: Validation failure (`website_name`, `platform`, or `domain` format).

---

### 10.2 List Websites
Retrieves a paginated list of websites.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/websites`
- **Authentication**: Required
- **Query Parameters**: `page`, `limit`, `search`, `sort`, `order`

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Websites fetched successfully",
  "data": [
    {
      "id": 1,
      "website_name": "Acme Store",
      "domain": "acme.example.com",
      "platform": "React",
      "status": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### 10.3 Get Website by ID
Retrieves single website details.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/websites/:id`
- **Authentication**: Required

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Website fetched successfully",
  "data": {
    "id": 1,
    "website_name": "Acme Store",
    "domain": "acme.example.com",
    "platform": "React",
    "status": true
  }
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric ID parameter.
- **404 Not Found**: Website does not exist (`"resource not found"`).

---

### 10.4 Update Website
Updates an existing website.

- **HTTP Method**: `PUT`
- **URL**: `/api/v1/websites/:id`
- **Authentication**: Required

#### Request Body
```json
{
  "website_name": "Acme Global Store",
  "domain": "acme-global.example.com",
  "platform": "React",
  "status": true
}
```

| Field | Type | Rules | Description |
| :--- | :--- | :--- | :--- |
| `website_name` | string | Required, min 2, max 150 chars | Updated website name. |
| `domain` | string | Required, valid domain syntax | Updated domain name (must not conflict with another website). |
| `platform` | string | Required, supported platform | Platform framework. |
| `status` | boolean | Optional | Website active status. |

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Website updated successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Invalid ID or binding errors.
- **404 Not Found**: Website ID not found.
- **409 Conflict**: Domain in use by another website.
- **422 Unprocessable Entity**: Business validation failures.

---

### 10.5 Delete Website
Soft-deletes a website record.

- **HTTP Method**: `DELETE`
- **URL**: `/api/v1/websites/:id`
- **Authentication**: Required

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Website deleted successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric ID parameter.
- **404 Not Found**: Website ID not found (`"resource not found"`).

---

## 11. Popup Endpoints

All popup endpoints require authentication.

### Supported Positions
The `position` field is validated against the following allowed values:
- `center`
- `top-banner`
- `bottom-left`
- `bottom-right`

### Date Scheduling Rules
- `start_time` and `end_time` are ISO 8601 / RFC 3339 formatted timestamps (or `null`).
- If **both** `start_time` and `end_time` are provided, `start_time` must be chronologically earlier than `end_time`.
- Otherwise, validation returns HTTP 422 with `"start time must be before end time"`.

---

### 11.1 Create Popup
Creates a new popup configuration associated with a website.

- **HTTP Method**: `POST`
- **URL**: `/api/v1/popups`
- **Authentication**: Required

#### Request Body
```json
{
  "website_id": 1,
  "title": "Spring Promotion",
  "content": "<p>Get 20% off with promo code SPRING20</p>",
  "position": "center",
  "status": true,
  "start_time": "2026-04-01T00:00:00Z",
  "end_time": "2026-04-30T23:59:59Z"
}
```

| Field | Type | Rules | Description |
| :--- | :--- | :--- | :--- |
| `website_id` | integer | Required | Associated website ID. |
| `title` | string | Required, min 2, max 255 chars | Title of the popup. |
| `content` | string | Required | HTML or text content. |
| `position` | string | Required, valid position | Placement on screen. |
| `status` | boolean | Optional (defaults `true`) | Active toggle. |
| `start_time` | string/null | Optional, RFC 3339 | Schedule start. |
| `end_time` | string/null | Optional, RFC 3339 | Schedule finish. |

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popup created successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Missing required binding fields.
- **404 Not Found**: Specified `website_id` does not exist (`"resource not found"`).
- **422 Unprocessable Entity**: Validation failure (empty title/content, invalid position, or invalid schedule order).

---

### 11.2 List Popups
Retrieves a paginated list of all popups across all websites.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/popups`
- **Authentication**: Required
- **Query Parameters**: `page`, `limit`, `search`, `sort`, `order`

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popups fetched successfully",
  "data": [
    {
      "id": 1,
      "website_id": 1,
      "title": "Spring Promotion",
      "content": "<p>Get 20% off with promo code SPRING20</p>",
      "position": "center",
      "status": true,
      "start_time": "2026-04-01T00:00:00Z",
      "end_time": "2026-04-30T23:59:59Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### 11.3 Get Popup by ID
Retrieves single popup details.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/popups/:id`
- **Authentication**: Required

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popup fetched successfully",
  "data": {
    "id": 1,
    "website_id": 1,
    "title": "Spring Promotion",
    "content": "<p>Get 20% off with promo code SPRING20</p>",
    "position": "center",
    "status": true,
    "start_time": "2026-04-01T00:00:00Z",
    "end_time": "2026-04-30T23:59:59Z"
  }
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric ID parameter.
- **404 Not Found**: Popup ID not found (`"resource not found"`).

---

### 11.4 List Popups by Website ID
Retrieves a paginated list of popups belonging to a specific website.

- **HTTP Method**: `GET`
- **URL**: `/api/v1/popups/website/:website_id`
- **Authentication**: Required
- **Query Parameters**: `page`, `limit`, `search`, `sort`, `order`

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popups fetched successfully",
  "data": [
    {
      "id": 1,
      "website_id": 1,
      "title": "Spring Promotion",
      "content": "<p>Get 20% off with promo code SPRING20</p>",
      "position": "center",
      "status": true,
      "start_time": "2026-04-01T00:00:00Z",
      "end_time": "2026-04-30T23:59:59Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric `website_id` path parameter.

---

### 11.5 Update Popup
Updates an existing popup.

- **HTTP Method**: `PUT`
- **URL**: `/api/v1/popups/:id`
- **Authentication**: Required

#### Request Body
```json
{
  "title": "Summer Promotion",
  "content": "<p>Summer discounts available now!</p>",
  "position": "top-banner",
  "status": true,
  "start_time": null,
  "end_time": null
}
```

*(Note: `website_id` cannot be reassigned via the update endpoint.)*

| Field | Type | Rules | Description |
| :--- | :--- | :--- | :--- |
| `title` | string | Required, min 2, max 255 chars | Updated title. |
| `content` | string | Required | Updated content. |
| `position` | string | Required, valid position | Placement position. |
| `status` | boolean | Optional | Status toggle. |
| `start_time` | string/null | Optional | Schedule start time. |
| `end_time` | string/null | Optional | Schedule end time. |

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popup updated successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric ID or missing required fields.
- **404 Not Found**: Popup ID not found.
- **422 Unprocessable Entity**: Business validation failures.

---

### 11.6 Delete Popup
Soft-deletes a popup record.

- **HTTP Method**: `DELETE`
- **URL**: `/api/v1/popups/:id`
- **Authentication**: Required

#### Successful Response (200 OK)
```json
{
  "success": true,
  "message": "Popup deleted successfully"
}
```

#### Possible Errors
- **400 Bad Request**: Invalid non-numeric ID.
- **404 Not Found**: Popup ID not found (`"resource not found"`).

---

## 12. HTTP Status Codes

| Status Code | Reason Phrase | Usage in API |
| :--- | :--- | :--- |
| **200** | OK | Successful GET requests, and successful mutations (Create, Update, Delete) via `common.Success`. |
| **400** | Bad Request | Invalid parameter types (e.g. non-numeric ID) or JSON payload binding failures. |
| **401** | Unauthorized | Missing/invalid `Authorization` header, invalid JWT token, or incorrect credentials during login. |
| **404** | Not Found | Target resource not found in database (`common.ErrNotFound`). |
| **409** | Conflict | Domain unique constraint duplicate collision (`common.ErrDuplicateDomain`). |
| **422** | Unprocessable Entity | Domain/business logic validation failures (returns `errors` map). |
| **500** | Internal Server Error | Server runtime failures, such as JWT signing failure or unhandled database errors. |

*(Note: HTTP 201 Created and HTTP 403 Forbidden are not currently emitted by the codebase.)*

---

## 13. Complete Endpoint Summary

| Method | Endpoint | Auth | Purpose |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | No | API version & health information |
| `POST` | `/api/v1/login` | No | Authenticate user and issue JWT |
| `GET` | `/api/v1/profile` | Yes | Get authenticated user profile |
| `POST` | `/api/v1/websites` | Yes | Create website |
| `GET` | `/api/v1/websites` | Yes | List websites (paginated, searchable, sortable) |
| `GET` | `/api/v1/websites/:id` | Yes | Get website by ID |
| `PUT` | `/api/v1/websites/:id` | Yes | Update website by ID |
| `DELETE` | `/api/v1/websites/:id` | Yes | Delete website by ID |
| `POST` | `/api/v1/popups` | Yes | Create popup |
| `GET` | `/api/v1/popups` | Yes | List popups (paginated, searchable, sortable) |
| `GET` | `/api/v1/popups/:id` | Yes | Get popup by ID |
| `GET` | `/api/v1/popups/website/:website_id` | Yes | List popups for a specific website |
| `PUT` | `/api/v1/popups/:id` | Yes | Update popup by ID |
| `DELETE` | `/api/v1/popups/:id` | Yes | Delete popup by ID |

---

## 14. Example Requests

### 14.1 Login
```bash
curl -X POST http://localhost:8082/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Password123!"
  }'
```

### 14.2 Fetch Profile
```bash
curl -X GET http://localhost:8082/api/v1/profile \
  -H "Authorization: Bearer <TOKEN>"
```

### 14.3 Create Website
```bash
curl -X POST http://localhost:8082/api/v1/websites \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "website_name": "Main Shop",
    "domain": "shop.example.com",
    "platform": "WordPress"
  }'
```

### 14.4 List Websites with Search & Sort
```bash
curl -X GET "http://localhost:8082/api/v1/websites?page=1&limit=10&search=shop&sort=domain&order=asc" \
  -H "Authorization: Bearer <TOKEN>"
```

### 14.5 Create Popup
```bash
curl -X POST http://localhost:8082/api/v1/popups \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "website_id": 1,
    "title": "Holiday Special",
    "content": "<h1>Save 30%</h1>",
    "position": "bottom-right",
    "status": true,
    "start_time": "2026-12-01T00:00:00Z",
    "end_time": "2026-12-25T23:59:59Z"
  }'
```

### 14.6 List Popups by Website
```bash
curl -X GET "http://localhost:8082/api/v1/popups/website/1?page=1&limit=5&sort=created_at&order=desc" \
  -H "Authorization: Bearer <TOKEN>"
```

---

## 15. Example Responses

### 15.1 Successful Login
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk3NTU2MDAsInVzZXJfaWQiOjEsImVtYWlsIjoiYWRtaW5AZXhhbXBsZS5jb20iLCJyb2xlIjoiU3VwZXIgQWRtaW4ifQ.signature",
    "user": {
      "id": 1,
      "name": "Administrator",
      "email": "admin@example.com",
      "role": "Super Admin"
    }
  }
}
```

### 15.2 Successful Profile
```json
{
  "success": true,
  "message": "Profile fetched successfully",
  "data": {
    "user_id": 1,
    "email": "admin@example.com",
    "role": "Super Admin"
  }
}
```

### 15.3 Single Website Response
```json
{
  "success": true,
  "message": "Website fetched successfully",
  "data": {
    "id": 1,
    "website_name": "Main Shop",
    "domain": "shop.example.com",
    "platform": "WordPress",
    "status": true
  }
}
```

### 15.4 Single Popup Response
```json
{
  "success": true,
  "message": "Popup fetched successfully",
  "data": {
    "id": 1,
    "website_id": 1,
    "title": "Holiday Special",
    "content": "<h1>Save 30%</h1>",
    "position": "bottom-right",
    "status": true,
    "start_time": "2026-12-01T00:00:00Z",
    "end_time": "2026-12-25T23:59:59Z"
  }
}
```

### 15.5 Paginated List Response
```json
{
  "success": true,
  "message": "Popups fetched successfully",
  "data": [
    {
      "id": 1,
      "website_id": 1,
      "title": "Holiday Special",
      "content": "<h1>Save 30%</h1>",
      "position": "bottom-right",
      "status": true,
      "start_time": "2026-12-01T00:00:00Z",
      "end_time": "2026-12-25T23:59:59Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

### 15.6 Validation Error Response (422 Unprocessable Entity)
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "start_time": "start time must be before end time"
  }
}
```

### 15.7 Resource Not Found (404 Not Found)
```json
{
  "success": false,
  "message": "resource not found"
}
```

### 15.8 Duplicate Domain Conflict (409 Conflict)
```json
{
  "success": false,
  "message": "website domain already exists"
}
```
