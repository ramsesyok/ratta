# Basic Design Document (Issue-Tracking Desktop Application ratta)

## BD-DOC Overview

### BD-DOC-001 Purpose of this document

This document describes the program-level basic design of the issue-tracking desktop application ratta (hereafter, “this system”).

Functional requirements (FR-xxx) and technical requirements (TR-xxx) for this system are defined in the Requirements Specification, and system requirements (SY-xxx, F-xxx, UI-xxx, D-xxx, T-xxx, NF-xxx) are defined in the Requirements Definition. This document concretizes them from the perspectives of program structure, functional composition, data structures, and screen specifications, and serves as a basis for detailed design, implementation, and test planning.

### BD-DOC-002 System name and summary

* System name: ratta
* System summary:

  * A desktop application for managing issues between a Contractor and a Vendor.
  * Runs in an air-gapped, on-premises environment.
  * Uses JSON files for persistence, and relies on external tools such as git for synchronization.

### BD-DOC-003 Scope of this document

This document includes:

* The ratta application itself implemented as a desktop application using Go + Wails
* Screen structure and major components of the front-end (Vue 3 + Vuetify 3)
* Data structures and file structure for issue JSON files, comments, and attachments
* Placement and handling of configuration files, log files, and Contractor-only configuration files
* Basic error handling and fault-tolerance policy
* Development and test policy (TDD and role separation of unit tests / E2E tests)

The following items are out of scope and will be handled in separate documents if needed:

* Configuration and operational procedures for version control systems such as git
* File server configuration and backup operations
* Security design details (such as encryption key management policy)

### BD-DOC-004 Terminology

* Contractor: Contractor mode
* Vendor: Vendor mode
* Project (issue data folder): Folder that stores issue JSON files (<PROJECT_ROOT>)
* Corrupted file: An issue JSON that cannot be interpreted as JSON, or cannot be minimally displayed due to missing required fields
* Schema mismatch: Readable as JSON, but does not match the schema (partial display may be possible)

### BD-DOC-005 Reference documents and reference rules

* Requirements Specification: requirements_specification.md (FR-xxx, TR-xxx)
* Requirements Definition: requirements_definition.md (SY-xxx, F-xxx, UI-xxx, D-xxx, T-xxx, NF-xxx)
* Detailed Design: docs/detailed_design.md
* Machine-readable definitions (appendix):

  * schemas/issue.schema.json
  * schemas/config.schema.json
  * schemas/contractor.schema.json

Reference rules:

* The basic design includes references (the referenced file names, locations, and responsibilities are fixed).
* Fixed values such as schema definitions are authoritative in the referenced documents (schemas/ and docs/).
* If values in this document conflict with referenced documents, the Requirements Definition and the referenced definitions take precedence, and this basic design should be revised.

---

## BD-POLICY Design policy

### BD-POLICY-001 Architecture policy

* BD-ARCH-001 Desktop client approach

  * Implement as a Windows desktop application using Wails.
  * Do not use a browser-based system with a permanent web server; the client runs locally and is self-contained.

* BD-ARCH-002 Separation of front-end and back-end

  * Back-end is implemented in Go and is responsible for issue-management business logic, JSON file I/O, configuration, logging, etc.
  * Front-end is implemented with Vue 3 + Vuetify 3 and is responsible for the UI and input validation.
  * Vue calls Go functions via the Wails binding mechanism.

* BD-ARCH-003 Data persistence policy

  * Store issue data as “one JSON file per issue.”
  * When updating JSON, write to a temporary file and then rename (atomic replacement) to prevent corruption due to abnormal termination mid-write.
  * Consistency across issues and comments is ensured by the JSON schema and application-side input control; conflict resolution caused by multiple terminals updating the same file is delegated to external merge functions of tools such as git.

* BD-ARCH-004 Mode switching (Contractor mode / Vendor mode)

  * At startup, determine the ratta operating mode (Contractor mode or Vendor mode) by checking whether a Contractor-only configuration file exists and by password verification.
  * Depending on the mode, switch availability/constraints of some functions such as status changes and category operations.

* BD-ARCH-005 Authentication and user management

  * Do not implement user management or login within this system.
  * As an authentication-equivalent element, only password verification stored in the Contractor-only configuration file is used.

### BD-POLICY-002 Technologies and languages

* BD-TECH-001 Back-end

  * Language: Go
  * Framework: Wails
  * Main external libraries

    * Issue ID generation: github.com/matoous/go-nanoid

      * Character set: `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
      * Length: 9
    * Comment ID generation: github.com/google/uuid (UUID v7)
    * JSON handling: Use the standard library encoding/json as the base, and use helper libraries if needed.

* BD-TECH-002 Front-end

  * Vue 3 (Composition API, based on the template created by `wails init -n myproject -t vue`)
  * Vuetify 3 (use the default theme)
  * Markdown rendering: markdown-it

* BD-TECH-003 Documentation structure

  * Place the following documents under the docs folder.

    * requirements_specification.md: Requirements Specification (English)
    * requirements_specification_ja.md: Requirements Specification (Japanese)
    * requirements_definition.md: Requirements Definition (English)
    * requirements_definition_ja.md: Requirements Definition (Japanese)

Related requirements: T-001, T-002, T-003

### BD-POLICY-003 Responsibility boundary (Go ⇔ Vue)

* BD-DUTY-001 Responsibilities on the Go side

  * File I/O (read, validate, update, attachment save)
  * Schema validation
  * ID generation
  * Mode decision (Contractor/Vendor)
  * Log output
  * Detection of corrupted / inconsistent files and providing error information

* BD-DUTY-001 Responsibilities on the Vue side

  * Display and input
  * In-screen validation (required fields, character count, forbidden characters, etc.)
  * Screen state management
  * Unified error display (centralize in the error details dialog)

Related requirements: SY-003, T-001

### BD-POLICY-004 Development and test policy

* BD-DEV-001 Adopt Test-Driven Development (TDD)

  * Implement business logic on the Go side using TDD with the Go standard testing package.
  * When adding new functionality, write unit tests first, then implement to make them pass.

* BD-DEV-002 Unit tests

  * Back-end (Go): Unit tests using the testing package
  * Front-end (Vue): Component-level unit tests using Vitest
  * Structure the system so that unit tests can be executed separately for back-end and front-end in principle.

* BD-DEV-003 E2E tests

  * Perform E2E tests including real UI operations and the Go back-end by accessing the dev server started by `wails dev` from Playwright.
  * Example representative scenarios

    * Create a new issue → an issue file is created in JSON and reflected in the issue list screen.
    * Add a comment → a comment is added in JSON and reflected in the issue detail screen.
    * In Vendor mode, status changes are constrained.

* BD-DEV-004 Final tests

  * Create test items based on the requirement IDs (FR-xxx, TR-xxx) defined in the Requirements Specification, and ensure traceability between requirements and tests.

Related requirements: T-001 (adopted technologies), NF-001 (supporting performance assumptions), NF-002 (supporting operational assumptions)

---

## BD-CRITERIA Design conditions

### BD-CRITERIA-001 System prerequisites

* Target OS: Windows 10 64bit or later
* Operating environment: Air-gapped on-premises environment
* This system runs standalone and does not depend on cloud services or external web APIs.

### BD-CRITERIA-002 Operational and data assumptions

* Target project: A single software development project handled between a Contractor and a Vendor.
* User scale: Assume about 20 total users (Contractor/Vendor combined). Assume about 10 concurrent users.
* Issue count / comment count: Assume a scale of several hundred each.
* Synchronization of issue/comment data is performed by an externally built version control system such as git. ratta itself does not provide synchronization features.

### BD-CRITERIA-003 Non-functional requirements (excerpt)

* Performance: Ensure practical response times even when handling several hundred issues/comments (details in Chapter 7).
* Reliability: Provide corruption detection and safe update methods for JSON data to reduce fatal data-loss risk.
* Maintainability: Clearly separate front-end and back-end, and implement with TDD and automated tests as a premise to make the system resilient to changes.

Related requirements: NF-001, D-006, T-006

---

## BD-SYSTEM System configuration

### BD-SYSTEM-001 Hardware configuration

* Client PC

  * CPU: Intel Core i5 equivalent or better
  * Memory: 8GB or more
  * Storage: Several GB or more free space (including issue data, logs, temporary files, etc.)

The storage location for issue data is one of the following:

* A folder under the local disk of the client PC
* A shared folder on an internal file server (SMB, etc.)

Related requirements: NF-001

### BD-SYSTEM-002 Software configuration

* OS: Windows 10 64bit or later
* Runtime / framework

  * Go runtime
  * Wails
  * Node.js / npm (for front-end build)
* Front-end

  * Vue 3, Vuetify 3, markdown-it
* Logs, configuration, Contractor-only information

  * Create the following folders at the same level as ratta.exe and store them:

    * auth folder: Contractor-only configuration file (contractor.json)
    * logs folder: log files (ratta.log, etc.)

### BD-SYSTEM-003 Project folder structure (for development)

The development project is based on the Wails template and uses the following structure:

* docs

  * requirements_definition.md
  * requirements_definition_ja.md
  * requirements_specification.md
  * requirements_specification_ja.md
  * basic_design.md
  * basic_design_ja.md
  * detailed_design.md
  * detailed_design_ja.md

* schemas

  * issue.schema.json
  * config.schema.json
  * contractor.schema.json

* build

  * darwin
  * windows

    * installer

* frontend

  * dist
  * src

    * assets

      * fonts
      * images
    * components
  * wailsjs

    * go

      * main
    * runtime

The issue data folder used at application runtime is specified separately from the above project structure.

### BD-SYSTEM-004 Distribution artifacts and directory structure (runtime)

* Under the executable directory

  * ratta.exe
  * auth/

    * contractor.json (if present, it is a candidate for Contractor mode)
  * logs/

    * ratta.log (with rotation)
    * audit.log (optional, if implemented)
  * config.json (application configuration)

Notes:

* docs/ and schemas/ are reference materials on the application side (source/distribution artifacts), and are not placed under <PROJECT_ROOT> (issue data folder).

Related requirements: F-002, F-005, T-006, SY-003

---

## BD-DEVREQ Requirement analysis for the program

Refer to the following documents for detailed requirements:

* Requirements Specification (FR-xxx, TR-xxx)

  * docs/requirements_specification_ja.md
* Requirements Definition (SY-xxx, F-xxx, UI-xxx, D-xxx, T-xxx, NF-xxx)

  * docs/requirements_definition_ja.md

In this document, ensure traceability with the Requirements Specification / Requirements Definition through the functional structure in Chapter 6 and the data specifications in Chapters 8 and 10.

---

## BD-FUNCTIONS Functions

### BD-FUNCTIONS-001 Function list

The main functions provided by ratta are as follows (items in parentheses are examples of corresponding requirement groups):

* BD-FB-001 Project startup and data folder management

  * At startup, display the project selection dialog and determine the issue data folder.
  * Save the determined issue data folder path to the application configuration file and use it as the default value on the next startup.
  * Treat one startup as one project; do not provide a project-switching function after startup.
  * Related requirements: F-001, UI-000, D-001, SY-001, NF-001, NF-002

* BD-FB-002 Mode decision and permission control

  * Determine Contractor mode / Vendor mode based on existence of the Contractor-only configuration file (auth/contractor.json) and password verification.
  * Apply mode-based constraints on status changes and category operations.
  * Related requirements: SY-002, F-002, F-004, F-301

* BD-FB-003 Initialization command (generate Contractor-only configuration file)

  * Provide `ratta.exe init contractor` and generate auth/contractor.json.
  * The initialization command can be executed without starting the GUI.
  * If the configuration file already exists, do not overwrite unless an explicit overwrite option (example: `--force`) is specified; exit with an error.
  * Related requirements: F-005, FR-024, TR-007

* BD-FB-004 Category and issue list display

  * Display the issue list per category.
  * Provide sort, filter, and paging.
  * Related requirements: F-101, F-102, F-103, UI-001, UI-002, UI-003, UI-004, NF-001, NF-002

* BD-FB-005 Issue creation and editing

  * Create and edit issue information such as title, description, requested response due date, and priority.
  * Validate required fields.
  * Related requirements: F-003, F-101, F-102, F-103, UI-005, D-003, D-004

* BD-FB-006 Status management

  * Define status types and transition rules, apply permission control per Contractor/Vendor mode, and control state transitions.
  * Related requirements: F-004, F-101

* BD-FB-007 Comment management

  * Display comments for each issue in chronological order and allow adding new comments.
  * Limit comment size to 100KB or less in UTF-8 bytes.
  * Do not provide comment edit/delete functions.
  * Related requirements: F-201, F-202

* BD-FB-008 Attachment file management

  * Manage registration, saving, and reference paths of attachment files per comment.
  * Related requirements: F-203, D-001, D-003

* BD-FB-009 Category structure management

  * Treat the directory structure as categories.
  * In Contractor mode only, provide category add/delete/rename.
  * Related requirements: F-301, D-001

* BD-FB-010 JSON file I/O and validation

  * Read issue JSON, validate schema, and detect corruption.
  * When writing, replace safely using a temporary file.
  * Related requirements: D-002, D-003, D-004, D-005, D-006, T-004, T-005

* BD-FB-011 Log output and error display

  * Output internal errors and important events to log files.
  * Perform log rotation.
  * Related requirements: T-006
  * Display error information detected during JSON read/validation in a dedicated dialog as a list.
  * Related requirements: T-006 (operability and maintainability)

### BD-FUNCTIONS-002 Flow between functions (overview)

* BD-FLOW-01 Execute initialization command (when `init contractor` is specified as a startup argument)

  * Parse startup arguments; if it is an initialization command, execute generation of auth/contractor.json without starting the GUI.
  * Execute generation of auth/contractor.json without starting the GUI.
  * If an existing file is present, do not overwrite unless `--force` is specified.
  * Exit after successful generation. On failure, show an error and exit.
  * Related requirements: F-005

* BD-FLOW-02 Start ratta

  * BD-FB-001 reads the application configuration file (config.json) and obtains the previous issue data folder path.
  * BD-FB-001 displays the project selection dialog and determines the issue data folder path (showing the previous value as the default).
  * BD-FB-001 saves the determined issue data folder path to the application configuration file.
  * BD-FB-002 checks whether the Contractor configuration file (auth/contractor.json) exists; if it exists, display the password verification dialog and determine the mode.

* BD-FLOW-03 Load issue data

  * BD-FB-010 scans the directories corresponding to categories, reads issue JSON files, and validates them.
  * Keep unreadable / schema-mismatched files as an error list.

* BD-FLOW-04 Display screens

  * BD-FB-004 displays the category list and the issue list of the selected category.
  * If errors exist, the user can open the error details dialog from a menu or an error menu.

* BD-FLOW-05 Issue operations

  * BD-FB-005 provides dialogs for issue creation/editing, and BD-FB-006 controls status changes.
  * On save, write back to JSON files via BD-FB-010.

* BD-FLOW-06 Comment and attachment operations

  * BD-FB-007 appends comments, and DB-FB-007 saves attachments and updates reference information in JSON.

* BD-FLOW-07 Category management

  * In Contractor mode only, add/delete categories via BD-FB-009 and reflect them in the directory structure.

---

## BD-PER Performance

### BD-PERF-001 Performance targets

Performance targets for ratta are as follows:

* BD-PT-001 Issue list display per category

  * Target counts: 500 issues, display 20 per page
  * Initial display: within 2 seconds
  * Apply sort/filter: within 1 second

* BD-PT-002 Issue detail screen display

  * Target comment count: assume 200
  * Detail display: within 1 second

* BD-PT-003 Re-display after comment creation

  * Update comment list: within 1 second

### BD-PERF-002 Performance design approach

* Since issue and comment counts are limited to several hundred, bulk-loading JSON per category at startup is practically acceptable.
* Validate and format JSON on the Go side, and pass only information required for display to the front-end.
* If performance problems become significant in the future, consider introducing an index cache file or partial loading in and after detailed design.

---

## BD-IO Input/output data specification (overview)

### BD-IO-001 Input data

* BD-IN-001 User input

  * Issue information: title, description, requested response due date, priority, status, etc.
  * Comment information: comment body (Markdown), author name, company type, etc.
  * Attachments: files linked to comments

* BD-IN-002 Configuration files

  * Application configuration file (config.json)
  * Contractor-only configuration file (auth/contractor.json)

* BD-IN-003 Existing JSON files

  * Issue JSON files

### BD-IO-002 Output data

* BD-OUT-001 Issue JSON files

  * One file per issue
  * Includes issue information, comment information, and attachment information.

* BD-OUT-002 Attachments

  * Binary files linked to each issue/comment.

* BD-OUT-003 Log files

  * Log file in the format logs/ratta.log.

### BD-IO-003 Main data items (examples)

* BD-DATA-001 Issue JSON (one item)

  * issue_id: Issue identifier (9-digit ID generated by nanoid)
  * category: Category name (corresponds to directory name)
  * title: Issue title
  * description: Issue description
  * status: Status
  * priority: Priority
  * origin_company: Origin company type (Contractor/Vendor)
  * assignee: Assignee name
  * created_at: Created datetime
  * updated_at: Updated datetime
  * due_date: Requested response due date
  * comments: Comment array
  * attachments: Attachments linked to the whole issue (if needed)
  * version: Schema version, etc.

* BD-DATA-002 Comment element

  * comment_id: Comment identifier (UUID v7)
  * body: Comment body (Markdown string)
  * author_name: Author name
  * author_company: Author company type (Contractor/Vendor)
  * created_at: Comment posted datetime
  * attachments: Attachment reference information array

* BD-DATA-003 Attachment reference information

  * attachment_id
  * file_name
  * relative_path (relative path from the issue folder)
  * mime_type
  * size_bytes, etc.

### BD-IO-004 Character encoding and formats

* JSON file encoding: UTF-8
* Comment size limit: 100KB or less in UTF-8 bytes
* Datetime storage format: ISO 8601 (with timezone) as a baseline.

  * Manage with second-level precision; do not handle sub-millisecond values (conform to Requirements Definition D-004).
* Datetime screen display format: Japanese notation (conform to Requirements Definition D-004)

  * Format: YYYY年MM月DD日 hh時mm分ss秒 (24-hour format)
  * When displaying, parse the stored value (timezone-included ISO 8601), convert it to the client PC local time, and format as above.
  * Example: 2025-12-06T10:30:00+09:00 → 2025年12月06日 10時30分00秒
* For JSON files, use 2-space indentation and LF for newlines.

  * In the JSON schema, fix key order by providing x-keyOrder.

---

## BD-SCREEN Screen specification (overview)

### BD-SCREEN-001 Overall screen structure

ratta consists of the following screens:

* BD-SC-000 Project selection dialog

  * Select/create the issue data folder (project) at startup
  * Show the previous issue data folder path as the default value

* BD-SC-001 Main screen

  * Category list (left pane)
  * Issue list (right pane)

* BD-SC-002 Issue detail dialog

  * Display/edit issue information
  * Display comment list / add comment
  * Attachment operations

* BD-SC-003 Error details dialog

  * List JSON read/validation errors, etc.

* BD-SC-004 Contractor authentication dialog

  * Password input when starting in Contractor mode

### BD-SCREEN-002 Details by screen

#### BD-SC-000 Project selection dialog

* Input / display

  * Issue data folder path (initially show the previous value saved in the configuration file (config.json))
* Operations

  * Open: validate the input path; if no issue, determine it as the issue data folder
  * Browse: select a path via a folder selection UI and reflect it in the input field
  * Create new: create a project folder at the specified path and determine it as the issue data folder
  * Cancel: exit the application
* Errors

  * If the folder does not exist, cannot be created, or cannot be accessed, show an error message and do not perform determination
* Post-determination processing

  * Save the determined issue data folder path to the configuration file

#### BD-SC-001 Main screen

* Header

  * Project name or data folder path
  * Current mode display (Contractor mode / Vendor mode)

* Left pane: Category list

  * Display directory structure as a tree or list
  * Provide category add/delete operations in Contractor mode only.

* Right pane: Issue list

  * Main display items

    * Issue ID
    * Title
    * Status
    * Priority
    * Origin company type (Contractor/Vendor)
    * Updated datetime (display format follows 8.4)
    * Requested response due date (display format follows 8.4)
  * Operations

    * Click a row to open the issue detail dialog
    * Sort by clicking column headers
    * Simple filters by status/priority/due date
    * Paging (20 items per page)

#### BD-SC-002 Issue detail dialog

* Basic information

  * Title (required)
  * Description (required)
  * Requested response due date (required)
  * Priority (required)
  * Status (with permission and transition constraints)
  * Origin company type (display only)
  * Created datetime / updated datetime (display only; display format follows 8.4)

* Comment display

  * Comment list sorted oldest-first
  * Render comment body with markdown-it and display line breaks, lists, etc. properly.
  * Display comment posted datetime (created_at) following the format in 8.4.

* Comment add area

  * Comment body input (Markdown text)
  * Author name input
  * Company type is determined by mode or by an input method (details conform to requirements).
  * Attachment file selection field

* Status change

  * Control selectable statuses according to Contractor/Vendor mode.

  * Status types (Japanese is for display; English is for internal representation)

    * 新規: Open
    * 対応中: Working
    * 問い合わせ中: Inquiry
    * 保留: Hold
    * 差し戻し: Feedback
    * 完了: Resolved
    * クローズ: Closed
    * 却下: Rejected

  * Status transition diagram (overview)
    Related requirements: F-004

    Closed / Rejected are terminal states and no further transitions occur.

    ```mermaid
    stateDiagram-v2
        [*] --> Open
        Open --> InProgress
        state InProgress {
            [*] --> Working
            [*] --> Inquiry
            [*] --> Hold
            Feedback --> Working
            Feedback --> Inquiry
            Feedback --> Hold

            Working --> [*]
            Inquiry --> [*]
            Hold --> [*]

            Working --> Inquiry
            Working --> Hold
            Inquiry --> Working
            Inquiry --> Hold
            Hold --> Working
            Hold --> Inquiry
        }
        InProgress --> Resolved
        Resolved --> Closed
        Resolved --> Feedback

        InProgress --> Rejected
        Closed --> [*]
        Rejected --> [*]
    ```

    Notes:

    * In Vendor mode, only transitions within Open / InProgress (Working, Inquiry, Hold, Feedback) / Resolved are allowed.
    * Transitions to Closed / Rejected are Contractor mode only.

#### BD-SC-003 Error details dialog

Related requirements: D-006, UI-004

* Items to display

  * JSON files that cannot be parsed
  * Schema mismatches such as missing required fields or type mismatches
* Display items (minimum)

  * File name
  * Error type
  * Error message
  * File path display and copy operation
* Policy

  * Users recover from errors manually (D-006)

#### BD-SC-004 Contractor authentication dialog

Related requirements: SY-002, F-002

* Display condition: when auth/contractor.json exists at startup
* Input: password
* Result

  * Verification success: start in Contractor mode
  * Verification failure: show a modal startup error and exit the application

---

## BD-FILES Data file specification

### BD-FILES-001 Issue data folder structure

Under the issue data folder, structure by category and issue as follows:

* <PROJECT_ROOT>/

  * <CategoryName>/

    * <issue_id>.json
    * <issue_id>.files/

      * Attachment files (<attachment_id>_<original_file_name>)

Notes:

* issue_id / attachment_id use 9-digit IDs generated by nanoid. Category names are treated as directory names.
* For category names, set UI constraints such as forbidden characters (example: \ / : * ? " < > |) and trailing dots.
* For attachment saved names (<attachment_id>_<original_file_name>), original_file_name may contain Windows forbidden characters, so sanitize it.
* If attachment file names collide, avoid by adding a suffix.

### BD-FILES-002 Configuration files

* BD-FC-001 Application configuration file

  * Location: same directory level as ratta.exe
  * Expected items

    * Issue data folder path
    * Issue data folder path (previous determined value; used as the default at startup)
    * Log settings (if needed)
  * Format: JSON (details conform to the Requirements Definition)
  * Update rule

    * If the issue data folder path is determined in the project selection dialog at startup, update the configuration file with that path.
  * Example (JSON)

    * {
    * "project_root_path": "D:/share/projectA",
    * "log": { "level": "info" }
    * }

* BD-FC-002 Contractor-only configuration file

  * Location: under the auth folder at the same level as the exe

    * auth/contractor.json
  * Generation method: generate by `ratta.exe init contractor` (if an existing file exists, do not overwrite unless `--force` is specified).
  * Format: JSON
  * Example expected items

    * mode: "contractor"
    * encrypted_password: encrypted password
  * Store the password encrypted. Use a design based on a symmetric cipher such as AES-256-GCM and key derivation via PBKDF2-HMAC-SHA256 (detailed values are defined in security design).

### BD-FILES-003 Log files

* Location: under the log folder at the same level as the exe

  * logs/ratta.log

    * Log file name when rotated: logs/ratta.log.1

* Rotation specification

  * Max size per file: 1MB
  * Max generations: 3 generations

* Log levels

  * Normal operation: INFO, ERROR
  * Verbose mode (specified by startup argument): DEBUG, INFO, ERROR

* Recorded content

  * Overview of exceptions and errors (may include stack traces)
  * Critical errors during JSON read/validation
  * Record user input values only to the minimum extent necessary, and avoid leaving personal or confidential information in logs for long periods.

* Related requirements: T-006

---

## BD-DATABASE Database specification

This system does not use external databases such as RDBMS or NoSQL, and manages issue data using a flat-file structure based on JSON files.

Therefore, items such as “database list” and “record specification” in this chapter are out of scope, and data structures are defined in Chapters 8 and 10 as JSON/file specifications.

---

## BD-SAFETY Fault-tolerance measures

### BD-SAFETY-001 Measures for JSON writing

* Use of temporary files

  * When writing the target issue JSON, create a temporary file in the same directory, write all data, then rename to replace the original file.
  * Temporary file name: <issue_id>.json.tmp.<pid>.<timestamp>
  * Prevent issue JSON files from becoming partially written due to abnormal termination mid-write.

* Handling of fsync

  * Prioritize write performance and implementation simplicity; do not perform fsync to synchronize to disk.
  * Treat failures while disk cache is not flushed as within acceptable operational risk.

### BD-SAFETY-002 Detect JSON corruption and schema mismatches

* At startup and as needed, read and validate all issue JSON files.
* Detect schema mismatches such as JSON parse errors or missing required fields, and keep them as error information.
* In the error details dialog, display error types and messages per file so users can manually fix them.

### BD-SAFETY-003 Data synchronization conflicts

* Assume that conflicts caused by multiple ratta instances editing the same issue JSON simultaneously are resolved by merge processing in external version control systems such as git.
* If an inconsistent JSON is produced as a merge result, it is detected by the validation process described in 12.2.
* Do not adopt measures such as preventing overwrite accidents.

### BD-SAFETY-004 Handling leftover temporary files

* Detect .tmp files at startup, and delete them under certain conditions (such as elapsed time exceeding a threshold or size 0), or list them in the error list to notify users.
* Deletion conditions and thresholds are described in the detailed design.

---

## BD-EVAL Evaluation policy for the design content

Evaluate the validity of this basic design from the following perspectives:

* BD-EV-001 Consistency with requirements

  * Each requirement ID defined in the Requirements Specification / Requirements Definition corresponds to the functional blocks in Chapter 6 and the data specifications in Chapters 8/10.

* BD-EV-002 Consistency and absence of contradictions

  * No contradictions exist among mode switching, permission control, data structures, screen specifications, and log specifications.

* BD-EV-003 Feasibility

  * The performance targets in Chapter 7 are achievable under the assumed hardware configuration.

* BD-EV-004 Testability

  * The structure makes it easy to conduct unit tests for Go/Vue and E2E tests using Playwright.
  * Traceability can be established between FR/TR requirement IDs and test items.
