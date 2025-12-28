# Detailed Design Document (Issue Tracking Desktop Application ratta)

---

## DD-SCOPE-001 Purpose

This document is the detailed design for the issue tracking desktop application ratta, based on the following documents:

* Requirements Specification (要求仕様書.md)
* Requirements Definition (要件定義書.md)
* Basic Design (基本設計書.md)

The primary goal is to provide a design that is easy for code-generation AI to consume (clear structure, explicit DTOs, explicit responsibilities).

## DD-SCOPE-002 Scope

In scope:

* Directory and file structure under the Project Root
* Data model (Issue JSON, attachments)
* Backend (Go) functions exposed via Wails binding
* Frontend (Vue 3 + Vuetify 3) screens and store design (Pinia)
* Validation, error handling, logging, and test approach

Out of scope:

* Exact UI layout details such as pixel-perfect design
* Exact ordering rules for JSON object keys

## DD-SCOPE-003 Notes and assumptions

* JSON key order is not fixed in this design. The specific ordering (order list) will be decided and configured at implementation time (not enumerated in this document).
* Time zone (TZ) follows the OS time zone.

---

## DD-TERMS-001 Terms

* Project Root: Root directory selected by the user, containing category directories and application-managed files.
* Category: A subdirectory directly under Project Root. Categories are handled as a flat list (no nesting).
* Issue JSON: One JSON file per issue, stored under a category directory.
* Contractor / Vendor:

  * Contractor: Ordering company side
  * Vendor: Contractor (subcontractor / partner company) side
* Mode:

  * Contractor mode
  * Vendor mode

---

## DD-ARCH-001 Architecture overview

* Desktop application built with Wails

  * Backend: Go
  * Frontend: Vue 3 + Vuetify 3
* Data storage is file-based (JSON + binary attachments) under the Project Root
* No server-side web application is assumed
* Runs in an offline/on-premises environment without external cloud services or external APIs

---

## DD-DIR-001 Directory structure

Under the Project Root:

* `<PROJECT_ROOT>/`

  * `<category>/`

    * `<issue_id>.json`
    * `<issue_id>.files/`

      * `<attachment_id>_<sanitized_original_name>`
  * `.tmp_rename/`

    * (temporary directory used for category rename; see DD-PERSIST and DD-BE category rename)
  * (application files are not stored here except issue data)

Next to the executable (application distribution folder):

* `config.json`
* `logs/ratta.log`
* `auth/contractor.json` (Contractor-only auth configuration file)

---

## DD-BE-001 Backend design (Go + Wails binding)

### DD-BE-002 Backend responsibilities

* Provide file-based operations (read/write/rename) as stable APIs for the UI
* Perform validation consistently (paths, names, JSON schema, mode constraints)
* Ensure atomic writes for JSON updates
* Return structured errors to the frontend (ApiErrorDTO)
* Detect and report load-time errors (corrupted JSON, leftover tmp files, inconsistencies)

### DD-BE-003 API list

Project bootstrap and Project Root:

* `GetAppBootstrap(): BootstrapDTO`

  * Overview:

    * Load config.json (if it exists)
    * Return ui_page_size, last_project_root_path, and whether auth/contractor.json exists
  * Primary caller:

    * ProjectSelectDialog (initial display)
  * On failure:

    * If reading config.json fails, continue with default values and write a warning log (non-fatal)

* `ValidateProjectRoot(path: string): ValidationResultDTO`

  * Overview:

    * Validate whether the specified path can be used as a Project Root (exists, permissions, is a directory, etc.)
  * Primary caller:

    * ProjectSelectDialog (Open, or pre-check before Create)
  * On failure:

    * If validation cannot be performed (I/O exception etc.), return as error
    * A validation NG is returned as DTO `is_valid=false`

* `CreateProjectRoot(path: string): ValidationResultDTO`

  * Overview:

    * Create the Project Root directory at the specified path (only when it does not exist)
  * Primary caller:

    * ProjectSelectDialog (Create new)
  * Side effects:

    * Only creates the directory. Does not create categories.
  * On failure:

    * If creation is not possible (permissions, existing file, etc.), return as error

* `SaveLastProjectRoot(path: string): void`

  * Overview:

    * Update `last_project_root_path` in config.json
  * Primary caller:

    * ProjectSelectDialog (after Project Root is confirmed)
  * Side effects:

    * Atomically updates config.json

Mode detection:

* `DetectMode(): ModeDTO`

  * Overview:

    * Check whether `auth/contractor.json` exists and return whether a password is required
    * Password verification itself is performed by `VerifyContractorPassword`
  * Primary caller:

    * Immediately after app startup (after Project Root is confirmed)

* `VerifyContractorPassword(password: string): ModeDTO`

  * Overview:

    * Verify whether the input password is correct using encrypted data in `auth/contractor.json`
  * Primary caller:

    * ContractorPasswordDialog
  * On failure:

    * Verification failure returns `E_PERMISSION` (message: “Password verification failed”)
    * ContractorPasswordDialog shows the failure message and exits the application on OK

Category / Issue:

* `ListCategories(): CategoryListDTO`

  * Overview:

    * Return the list of categories (subdirectories) directly under Project Root (flat only; exclusion rules in DD-LOAD-002)
    * Categories remaining under `.tmp_rename` are included as read-only categories (`CategoryDTO.is_read_only=true`)
  * Primary caller:

    * MainView (initial display and refresh for the left pane)

* `CreateCategory(name: string): CategoryDTO`

  * Overview:

    * Create a category directory (Contractor only)
  * Forbidden rules:

    * Duplicate category name and case-only differences are not allowed (treated as an error)
  * On failure:

    * Vendor mode returns `E_PERMISSION`
    * Prohibited characters or length exceeded returns `E_VALIDATION`

* `RenameCategory(oldName: string, newName: string): CategoryDTO`

  * Overview:

    * Rename a category directory (Contractor only)
    * When renaming, update `issue.category` field in existing Issue JSON files
  * On failure:

    * Vendor mode returns `E_PERMISSION`
    * Invalid `newName` returns `E_VALIDATION`
    * `oldName` not found returns `E_NOT_FOUND`
    * If `.tmp_rename` remains and recovery is needed, return `E_CONFLICT`
  * Process (recommended order):

    1. Pre-check:

       * Validate `newName` (prohibited characters, trailing dot/space, length, duplication including case-only)
       * Confirm `oldName` exists
       * If a directory remains under `<PROJECT_ROOT>/.tmp_rename`, treat as incomplete recovery and return `E_CONFLICT`
    2. Rename folder:

       * Rename `<PROJECT_ROOT>/<old>` to `<PROJECT_ROOT>/.tmp_rename/<new>` (same volume assumption)
    3. Update Issue JSON:

       * Scan `*.json` under the moved folder and update each `issue.category` to `newName` with atomic updates
    4. Finalize:

       * If no issues, rename `.tmp_rename/<new>` to `<PROJECT_ROOT>/<new>`
  * Failure handling (explicit rollback boundaries):

    * If failure occurs during steps 1 to 2 (folder operation):

      * Stop processing
      * Try to restore folder name (if possible, revert to `<PROJECT_ROOT>/<oldName>`)
      * If restoration fails, leave `.tmp_rename` and treat it as a startup-detected item
    * If failure occurs after starting step 3 (updating `issue.category`):

      * Stop processing
      * Do not roll back `issue.category` (manual recovery operation)

* `DeleteCategory(name: string): void`

  * Overview:

    * Delete a category directory (Contractor only)
  * On failure:

    * Vendor mode returns `E_PERMISSION`
    * Read-only category returns `E_CONFLICT`

Issue operations (overview level; the concrete list follows the same style as above):

* `ListIssues(category: string, query: IssueListQueryDTO): IssueListDTO`
* `GetIssue(category: string, issueId: string): IssueDetailDTO`
* `UpdateIssue(category: string, issueId: string, payload: UpdateIssueDTO): IssueDetailDTO`
* `AddComment(category: string, issueId: string, payload: AddCommentDTO): IssueDetailDTO`

Load error list:

* `GetLoadErrors(scope: ErrorScopeDTO): ErrorListDTO`

  * Overview:

    * Return a list of errors detected at startup or during load, such as corrupted files, inconsistencies, or leftover tmp artifacts
  * Primary caller:

    * ErrorDetailDialog
  * Notes:

    * If `scope` is category, return only errors for the specified category

---

## DD-BEDTO-001 DTO field definitions

The following defines DTOs used via Wails binding. Type notation is described in a TypeScript-compatible style.

Common tokens:

* ModeToken

  * `"Contractor" | "Vendor"`

* CompanyToken

  * `"Contractor" | "Vendor"`

* StatusToken

  * `"Open" | "Working" | "Inquiry" | "Hold" | "Feedback" | "Resolved" | "Closed" | "Rejected"`

* PriorityToken

  * `"High" | "Medium" | "Low"`

BootstrapDTO:

* `has_config: boolean`

  * Whether config.json exists and could be read (continue even if read fails, as false)
* `last_project_root_path: string | null`
* `ui_page_size: number`

  * Default 20 (read from config.json if stored)
* `log_level: "info" | "debug"`
* `has_contractor_auth_file: boolean`

  * Whether auth/contractor.json exists (if it exists, ContractorPasswordDialog is required)

ValidationResultDTO:

* `is_valid: boolean`
* `normalized_path: string`

  * Normalized path (if it exists)
* `message: string`

  * Short user-facing message
* `details: string | null`

  * Details of failure reason (optional)

ModeDTO:

* `mode: ModeToken`

  * Always returns `"Vendor"` until ContractorPasswordDialog verification is completed
  * Returns `"Contractor"` after successful verification (returns error if verification fails)
* `requires_password: boolean`

  * If true, VerifyContractorPassword must be called

CategoryDTO:

* `name: string`
* `is_read_only: boolean`

  * If true, the category is read-only (edit APIs return `E_CONFLICT`)
* `path: string`

  * Normal: absolute path `<PROJECT_ROOT>/<category>`
  * Read-only category: absolute path `<PROJECT_ROOT>/.tmp_rename/<category>`
* `issue_count: number`

CategoryListDTO:

* `categories: CategoryDTO[]`
* `errors: number`

  * Number of errors during category scan (optional)

IssueListQueryDTO:

* `page: number`

  * 1-based
* `page_size: number`

  * Default 20
* `sort_by: "updated_at" | "due_date" | "priority" | "status" | "title"`
* `sort_order: "asc" | "desc"`
* (filter fields follow the stores/issues query model; see DD-STORE)

ApiErrorDTO (common error type to UI):

* `error_code: string`
* `message: string`
* `detail?: string`
* `target_path?: string`
* `hint?: string`

---

## DD-DATA-001 Data design

### DD-DATA-002 Date/time rules

* Datetime fields:

  * ISO 8601 with time zone, second precision
* Date-only fields:

  * `YYYY-MM-DD`
* JSON key ordering:

  * Not fixed in this design
  * Concrete ordering will be defined at implementation time
* Time zone:

  * Use OS time zone

### DD-DATA-003 Issue JSON (one issue)

* `version: int` (required, starts at 1)
* `issue_id: string` (required, nanoid 9 chars)
* `category: string` (required, matches directory name)
* `title: string` (required, max 255 chars)
* `description: string` (required, max 255 chars)
* `status: string` (required, internal representation is English token)
* `priority: string` (required, `High|Medium|Low`)
* `origin_company: string` (required, `Contractor|Vendor`)
* `assignee: string` (optional)
* `created_at: string` (required, ISO 8601 with TZ, second precision)
* `updated_at: string` (required, ISO 8601 with TZ, second precision)
* `due_date: string` (required, `YYYY-MM-DD`)
* `comments: Comment[]` (required, can be empty)

### DD-DATA-004 Comment

* `comment_id: string` (required, UUID v7)
* `body: string` (required, Markdown, UTF-8 bytes <= 100KB)
* `author_name: string` (required, max 255 chars)
* `author_company: string` (required, `Contractor|Vendor`)
* `created_at: string` (required, ISO 8601 with TZ, second precision)
* `attachments: AttachmentRef[]` (required, can be empty)

Comments are append-only (no edit and no delete).

### DD-DATA-005 AttachmentRef and stored file naming

AttachmentRef (in JSON):

* `attachment_id: string` (required, nanoid 9 chars)
* `file_name: string` (required, original file name, max 255 chars)
* `stored_name: string` (required, stored file name)
* `relative_path: string` (required, `<issue_id>.files/<stored_name>`)
* `mime_type: string` (optional)
* `size_bytes: int` (optional)

Storage location (binary file):

* `<PROJECT_ROOT>/<category>/<issue_id>.files/<attachment_id>_<sanitized_original_name>`

Sanitization rules (Windows prohibited characters):

* Replace `\ / : * ? " < > |` with `_`
* Replace trailing dot `.` with `_`
* Replace trailing space with `_`

---

## DD-PERSIST-001 Persistence and atomic update

### DD-PERSIST-002 Atomic write for JSON

* Use write-to-temp and rename strategy (atomic update)
* Temp file naming:

  * `*.json.tmp.<pid>.<timestamp>`

### DD-PERSIST-003 fsync

* Not performed

### DD-PERSIST-004 Handling leftover tmp artifacts

At startup, detect `*.tmp.*`.

Compute elapsed time from last modified time (mtime) and apply:

* Less than 24 hours:

  * Delete
  * If deletion fails, record as `E_IO_WRITE` and add to the error list
* 24 hours or more:

  * Do not delete
  * Add to the error list (include `target_path`, `message`, and `hint`)

---

## DD-UI-001 Screen design (Vue + Vuetify)

### DD-UI-002 Screen list

* ProjectSelectDialog
* MainView (Category list + Issue list)
* IssueDetailDialog
* ErrorDetailDialog
* ContractorPasswordDialog

### DD-UI-003 State management (recommended)

* `stores/app` (mode, project path, bootstrap state)
* `stores/categories` (category list, selected category)
* `stores/issues` (list cache, filter, sort, paging)
* `stores/errors` (load error list)

### DD-UI-004 ProjectSelectDialog

Elements:

* Display previous `<PROJECT_ROOT>` as initial value
* Browse button (directory select)
* Create new button (create folder at specified path)
* Open button (Validate → Save config → go to next screen)
* Cancel (exit)

Backend calls:

* `GetAppBootstrap`
* `ValidateProjectRoot`
* `CreateProjectRoot`
* `SaveLastProjectRoot`

### DD-UI-005 MainView (Category and Issue list)

Category (left):

* Flat list display
* Contractor only: add, rename, delete
* Read-only category (`is_read_only=true`) must be identifiable in the list, and edit operations for that category are disabled

Issue list (right):

* Columns (per basic design):

  * Issue ID, Title, Status, Priority, Origin Company, Updated At, Due Date
  * If `is_schema_invalid: true`, show a warning
* Sort:

  * Toggle by clicking column headers
* Filter:

  * Status, Priority, Due date
* Paging:

  * Fixed 20 items (can become configurable later)
* Row click:

  * Open IssueDetailDialog
* End-state display:

  * `Closed|Rejected` are grayed out

### DD-UI-006 IssueDetailDialog

Display:

* Title, Description, Due date, Priority, Status, Origin company, Created/Updated at
* Comment list (oldest first)
* Markdown rendering: markdown-it

Edit:

* Initial state is view mode
* Switch to edit mode with an Edit button
* Validate required fields on frontend (Title, Description, Due date, Priority)
* On Save, call backend `UpdateIssue`
* For schema-inconsistent issues (`is_schema_invalid: true`):

  * Disallow update operations and guide user to ErrorDetailDialog
* If selected category is read-only (`is_read_only=true`):

  * Disallow update and add-comment operations
  * Show that it is read-only

Status:

* Limit candidates by mode
* Backend also rejects invalid transitions

Add comment (including attachment):

* Body (Markdown), author name, attachment selection (optional)
* Send: `AddComment` → refresh display
* If body size exceeds 100KB (UTF-8 bytes), sending is not allowed
* No feature to add attachments after posting a comment

  * If needed, post a new comment

### DD-UI-007 ErrorDetailDialog

Display items:

* Target path
* Error type (code)
* Message
* Detail (collapsible)
* Copy button (path, message)

Entry points:

* Open from menu or header error icon
* Category-level filtering

### DD-UI-008 ContractorPasswordDialog

* Password input
* On failure, show “verification failed”, and exit the application when the user closes it

---

## DD-STORE-001 Pinia store design

### DD-STORE-002 Purpose and responsibilities

* Share UI state across screens (selected category, list cache, sort/filter/page, currently opened issue detail, etc.)
* Encapsulate backend calls (Wails binding) within actions; screens call only store APIs
* Cache list views; always reload issue detail from disk
* Show schema-inconsistent issues in the list with warnings, but reject update operations and guide to error details
* Aggregate errors into `stores/errors`; screens display errors using `stores/errors` as the single reference point

### DD-STORE-003 Store list

* `stores/app`

  * Bootstrap, Project Root, mode (Contractor/Vendor), common settings
* `stores/categories`

  * Category list, selected category
* `stores/issues`

  * Issue list cache (per category), sort, filter, paging
* `stores/issueDetail`

  * Open issue detail, edit state, save, add comment
* `stores/errors`

  * Error aggregation and display-oriented data

### DD-STORE-004 Type definitions (common)

* Use Backend DTOs as the base, and keep UI-only state (e.g., isDirty) as separate fields

TypeScript types (examples):

```ts
export type Mode = "Contractor" | "Vendor";

export type SortKey = "updated_at" | "due_date" | "priority" | "status" | "title";
export type SortDir = "asc" | "desc";

export type IssueStatus =
  | "Open"
  | "Working"
  | "Inquiry"
  | "Hold"
  | "Feedback"
  | "Resolved"
  | "Closed"
  | "Rejected";
export type IssuePriority = "High" | "Medium" | "Low";
export type Company = "Contractor" | "Vendor";

export interface ApiErrorDTO {
  error_code: string;
  message: string;
  detail?: string;
  target_path?: string;
  hint?: string;
}
```

### DD-STORE-005 stores/app state

```ts
export interface AppState {
  mode: Mode;
  projectRoot: string | null;
  pageSize: number; // 20
  bootstrapLoaded: boolean;

  contractorAuthRequired: boolean;
  isBusy: boolean;
}
```

### DD-STORE-006 stores/categories state

```ts
export interface CategoryDTO {
  name: string;
  issueCount?: number;
}

export interface CategoriesState {
  items: CategoryDTO[];
  selectedCategory: string | null;

  isLoading: boolean;
  lastLoadedAt: string | null;
}
```

### DD-STORE-007 stores/issues state

```ts
export interface IssuesQueryState {
  sort: { key: SortKey; dir: SortDir };
  filter: {
    text: string;
    status: IssueStatus[];         // if empty, all
    priority: IssuePriority[];     // if empty, all
    dueDateFrom: string | null;    // YYYY-MM-DD
    dueDateTo: string | null;      // YYYY-MM-DD
    schemaInvalidOnly: boolean;
  };
  page: number; // 1-based
}

export interface IssueSummaryDTO {
  issue_id: string;
  category: string;
  title: string;
  status: IssueStatus;
  priority: IssuePriority;
  origin_company: Company;
  updated_at: string;
  due_date: string;
  is_schema_invalid: boolean;
}

export interface IssuesCacheEntry {
  items: IssueSummaryDTO[];
  total: number;
  lastLoadedAt: string | null;
  isLoading: boolean;
}

export interface IssuesState {
  issuesByCategory: Record<string, IssuesCacheEntry>;
  queryByCategory: Record<string, IssuesQueryState>;
  defaultQuery: IssuesQueryState;
}
```

### DD-STORE-008 stores/issueDetail state

```ts
export interface AttachmentDTO {
  attachment_id: string;
  file_name: string;
  stored_name: string;
  relative_path: string;
  mime_type?: string;
  size_bytes?: number;
}

export interface CommentDTO {
  comment_id: string;
  body: string;
  author_name: string;
  author_company: Company;
  created_at: string;
  attachments: AttachmentDTO[];
}

export interface IssueDetailDTO {
  version: number;
  issue_id: string;
  category: string;
  title: string;
  description: string;
  status: IssueStatus;
  priority: IssuePriority;
  origin_company: Company;
  assignee?: string;
  created_at: string;
  updated_at: string;
  due_date: string;
  comments: CommentDTO[];
  is_schema_invalid: boolean;
}

export interface IssueDetailState {
  current: IssueDetailDTO | null;
  currentCategory: string | null;
  isLoading: boolean;

  isDirty: boolean;
  lastLoadedAt: string | null;
}
```

### DD-STORE-009 stores/errors state

```ts
export type ErrorSource =
  | "app"
  | "categories"
  | "issues"
  | "issueDetail"
  | "comments"
  | "attachments"
  | "project"
  | "backend";

export interface UiErrorItem {
  id: string;
  occurred_at: string;
  source: ErrorSource;
  action: string;
  category?: string;
  issue_id?: string;
  severity: "info" | "warn" | "error";
  api?: ApiErrorDTO;
  raw?: unknown;
  user_message: string;
  is_read: boolean;
}

export interface ErrorsState {
  items: UiErrorItem[];
}
```

### DD-STORE-010 actions definition (common policy)

* All backend calls are encapsulated inside actions
* On failure, always register into `stores/errors` (DD-STORE-016)

### DD-STORE-011 stores/errors actions

* `capture(e, ctx)`

  * Overview: Normalize exceptions/errors into UiErrorItem and append to items
* `captureApiError(apiError, ctx)`

  * Overview: Receive ApiErrorDTO from backend and register it
* `captureMany(list, ctx)`

  * Overview: Register multiple errors detected during load (corruption list etc.)
* `markRead(id)`, `markAllRead()`

  * Overview: Read/unread management
* `clearAll()`, `clearBySource(source)`

  * Overview: Clear error list
* `loadFromBackend(scope)`

  * Overview: Call backend `GetLoadErrors(scope)` and import results via `captureMany` (use as needed)

### DD-STORE-012 stores/app actions

* `bootstrap()`

  * Overview: Call GetAppBootstrap and reflect pageSize, last project root, whether auth/contractor.json exists, etc.
* `selectProjectRoot(path)`

  * Overview: ValidateProjectRoot → SaveLastProjectRoot, then update projectRoot
* `createProjectRoot(path)`

  * Overview: CreateProjectRoot → SaveLastProjectRoot, then update projectRoot
* `detectMode()`

  * Overview: Call DetectMode and determine contractorAuthRequired and mode
* `verifyContractorPassword(password)`

  * Overview: Call VerifyContractorPassword to finalize Contractor mode
  * On failure: register to errors and provide UI flow to exit

### DD-STORE-013 stores/categories actions

* `loadCategories()`

  * Overview: Call ListCategories and refresh category list
* `selectCategory(name)`

  * Overview: Update selectedCategory and call `stores/issues.loadIssues(name)`
* `createCategory(name)`

  * Overview: Contractor only. Call CreateCategory and refresh list
* `renameCategory(oldName, newName)`

  * Overview: Contractor only. Call RenameCategory and update selectedCategory and issuesByCategory key consistency
* `deleteCategory(name)`

  * Overview: Contractor only. Call DeleteCategory and refresh list and related cache

Permission control:

* In Vendor mode, do not execute and register an `E_PERMISSION` equivalent error into errors

### DD-STORE-014 stores/issues actions

* `loadIssues(category, opts)`

  * Overview: Call ListIssues and update cache `issuesByCategory[category]`
  * Notes: Backend receives opts and applies sort/filter/page formatting
* `refreshIssues(category)`

  * Overview: Force loadIssues
* `setSort(category, sort)`

  * Overview: Update queryByCategory and apply to UI ordering
* `setFilter(category, filter)`

  * Overview: Update queryByCategory and apply to UI filtering
* `setPage(category, page)`

  * Overview: Update queryByCategory and apply to paging
* `invalidateCategory(category)`

  * Overview: Discard cache and require loadIssues next time
* `applyIssueUpdatedToCache(issueDetail)`

  * Overview: After saveIssue/addComment succeeds, update the corresponding summary in the list cache

### DD-STORE-015 stores/issueDetail actions

* `openIssue(category, issueId)`

  * Overview: Call GetIssue and set latest IssueDetailDTO into current (always reload from disk)
* `reloadCurrent()`

  * Overview: Re-run GetIssue if current exists
* `saveIssue(update)`

  * Overview: Call UpdateIssue and update current
  * Constraint: If `current.is_schema_invalid` is true, reject save and register to errors
* `addComment(payload)`

  * Overview: Call AddComment and update current
  * Constraint: Attachment is optional, but only allowed together with posting a comment

### DD-STORE-016 Error aggregation rule (aggregate into stores/errors)

* All store actions must wrap backend calls with try/catch, and on failure always register into `stores/errors`
* Error display to screens is unified as follows:

  * Even if transient notifications (toast etc.) are shown, registration is performed first
  * ErrorDetailDialog displays from `stores/errors.items`

---
