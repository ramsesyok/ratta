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

- [ ] TASK-0202 Implement tmp residue scan policy
  - Scope
    - Detect *.tmp.* at startup
    - If mtime < 24h: delete; if >= 24h: keep and record as load error
  - Tests
    - Unit tests with temp directories and mocked mtimes
  - Acceptance
    - Errors are recorded with target_path and hints

- [ ] TASK-0203 Implement attachment storage helper
  - Scope
    - Create <issue_id>.files if needed
    - Sanitize filenames for Windows invalid characters and trailing dot/space
    - Ensure 255 char limit and collision suffixing
    - Save via staging name then rename
  - Tests
    - Unit tests for sanitization, collision handling, rollback on failure
  - Acceptance
    - If JSON update fails after attachments staged, attachments are deleted (rollback)

- [ ] TASK-0204 Implement schema validation wrapper
  - Scope
    - Load/compile schemas from schemas/ directory
    - Validate issue/config/contractor JSON
    - Produce structured errors suitable for ApiErrorDTO.detail
  - Tests
    - Unit tests: valid JSON passes, invalid JSON produces actionable error summary
  - Acceptance
    - External refs are not resolved

- [ ] TASK-0205 Implement canonical JSON writer
  - Scope
    - Indent 2 spaces, LF
    - Fixed key order (define order list in code)
  - Tests
    - Golden-file tests verifying stable output for the same input
  - Acceptance
    - Running save twice produces identical JSON bytes

- [ ] TASK-0206 Implement config.json repository
  - Scope
    - Location: alongside ratta.exe
    - Fields per docs/detailed_design.md (format_version, last_project_root_path, log.level, ui.page_size)
    - Update via atomic write
  - Tests
    - Unit tests: missing config uses defaults; corrupt config logs warning and continues
  - Acceptance
    - SaveLastProjectRoot updates last_project_root_path correctly

- [ ] TASK-0207 Implement structured logging with rotation
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

- [ ] TASK-0301 Implement crypto for contractor.json
  - Scope
    - PBKDF2-HMAC-SHA256 key derivation (iterations/salt length per docs)
    - AES-256-GCM encrypt fixed plaintext "contractor-mode"
    - Store salt/nonce/ciphertext in base64 with format_version
  - Tests
    - Unit tests: round-trip success, wrong passwor
