# tasks.md

## Task execution rules

- Implement tasks top-to-bottom
- Each task must include tests (TDD) for the behavior it introduces
- Prefer unit tests first; add Playwright only for representative end-to-end flows
- Refer to docs/basic_design.md and docs/detailed_design.md for IDs and exact behavior

## Phase 0: Repository foundation

- [x] TASK-0001 Initialize repository structure
  - Deliverables
    - Wails project initialized
    - Go module initialized
    - Vue 3 + Vuetify 3 frontend wired into Wails
    - Pinia installed and wired
    - Baseline build and dev commands documented in README
  - Acceptance
    - wails dev starts a window
    - go test ./... passes (even if empty)
    - npm test passes (baseline)

- [x] TASK-0002 Establish lint/format conventions
  - Deliverables
    - Go formatting via gofmt
    - Frontend formatting via project standard (eslint/prettier if used)
    - Ensure JSON output formatting strategy is implemented in code (indent 2, LF) but may be a placeholder until TASK-0205
  - Acceptance
    - CI-like local scripts exist (make/dev scripts are acceptable)

## Phase 1: Schemas and shared primitives

- [x] TASK-0101 Add JSON Schema files
  - Deliverables
    - schemas/issue.schema.json
    - schemas/config.schema.json
    - schemas/contractor.schema.json
    - Each schema includes $schema and does not allow external $ref
  - Tests
    - Go unit tests: schema loader rejects HTTP refs and loads local schemas
  - Acceptance
    - Schema compilation succeeds at runtime

- [x] TASK-0102 Implement ID utilities
  - Deliverables
    - issue_id generator (nanoid 9)
    - attachment_id generator (nanoid 9)
    - comment_id generator (UUID v7)
  - Tests
    - Unit tests for length/format expectations and uniqueness sampling
  - Acceptance
    - Deterministic tests (no flaky timing assumptions)

- [x] TASK-0103 Implement time utilities
  - Deliverables
    - ISO 8601 with TZ generation (second precision)
    - UI formatting helper (Japanese display format) on frontend
  - Tests
    - Go unit tests for serialization format
    - Frontend unit tests for display formatting
  - Acceptance
    - No millisecond persistence

## Phase 2: Backend infrastructure (file I/O, validation, logging, config)

- [x] TASK-0201 Implement atomic write helper
  - Scope
    - Write tmp file in same directory, then rename to target
    - tmp naming: <issue_id>.json.tmp.<pid>.<timestamp> (or equivalent)
    - No fsync
  - Tests
    - Unit tests covering: write success, rename failure, partial tmp cleanup behavior
  - Acceptance
    - Existing file is not corrupted on simulated failure

- [x] TASK-0202 Implement tmp residue scan policy
  - Scope
    - Detect *.tmp.* at startup
    - If mtime < 24h: delete; if >= 24h: keep and record as load error
  - Tests
    - Unit tests with temp directories and mocked mtimes
  - Acceptance
    - Errors are recorded with target_path and hints

- [x] TASK-0203 Implement attachment storage helper
  - Scope
    - Create <issue_id>.files if needed
    - Sanitize filenames for Windows invalid characters and trailing dot/space
    - Ensure 255 char limit and collision suffixing
    - Save via staging name then rename
  - Tests
    - Unit tests for sanitization, collision handling, rollback on failure
  - Acceptance
    - If JSON update fails after attachments staged, attachments are deleted (rollback)

- [x] TASK-0204 Implement schema validation wrapper
  - Scope
    - Load/compile schemas from schemas/ directory
    - Validate issue/config/contractor JSON
    - Produce structured errors suitable for ApiErrorDTO.detail
  - Tests
    - Unit tests: valid JSON passes, invalid JSON produces actionable error summary
  - Acceptance
    - External refs are not resolved

- [x] TASK-0205 Implement canonical JSON writer
  - Scope
    - Indent 2 spaces, LF
    - Fixed key order (define order list in code)
  - Tests
    - Golden-file tests verifying stable output for the same input
  - Acceptance
    - Running save twice produces identical JSON bytes

- [x] TASK-0206 Implement config.json repository
  - Scope
    - Location: alongside ratta.exe
    - Fields per docs/detailed_design.md (format_version, last_project_root_path, log.level, ui.page_size)
    - Update via atomic write
  - Tests
    - Unit tests: missing config uses defaults; corrupt config logs warning and continues
  - Acceptance
    - SaveLastProjectRoot updates last_project_root_path correctly

- [x] TASK-0207 Implement structured logging with rotation
  - Scope
    - logs/ratta.log
    - 1MB max, 3 generations
    - Level control from config (info/debug)
    - Avoid logging sensitive user inputs
  - Tests
    - Unit tests: rotation behavior in a temp dir
  - Acceptance
    - Log file is created and rotated as specified

## Phase 3: Contractor auth and CLI

- [x] TASK-0301 Implement crypto for contractor.json
  - Scope
    - PBKDF2-HMAC-SHA256 key derivation (iterations/salt length per docs)
    - AES-256-GCM encrypt fixed plaintext "contractor-mode"
    - Store salt/nonce/ciphertext in base64 with format_version
  - Tests
    - Unit tests: round-trip success, wrong password fails, corrupted fields fail
  - Acceptance
    - Failure maps to E_CRYPTO with useful hint

- [x] TASK-0302 Implement CLI: ratta.exe init contractor
  - Scope
    - Console password input (hidden), confirmation required
    - --force to overwrite existing auth/contractor.json
    - Creates auth/ directory if needed
  - Tests
    - Unit tests for file creation and overwrite rules (abstract console input)
  - Acceptance
    - Existing file without --force fails with non-zero exit

## Phase 4: Backend domain and use-cases

- [x] TASK-0401 Implement domain types and validation
  - Scope
    - Status tokens and end-state rules
    - Priority tokens
    - Category name validation (Windows invalid chars, trailing dot/space, length)
    - Issue field validation (required fields, 255 char constraints, due_date format)
    - Comment validation (author_name required, body <= 100KB UTF-8 bytes, attachments <= 5)
  - Tests
    - Unit tests per rule
  - Acceptance
    - Validation errors map to E_VALIDATION

- [x] TASK-0402 Implement mode detection and permission checks
  - Scope
    - DetectMode: checks existence of auth/contractor.json
    - VerifyContractorPassword: verifies and returns Contractor mode
    - Status transition permission rules (Vendor vs Contractor)
  - Tests
    - Unit tests: permission matrix for transitions and edit restrictions
  - Acceptance
    - Backend rejects disallowed transitions even if UI tries

## Phase 5: Backend data loading, error registry, and Wails API

- [x] TASK-0501 Implement project root validation and creation
  - Scope
    - ValidateProjectRoot, CreateProjectRoot, SaveLastProjectRoot
  - Tests
    - Unit tests for path validation cases
  - Acceptance
    - Normalized path returned when possible

- [x] TASK-0502 Implement category scanning and read-only category detection
  - Scope
    - Flat scan of <PROJECT_ROOT> directories
    - Exclude .git and dot dirs except .tmp_rename
    - Include .tmp_rename/<category> as read-only CategoryDTO
  - Tests
    - Unit tests with sample directory trees
  - Acceptance
    - Nested subfolders not treated as categories

- [x] TASK-0503 Implement issue scanning and classification
  - Scope
    - Load *.json under a category
    - Classify: JSON parse failure vs schema invalid vs valid
    - Parse failure excluded from list but registered as load error
    - Schema invalid included with is_schema_invalid=true and read-only update policy
    - Unsupported issue.version treated as schema invalid (read-only)
  - Tests
    - Unit tests for each classification type
  - Acceptance
    - GetLoadErrors returns the recorded items

- [x] TASK-0504 Implement issue persistence and core operations
  - Scope
    - GetIssue (always reload from disk)
    - CreateIssue (origin_company from current mode; comments starts empty)
    - UpdateIssue (reject end-state; reject schema-invalid; update updated_at)
    - ListIssues (supports sort/filter/page per query DTO)
  - Tests
    - Unit tests: create/update rules, paging and sorting determinism
  - Acceptance
    - ListIssues result is stable and matches query inputs

- [x] TASK-0505 Implement comment add with attachments
  - Scope
    - AddComment appends to comments array
    - Saves attachments first with staging, then updates JSON
    - If JSON update fails, delete attachments (rollback)
  - Tests
    - Unit tests for rollback and success paths
  - Acceptance
    - Comments displayed oldest-to-newest based on array order

- [x] TASK-0506 Implement category mutators (Contractor only)
  - Scope
    - CreateCategory (reject duplicates including case-insensitive)
    - DeleteCategory (only if no *.json; treat only .files as empty)
    - RenameCategory with .tmp_rename workflow and error boundaries
  - Tests
    - Unit tests: rename steps, conflict when .tmp_rename exists, permission errors
  - Acceptance
    - Read-only categories block edits and deletes with E_CONFLICT

- [x] TASK-0507 Implement Wails binding layer and DTO mapping
  - Scope
    - Expose APIs listed in docs/detailed_design.md
    - Standard ResponseDTO (ok/data/error) mapping for frontend
  - Tests
    - Unit tests for error mapping to ApiErrorDTO
  - Acceptance
    - Frontend can call all APIs without importing internal Go structs

## Phase 6: Frontend stores and UI

- [x] TASK-0601 Implement Wails API client wrappers
  - Scope
    - Typed wrappers around generated Wails bindings
    - Normalize ResponseDTO and throw or return consistent results to stores
  - Tests
    - Unit tests with mocked responses
  - Acceptance
    - Errors flow into stores/errors consistently

- [x] TASK-0602 Implement Pinia stores skeleton
  - Scope
    - stores/app, stores/categories, stores/issues, stores/issueDetail, stores/errors
    - Implement actions per docs/detailed_design.md, including selectCategory calling loadIssues
  - Tests
    - Store unit tests for state transitions and error capture
  - Acceptance
    - All backend call failures are captured into stores/errors

- [x] TASK-0603 Implement ProjectSelectDialog
  - Scope
    - Uses bootstrap to prefill last project root
    - Validate/Open, Create new, Cancel exits
  - Tests
    - Component tests: validation error display, happy path
  - Acceptance
    - Selecting a project root persists last_project_root_path

- [x] TASK-0604 Implement ContractorPasswordDialog
  - Scope
    - Shown when auth/contractor.json exists
    - On failure shows message then exits flow
  - Tests
    - Component tests: failure message and close behavior routing
  - Acceptance
    - Mode becomes Contractor only after successful verification

- [x] TASK-0605 Implement MainView (categories + issue list)
  - Scope
    - Left: categories list, Contractor-only controls (create/rename/delete)
    - Right: issue list with columns, paging (20), sorting, filtering
    - End-state issues greyed out
    - Schema-invalid issues show warning and block update actions
  - Tests
    - Component/store integration tests for sorting/filter/paging
  - Acceptance
    - Selecting a category loads and displays issues

- [x] TASK-0606 Implement IssueDetailDialog
  - Scope
    - View mode by default; edit mode after explicit action
    - UpdateIssue flow with required field validation
    - Comment add with optional attachments (up to 5)
    - Markdown rendering for comment body (markdown-it)
    - Read-only category and schema-invalid blocks editing/commenting, routes to error detail
  - Tests
    - Component tests for edit toggling, validation, blocked behavior
  - Acceptance
    - Detail reloads from disk on open/reload

- [ ] TASK-0607 Implement ErrorDetailDialog
  - Scope
    - Displays aggregated errors from stores/errors
    - Supports scope filtering (all/category)
    - Copy buttons for paths/messages
  - Tests
    - Component tests for rendering and filter
  - Acceptance
    - Backend load errors are visible and actionable

## Phase 7: End-to-end quality gates and packaging

- [ ] TASK-0701 Add representative Playwright E2E scenarios
  - Scenarios
    - Create issue -> JSON created -> appears in list
    - Add comment (no attachments) -> JSON updated -> appears in detail
    - Add comment (with attachments) -> files created -> appears in detail
    - Vendor cannot set Closed/Rejected
    - Schema-invalid JSON shows warning, update disabled, error visible
    - Parse-broken JSON excluded from list, error visible
    - Detail reload picks up external file change
  - Acceptance
    - E2E suite runs locally and is stable

- [ ] TASK-0702 Build and distribution checks
  - Scope
    - wails build outputs a zip-distributable directory layout
    - logs/, auth/, config.json behaviors validated
  - Acceptance
    - Fresh unzip run works without installer steps

- [ ] TASK-0703 Documentation alignment check
  - Scope
    - Confirm implementation matches docs/basic_design.md and docs/detailed_design.md
    - Record any intentional deviations in a short DEV_NOTES.md (only if required)
  - Acceptance
    - No untracked behavioral divergence
