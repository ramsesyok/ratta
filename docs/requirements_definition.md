# Requirements Definition (Issue-Tracking Desktop Application)

## 1. Purpose of This Document

This document defines the concrete requirements of this system in order to realize the functional requirements (FR-xxx) and technical requirements (TR-xxx) defined in the “Requirements Specification.”
Each requirement is assigned a Requirement ID, and the corresponding Spec Requirement ID(s) are listed to ensure traceability to the Requirements Specification.

---

## 2. System Overview and Preconditions

### SY-001 Purpose and Scope of the System

Requirement ID: SY-001
Related Requirement ID(s): FR-001, FR-020

This system is a desktop application that manages issues for a single project shared between the Contractor and the Vendor. It replaces the current file-splitting operation, and its primary purpose is to leave the exchange of comments per issue in a form that is easy to review visually as a log. Also, regardless of whether the system is running in Contractor mode or Vendor mode, both companies can view all issues that exist in the same project, and no company-based visibility differences are introduced.

### SY-002 Contractor/Vendor Mode and Handling of User Names

Requirement ID: SY-002
Related Requirement ID(s): FR-003, FR-004, TR-007

This system determines the company mode at startup and operates either in Contractor mode or Vendor mode. The company mode is determined by whether a Contractor-only configuration file exists at a designated location and by its contents. If the Contractor-only configuration file exists, the system compares the password written in that file with the password input by the user at startup, and starts in Contractor mode only if they match. If they do not match, the system does not start in Contractor mode and treats this as a startup error. If the Contractor-only configuration file does not exist, the system starts in Vendor mode.

This system does not provide user management or login functionality. User names recorded in comments and issue creator names are entered manually as an arbitrary string by the user, and the application does not restrict the format or uniqueness of the user name.

### SY-003 Data Synchronization and Responsibility Boundary

Requirement ID: SY-003
Related Requirement ID(s): FR-005, TR-006, TR-009, TR-010

This system does not implement issue data synchronization or conflict resolution as internal functionality, and directly reads and writes only the JSON files stored in a local or shared folder. Data synchronization and merge processing between the Contractor and the Vendor are assumed to be performed operationally using a version control tool such as git.

Backup and restore of issue data are delegated to the existing file system and backup operations, and this system is responsible only for creating and updating the issue JSON files. The system runs standalone in an on-premises environment without internet connectivity, and does not depend on external cloud services or external APIs.

---

## 3. Functional Requirements

### 3.1 Project and Mode

#### F-001 Project Unit and Startup Unit

Requirement ID: F-001
Related Requirement ID(s): FR-002

This system targets only one project per application launch. The path of the target project’s data root folder is specified by a configuration file or by a startup selection screen. A path specified on the startup selection screen is saved to the configuration file and used as the default value on the next startup.

The system does not provide a feature to switch projects via screen operations after startup. Running multiple processes of this application simultaneously on the same PC, where each process handles a different project’s data root folder, is allowed.

#### F-002 How Contractor/Vendor Mode Is Determined

Requirement ID: F-002
Related Requirement ID(s): FR-004, TR-007

At startup, this system checks whether a Contractor-only configuration file exists at a designated path. If it exists, the system processes it as a candidate for Contractor mode. When the Contractor-only configuration file exists, the system accepts password input from the user, and starts in Contractor mode if the input matches the value in the configuration file. If they do not match, the system does not start in Contractor mode; it displays an error and exits.

If the Contractor-only configuration file does not exist, the system starts in Vendor mode. This prevents a terminal that should run in Contractor mode from being mistakenly used in Vendor mode.

#### F-003 Creating Issues and Recording the Originating Company

Requirement ID: F-003
Related Requirement ID(s): FR-014, FR-020

This system allows users to create new issues in both Contractor mode and Vendor mode. When registering a new issue, the issue information includes a field indicating the originating company. For issues created in Contractor mode, the system records the Contractor; for issues created in Vendor mode, the system records the Vendor.

Regardless of which company created an issue, all issues within the same project are viewable by both the Contractor and the Vendor.

#### F-004 Contractor/Vendor Permissions for Status Changes

Requirement ID: F-004
Related Requirement ID(s): FR-003, FR-011, FR-012, FR-013

This system applies different operation permissions for issue status changes depending on Contractor mode or Vendor mode.

In Contractor mode, the user can change the issue status among all statuses that belong to open states, and additionally is allowed to change directly from any open status to one of the finished statuses: “完了”, “クローズ”, or “却下”. On the other hand, once an issue reaches a finished state, it cannot be returned to an open state from any status.

In Vendor mode, the user can change among open statuses within the set: “新規”, “対応中”, “問い合わせ中”, “保留”, “差し戻し”, and “完了”. However, the user cannot change an issue’s status to “クローズ” or “却下”, and cannot change the status of an issue that is already in a finished state.

#### F-005 Initialization Command for the Contractor-Only Configuration File

Requirement ID: F-005
Related Requirement ID(s): FR-024

This system provides an initialization command that generates the Contractor-only configuration file (`auth/contractor.json`). The initialization command can be executed without launching the GUI, and accepts shared password input (including confirmation input).

The shared password is not saved in plaintext; it is stored in the configuration file in a protected form using a method such as encryption. If a generated configuration file already exists, the system exits with an error without overwriting it unless explicit overwrite is specified.

If generation succeeds, the system creates the `auth` folder, saves the configuration file under it, and exits.

### 3.2 Issue Information and Status Management

#### F-101 Issue Information Structure and Status Types

Requirement ID: F-101
Related Requirement ID(s): FR-007, FR-009, FR-010

Each issue handled by this system contains at least: issue identifier, category name, title, description, status, priority, originating company, created timestamp, updated timestamp, and requested response deadline.

The status values are eight types: “新規”, “対応中”, “問い合わせ中”, “保留”, “差し戻し”, “完了”, “クローズ”, and “却下”, defined as follows.

* “新規”: immediately after drafting, not yet handled
* “対応中”: work is in progress
* “問い合わせ中”: waiting for a response due to an inquiry to a third party, etc.
* “保留”: handling is paused for some reason
* “差し戻し”: the creator side requests reconsideration of the completion content
* “完了”: the responding side has completed work and is waiting for confirmation by the creator side
* “クローズ”: finished state where stakeholders including the creator agree that handling is complete
* “却下”: finished state where all parties judge that handling is unnecessary

“新規”, “対応中”, “問い合わせ中”, “保留”, “差し戻し”, and “完了” are treated as open states, and “クローズ” and “却下” are treated as finished states. Finished issues are not reopened; if handling is needed again, a new issue is created.

#### F-102 Handling of the Requested Response Deadline

Requirement ID: F-102
Related Requirement ID(s): FR-007, FR-008, FR-017

This system allows the user to enter a requested response deadline when drafting or editing an issue, and saves it as part of the issue information. The requested response deadline is displayed as date information on the issue list screen and the issue detail screen, and can be used on the issue list for sorting and simple filtering.

Advanced search conditions that combine multiple conditions (for example, “open state and overdue”) are treated as future extensions and are not included in the current requirements.

#### F-103 Management of Created and Updated Timestamps for Issues

Requirement ID: F-103
Related Requirement ID(s): FR-019

This system manages created and updated timestamps for all issues. The created timestamp records the time when the issue is first saved. The updated timestamp records the most recent time when a change occurs in issue information (status, description, requested response deadline, etc.).

On the issue list screen and the issue detail screen, at least the updated timestamp is displayed so that the user can identify the most recent change time of an issue.

### 3.3 Comments and Attachments

#### F-201 Comment Structure and Editing Policy

Requirement ID: F-201
Related Requirement ID(s): FR-015, FR-016, FR-019

This system provides a function to accumulate comments for each issue in chronological order. Each comment contains: comment identifier, body, poster company classification, poster name, posted timestamp, and (as needed) updated timestamp (including future extensions).

The comment body is saved and displayed in Markdown format so that long text including line breaks and source code fragments can be handled. The maximum size per comment is approximately 100KB, and normal usage assumes that several tens of KB is sufficient for describing information.

Because comments emphasize log-like properties, the system does not provide editing or deletion as application functionality after registration. If physical deletion is required due to an obvious mistaken post or accidental inclusion of confidential information, it is handled operationally outside the application using git or a text editor.

In both Contractor mode and Vendor mode, users can append new comments to issues that are viewable.

#### F-202 Comment Order and Timestamp Display Format

Requirement ID: F-202
Related Requirement ID(s): FR-015, FR-016, FR-019

This system displays comments associated with an issue from oldest to newest, in a format where new comments are added below the existing list.

The posted timestamp is displayed using a Japanese date/time format, with the format as: `YYYY年MM月DD日 hh時mm分ss秒`. This allows Contractor and Vendor users to intuitively understand the comment timeline.

#### F-203 Attachments for Comments

Requirement ID: F-203
Related Requirement ID(s): FR-021, TR-011, FR-015

This system provides a function to associate attachments with comments. Supported attachment types are image files and text files (including source code files). Other file types are not mandatory requirements.

Multiple attachments can be associated with a single comment. On the issue detail screen, links or thumbnails to each attachment are displayed together with the comment body.

The attachment files themselves are stored in a designated directory under the project, and the comment holds information that can uniquely identify the attachment (such as file name or relative path).

### 3.4 UI Requirements

#### UI-000 Project Selection at Startup

Requirement ID: UI-000
Related Requirement ID(s): FR-002

At startup, this system can provide a screen (dialog) for selecting the target project’s data root folder. The dialog displays the previous data root folder path saved in the configuration file as the default value.

The user can choose one of the following: open using the default value, browse and open a different folder, or create a new project folder and open it. If the user confirms and opens a target folder, the system saves the confirmed data root folder path to the configuration file.

If the user cancels, the system aborts startup and exits. If the specified folder does not exist or cannot be accessed, the system displays an error and allows re-selection.

#### UI-001 Issue List Display by Category

Requirement ID: UI-001
Related Requirement ID(s): FR-006, FR-017, FR-020

This system provides a screen that displays an issue list for each category under the project. The user first selects a category, and then can view a list of all issues that belong to that category.

In both Contractor mode and Vendor mode, all categories and issues within the same project are viewable, and the system does not restrict display targets based on company mode.

#### UI-002 Items Displayed in the Issue List

Requirement ID: UI-002
Related Requirement ID(s): FR-007, FR-017

On the issue list screen, this system displays at least the following items for each issue: issue title, status, priority, originating company, updated timestamp, and requested response deadline.

As needed, the issue identifier is also displayed so that the user can uniquely identify an issue. These items are used by the subsequent sorting and filtering functions.

#### UI-003 Sorting and Filtering of the Issue List

Requirement ID: UI-003
Related Requirement ID(s): FR-017, FR-022

This system provides sorting and simple filtering for the issue list using status, updated timestamp, requested response deadline, and priority. For example, the user can sort by higher priority first, or display only a specific status.

Advanced search conditions that combine multiple conditions (such as “open state and overdue”) are not included in the current requirements.

The default number of issues displayed per page is 20, and the list can be switched using paging.

#### UI-004 Display Method for Finished-State Issues

Requirement ID: UI-004
Related Requirement ID(s): FR-010

This system displays issues in finished states (“クローズ” and “却下”) as grayed out in the issue list. This allows the user to visually distinguish open-state issues from finished-state issues on the list screen.

Finished-state issues are still included as view targets, and the list display is designed so that it is easy to infer (from the list display as well) that status changes from finished state back to open state are not possible.

#### UI-005 Issue Detail Screen and Edit Mode

Requirement ID: UI-005
Related Requirement ID(s): FR-007, FR-008, FR-009, FR-015, FR-016

This system provides an issue detail screen that displays an issue’s detailed information and its associated comments. The initial state of the issue detail screen is view mode, and the user explicitly selects an “Edit” button to switch to edit mode.

In edit mode, the user can update items within the range allowed by Contractor mode or Vendor mode, such as issue title, description, requested response deadline, priority, and status.

Issue title, description, requested response deadline, and priority are required input items. The labels of these items show a red “＊” mark to indicate that they are required. Because the category is determined by the folder structure, the screen only displays the automatically set value and does not allow changes through normal operations.

Status change and comment addition are treated as separate operation flows. Status changes are handled as issue information update operations, while comment addition is handled as a dedicated append-only operation.

#### UI-006 Adding a Comment

Requirement ID: UI-006
Related Requirement ID(s): FR-015, FR-016

This system provides an “Add Comment” button for all viewable issues. When the user selects the button, the system shows an area for entering the body in Markdown format, a poster name input field, and (as needed) an attachment selection field.

When the comment is submitted, it is added to the end of the existing comment list. The system does not provide edit or delete buttons for registered comments; comments are always append-only.

### 3.5 Category Structure Management

#### F-301 Permissions for Managing Category Structure

Requirement ID: F-301
Related Requirement ID(s): FR-006, FR-018, TR-003

For category structure under a project (adding, deleting, renaming categories), this system allows change operations only in Contractor mode.

When a category is added in Contractor mode, the system creates a new category subfolder directly under the project’s data root folder. When deleting a category, the operation assumes that no issue files exist in that category, or that issues have been moved to another category before deletion.

For ease of management, category names and category subfolder names use the same name in principle.

In Vendor mode, only viewing the category structure is permitted, and operations to add, delete, or rename categories are not allowed.

---

## 4. Data Requirements

#### D-001 Directory Structure and File Placement

Requirement ID: D-001
Related Requirement ID(s): TR-001, TR-002, TR-003, FR-006

Project data handled by this system uses a directory structure where category subfolders exist under the data root folder. Each issue is saved as a single JSON file (one issue per file). Files attached to an issue are stored under a subfolder such as `<issue_id>.files`.

The path to the project’s data root folder is specified by a configuration file or by a startup selection screen, and is fixed at application startup. A path confirmed on the startup selection screen is saved to the configuration file.

#### D-002 Issue JSON File Name and Issue Identifier

Requirement ID: D-002
Related Requirement ID(s): TR-002, TR-004, FR-007

This system uses a file naming convention where the issue identifier is given the extension “.json” as the issue JSON file name. For example, if the issue identifier is “abc123456”, the corresponding file name is “abc123456.json”.

The issue identifier is managed to be unique within a project. The specific numbering method (sequential numbers, date + sequential number, etc.) is defined in the design phase.

The application operates assuming the mapping between issue identifier and file name, and loads the corresponding JSON file by identifying it from the issue identifier when displaying the issue list or issue details.

#### D-003 JSON Schema and Common Fields

Requirement ID: D-003
Related Requirement ID(s): TR-004, FR-007, FR-015, FR-019

Issue JSON files handled by this system follow a common JSON schema. The schema includes at least the following fields: issue identifier, category name, title, description, status, priority, originating company, created timestamp, updated timestamp, requested response deadline, comment array, assignee name, attachment reference information, and a “version” field indicating the schema version.

Each element of the comment array includes: comment identifier, body, poster company classification, poster name, posted timestamp, and attachment reference information.

The “version” field is for internal management to prepare for future schema extension and migration, and is not used for normal screen display. The schema is fixed at the current stage, and the detailed field structure and data types are defined in the subsequent design phase.

#### D-004 Character Encoding and Saved Date/Time Format

Requirement ID: D-004
Related Requirement ID(s): TR-008, FR-019

All JSON files saved by this system use UTF-8 as the character encoding.

Fields that represent date/time (created timestamp, updated timestamp, comment posted timestamp, etc.) are saved in ISO 8601 format including time zone information; for example: `2025-12-06T10:30:00+09:00`. Internally, the system manages timestamps with second-level precision, and does not handle milliseconds or smaller units.

On the screen, timestamps are converted and displayed in Japanese notation format: `YYYY年MM月DD日 hh時mm分ss秒`.

#### D-005 JSON File Formatting Rules

Requirement ID: D-005
Related Requirement ID(s): TR-004

When writing issue JSON files, this system applies a formatting rule with a two-space indentation width, and fixes the key order. This prevents unnecessary diffs when viewing changes with a version control tool such as git.

The key order is defined in the subsequent design phase, and the application always outputs JSON in the same order.

#### D-006 Handling of Corrupted Files and Schema-Inconsistent Files

Requirement ID: D-006
Related Requirement ID(s): TR-004, TR-006

When loading the issue list, this system parses each issue JSON file and detects files that cannot be interpreted as JSON or files that are missing required fields.

For files that fail JSON parsing, the system does not completely exclude them from the issue list without notice; instead, it displays at least the file name and an error message in the issue list screen or in a dedicated error display area. This enables the user to perform recovery work manually.

For files that do not match the schema but from which some fields can be read, the system displays the readable items as much as possible and clearly indicates the schema inconsistency using an icon or message.

---

## 5. Technical Requirements

#### T-001 Adopted Technologies and Application Form

Requirement ID: T-001
Related Requirement ID(s): TR-001, TR-009

This system is implemented as a desktop application using Wails, and adopts Vue 3 and Vuetify 3 for the frontend. The application runs locally on each user’s PC and does not assume a server-resident web application or a browser-based system.

The system can operate in an on-premises environment without internet connectivity and does not depend on external cloud services or external APIs.

#### T-002 Supported OS and Architecture

Requirement ID: T-002
Related Requirement ID(s): TR-001, TR-009

The officially supported OS for this system is Windows, targeting 64-bit environments on Windows 10 or later. The system assumes a version where the WebView used by Wails is available as standard.

Operation on Linux environments is not a mandatory requirement and is out of scope for this requirements definition. 32-bit OS environments are not supported, and installation and operation are not guaranteed.

#### T-003 Application Distribution Format

Requirement ID: T-003
Related Requirement ID(s): TR-001

This system is distributed as a zip archive rather than an installer. Users extract the distributed zip file to an arbitrary folder and use the application by starting the executable within the extracted folder.

For version upgrades, the design aims to support upgrades in principle by replacing only the application itself and the configuration file, and considers backward compatibility so that data schema migration work is not required.

#### T-004 Safe Writing When Updating JSON Files

Requirement ID: T-004
Related Requirement ID(s): TR-005

When updating an issue JSON file, this system does not overwrite the existing file directly. Instead, it writes the entire content to a temporary file and then replaces it with the original file name using a file-system rename operation.

This approach reduces the risk that an existing issue file will be corrupted even if the application terminates abnormally during writing.

#### T-005 Relationship Between Locking and Git Operations

Requirement ID: T-005
Related Requirement ID(s): TR-006

This system does not implement an exclusive lock mechanism within the application for issue files. Conflicts when updates occur to the same issue from multiple terminals are delegated to merge processing by a version control tool such as git.

The application is responsible for correct reading and writing of JSON files and for detection of corruption or schema inconsistency. If an inconsistent JSON file is produced as a merge result, the system notifies the user via the error display defined in D-006.

#### T-006 Log Output

Requirement ID: T-006
Related Requirement ID(s): None (supplemental requirement added during the requirements definition phase)

This system outputs internal error logs as files. The log output destination is under the application folder, and the system implements a rotation function based on log file size or number of generations.

Normal operation logs and access logs are not mandatory requirements; the system records only the information necessary for error analysis.

---

## 6. Non-Functional and Operational Requirements

#### NF-001 Assumed Usage Scale

Requirement ID: NF-001
Related Requirement ID(s): FR-002, FR-005, TR-009, TR-012

The number of users is assumed to be 10 or fewer for each of the Contractor and the Vendor. Simultaneous use for the same project is assumed to be up to about 10 users, and performance is ensured such that displaying the issue list and displaying/updating issue details can be performed with practical response times at this usage scale.

The number of issues and comments is assumed to be on the order of several hundred, and extremely large-scale project management is out of scope.

#### NF-002 Git Synchronization Operations Between the Contractor and the Vendor

Requirement ID: NF-002
Related Requirement ID(s): FR-005, FR-023, TR-006, TR-010, TR-013

In this system, data synchronization between the Contractor and the Vendor is performed operationally using external tools such as git. As an operational precondition, the Contractor deploys the repository for the Vendor after the morning meeting and after the end-of-day meeting, and the Vendor sends the repository to the Contractor around noon and after the end-of-day meeting.

Based on these synchronization timings, the application does not require real-time synchronization functionality.

---

The above is the requirements definition of this system as confirmed at this time.
In the subsequent detailed design phase, the mapping between the Requirement IDs in this document and the FR/TR IDs in the Requirements Specification will be maintained, while performing concrete design such as screen layouts, JSON schema definitions, and numbering schemes.
