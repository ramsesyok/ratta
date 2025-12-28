# 詳細設計書（課題管理デスクトップアプリケーション ratta）

---

## DD-SCOPE-001 文書の目的と範囲

### DD-SCOPE-002 目的

* 基本設計で定義された構成・機能・データ仕様を、実装可能な粒度（モジュール、I/F、画面、データ、例外、テスト）に落とし込む。
* 生成AIが実装タスクへ分解しやすいよう、各セクションを独立した実装単位として記述する。

### DD-SCOPE-003 範囲

* Wails + Go（バックエンド）と Vue 3 + Vuetify 3（フロントエンド）の実装詳細
* JSON スキーマとファイル配置、I/O、破損検出、エラー表示
* 画面、状態管理、入力バリデーション
* CLI 初期化コマンド（`ratta.exe init contractor`）
* テスト方針（Go/Vue/Playwright）

---

## DD-TERM-001 用語・前提

### DD-TERM-002 用語

* Project Root: 課題データフォルダ（`<PROJECT_ROOT>`）
* Category: `<PROJECT_ROOT>` 直下のサブフォルダ（シート相当、フラット構成のみ）
* Issue: 1課題 = 1 JSON（`<issue_id>.json`）
* Attachment: `<issue_id>.files/` 配下のファイル
* Mode: Contractor / Vendor（起動時に決定）

### DD-TERM-003 非前提（やらないこと）

* 競合解消、同期、ロック（git 運用に委譲）
* バックアップ、リストア（運用に委譲）
* ユーザ管理、ログイン（持たない）
* カテゴリの多階層（サブフォルダをカテゴリとして扱うことはしない）


---

## DD-ARCH-001 全体アーキテクチャ

### DD-ARCH-002 構成

* Backend（Go + Wails）

  * 課題データのロード、検証、永続化
  * モード判定、権限制御（許可判定）
  * エラーログ出力、エラー情報の集約
  * CLI 初期化コマンド
* Frontend（Vue 3 + Vuetify 3）

  * 画面、入力、画面内バリデーション
  * 表示用整形（日時表示、ステータス表示名）
  * エラー表示の統一（エラー詳細ダイアログ）

### DD-ARCH-003 Wails 連携方針

* Frontend から Backend の公開メソッドを呼ぶ（Wails binding）
* Backend は画面に必要な DTO を返す（Go の内部構造を直接晒さない）
* エラーは共通エラー DTO で返し、Frontend はエラー詳細ダイアログに集約表示する

---

## DD-DIR-001 ディレクトリ設計

### DD-DIR-002 実行時ディレクトリ（ratta.exe と同階層）

* `ratta.exe`
* `config.json`（アプリ設定）
* `auth/contractor.json`（存在する場合、Contractor 候補）
* `logs/ratta.log`（ローテーションあり）

### DD-DIR-003 Project Root（課題データ）

* `<PROJECT_ROOT>/<category>/<issue_id>.json`
* `<PROJECT_ROOT>/<category>/<issue_id>.files/<attachment_file>`
* `<PROJECT_ROOT>` 配下に `.git/` 等が存在してもよい（走査で除外する）
* `<PROJECT_ROOT>/<category>` はフラットのみ（`<category>/<subdir>` をカテゴリとして走査しない）

---

## DD-BE-001 Backend 設計（Go）

### DD-BE-002 パッケージ構成（案）

* `cmd/ratta/`（Wails エントリ）
* `internal/app/`（アプリユースケース）
* `internal/domain/`（ドメインモデル、バリデーション）
* `internal/infra/fs/`（ファイル I/O、アトミック更新、走査）
* `internal/infra/schema/`（JSON Schema 検証）
* `internal/infra/log/`（ロガー、ローテーション）
* `internal/infra/crypto/`（contractor.json の検証・生成）
* `internal/present/`（Wails 公開 DTO、エラー DTO）

設計方針

* app と infra の境界を保つ（ファイル操作や暗号等は infra に閉じる）
* present で UI 向け DTO へ変換し、domain を直接返さない

### DD-BE-003 公開 API（Wails binding）設計

Backend は Wails binding として公開メソッドを提供し、Frontend（Vue）から呼び出される。

#### DD-BEAPI-001 API 共通方針

- Backend の公開メソッドは、UI に必要な DTO を返す（Go の内部構造を直接返さない）
- 失敗時は、共通エラー DTO（DD-BE-004）へ変換可能な形でエラーを返す
  - Go メソッドは ResponseDTO（ok, data, error）を返し、Frontend 側は ok を判定することとする。
- すべての更新系 API はアトミック更新（DD-PERSIST）を適用する
- 権限制御（Contractor/Vendor）は Backend 側で必ず最終判定する（Frontend は UI で候補を絞るのみ）
- スキーマのバージョンが未対応の場合は is_schema_invalid=true として読み取りのみ許可、更新不可とする
  - Issue.version は 1のみ対応。1以外は is_schema_invalid=true とし、一覧表示は警告付き、更新系は不可、詳細は可能な範囲で表示。

#### DD-BEAPI-002 API 一覧と概要

設定／起動

- GetAppBootstrap(): BootstrapDTO
  - 概要
    - 起動直後に必要な設定値を返す（前回の Project Root、UI 設定、ログ設定、auth/contractor.json の有無など）
  - 主な呼び出し元
    - ProjectSelectDialog（初期表示）
  - 失敗時
    - config.json 読み込みに失敗した場合はデフォルト値で継続し、警告としてログ出力する（致命ではない）

- ValidateProjectRoot(path: string): ValidationResultDTO
  - 概要
    - 指定パスが Project Root として利用可能か検証する（存在、アクセス権、ディレクトリであること等）
  - 主な呼び出し元
    - ProjectSelectDialog（開く、または新規作成前の確認）
  - 失敗時
    - 検証不能（I/O 例外等）は error として返す。検証結果 NG は DTO の is_valid=false で返す

- CreateProjectRoot(path: string): ValidationResultDTO
  - 概要
    - 指定パスに Project Root を作成する（存在しない場合のみ）
  - 主な呼び出し元
    - ProjectSelectDialog（新規作成）
  - 副作用
    - ディレクトリ作成のみ。カテゴリは作らない
  - 失敗時
    - 作成不可（権限、既存ファイル等）は error として返す

- SaveLastProjectRoot(path: string): void
  - 概要
    - config.json の last_project_root_path を更新する
  - 主な呼び出し元
    - ProjectSelectDialog（Project Root 確定後）
  - 副作用
    - config.json をアトミック更新する

モード判定

- DetectMode(): ModeDTO
  - 概要
    - auth/contractor.json の有無を確認し、パスワード要求の有無を返す
    - パスワード照合自体は VerifyContractorPassword で行う
  - 主な呼び出し元
    - 起動直後（Project Root 確定後）

- VerifyContractorPassword(password: string): ModeDTO
  - 概要
    - auth/contractor.json の暗号データを用いて、入力パスワードが正しいか検証する
  - 主な呼び出し元
    - ContractorPasswordDialog
  - 失敗時
    - 照合失敗は E_PERMISSION として返す（message は「パスワード照合に失敗しました」）
    - ContractorPasswordDialog が照合失敗メッセージを表示し、OK押下でアプリを終了する

カテゴリ／課題

- ListCategories(): CategoryListDTO
  - 概要
    - Project Root 直下のカテゴリ（サブフォルダ）一覧を返す（フラットのみ、除外規則は DD-LOAD-002）
    - .tmp_rename 配下に残っているカテゴリは読み取り専用カテゴリとして一覧に含める（CategoryDTO.is_read_only=true）
  - 主な呼び出し元
    - MainView（左ペインの初期表示・更新）

- CreateCategory(name: string): CategoryDTO
  - 概要
    - カテゴリディレクトリを作成する（Contractor のみ）
　- 禁則処理
    - カテゴリ名の重複・大小文字違いは許容せずエラーとする
  - 失敗時
    - Vendor モードの場合は E_PERMISSION
    - 禁止文字や長さ超過は E_VALIDATION

- RenameCategory(oldName: string, newName: string): CategoryDTO
  - 概要
    - カテゴリ名（ディレクトリ名）を変更する（Contractor のみ）
    - カテゴリ名変更時、既存Issue JSONのcategoryフィールドを更新する
  - 失敗時
    - Vendor モードの場合は E_PERMISSION
    - newName が不正な場合は E_VALIDATION
    - oldName が存在しない場合は E_NOT_FOUND
    - .tmp_rename が残っていてリカバリが必要な場合は E_CONFLICT
  - 処理概要
    - 事前検証
      - newName のバリデーション（禁止文字、末尾、長さ、大小文字違い含む重複判定）
      - oldName 存在確認
      - <PROJECT_ROOT>/.tmp_rename 配下にディレクトリが残っている場合はリカバリ未完了として E_CONFLICT
    - 実処理（推奨順）
      1. <PROJECT_ROOT>/<old> を <PROJECT_ROOT>/.tmp_rename/<new> にリネーム（同一ボリューム前提）
      2. 移動後フォルダ配下の *.json を走査し、各 issue.category を newName に更新してアトミック更新
      3. 問題なければ .tmp_rename/<new> を最終 <PROJECT_ROOT>/<new> にリネーム
    - 失敗時（ロールバック境界を明確化）：
      - 手順1〜2（フォルダ操作）で失敗した場合：
        - 処理を中止する
        - フォルダ名の復帰を行う（可能なら categories/<oldName> に戻す）
        - 復帰に失敗した場合は .tmp_rename を残し、起動時検出の対象とする
      - 手順3（issue.category 更新）開始後に失敗した場合：
        - 処理を中止する
        - issue.category のロールバックは行わない（手動リカバリ運用）
        - .tmp_rename を残し、該当カテゴリは読み取り専用として扱う（編集不可）

  - 備考
    - カテゴリ名の重複・大小文字違いは許容せずエラーとする

- DeleteCategory(name: string): void
  - 概要
    - 空のカテゴリディレクトリを削除する（Contractor のみ）
  - 失敗時
    - Vendor モードの場合は E_PERMISSION
    - 読み取り専用カテゴリの場合は E_CONFLICT
    - 空でない場合は E_CONFLICT（課題が残っている）
  - 備考
    - 空のカテゴリとは、*.jsonが無い場合とする。また*.jsonが無く、.filesだけある場合も空のカテゴリとする。

- ListIssues(category: string, query: IssueListQueryDTO): IssueListDTO
  - 概要
    - 指定カテゴリ配下の課題をロードし、一覧向け DTO を返す
    - スキーマ不整合は一覧に警告付きで含める（DD-LOAD-004）
  - 主な呼び出し元
    - MainView（右ペインの一覧）
  - キャッシュ
    - 一覧はキャッシュを前提とする（DD-LOAD-005）

- GetIssue(category: string, issueId: string): IssueDetailDTO
  - 概要
    - 課題詳細をディスクから再ロードして返す（最新化優先）
  - 主な呼び出し元
    - IssueDetailDialog を開く時
  - 失敗時
    - 破損・スキーマ不整合で詳細を構築できない場合は error
    - スキーマ不整合で部分表示する場合は is_schema_invalid を true として返す

- CreateIssue(category: string, dto: IssueCreateDTO): IssueDetailDTO
  - 概要
    - 新規課題 JSON を生成して保存し、作成した課題詳細を返す
  - 副作用
    - <issue_id>.json を新規作成する（comments は空配列で開始）
  - ルール
    - status は Open で開始する
    - origin_company は現在モードから自動設定する
  - 失敗時
    - category が存在しない場合は E_NOT_FOUND
    - category が読み取り専用カテゴリの場合は E_CONFLICT
    - 入力不正は E_VALIDATION

- UpdateIssue(category: string, issueId: string, dto: IssueUpdateDTO): IssueDetailDTO
  - 概要
    - 課題の更新を保存し、更新後の課題詳細を返す
  - ルール
    - 終了状態（Closed, Rejected）は更新不可
    - スキーマ不整合課題（is_schema_invalid=true）は更新不可
    - updated_at を更新する
  - 失敗時
    - 権限違反（Vendor で Closed/Rejected など）は E_PERMISSION
    - category が読み取り専用カテゴリの場合は E_CONFLICT
    - 競合は扱わない（git で解決）

コメント／添付

- AddComment(category: string, issueId: string, dto: CommentCreateDTO): IssueDetailDTO
  - 概要
    - コメントを追記し、添付があれば保存し、更新後の課題詳細を返す
  - ルール
    - コメント本文は UTF-8 bytes <= 100KB
    - コメント投稿と同時の添付のみを扱う（後付け API は設けない）
    - 添付ファイルは、1コメントにつき5個まで
  - 失敗時
    - 入力不正は E_VALIDATION
    - category が読み取り専用カテゴリの場合は E_CONFLICT
    - 添付保存失敗は E_IO_WRITE（課題 JSON は更新しない）
  - 処理概要
    - 添付ありの場合の手順例
      1. 入力バリデーション
      2. 添付保存先ディレクトリ作成（<issue_id>.files）
      3. 添付を一時名でコピー（例: <stored_name>.uploading.<pid>）
      4. 全添付が成功したら最終名へリネーム
      5. JSON の comments に追記しアトミック更新
      6. JSON 更新が成功したら完了
    - 途中失敗時の方針
      5.で失敗した場合は、添付を必ず削除してロールバックする。

エラー一覧

- GetLoadErrors(scope: ErrorScopeDTO): ErrorListDTO
  - 概要
    - 起動時やロード時に検出した破損・不整合・tmp 残骸等のエラー一覧を返す
  - 主な呼び出し元
    - ErrorDetailDialog
  - 補足
    - scope が category の場合は指定カテゴリのエラーのみ返す

#### DD-BEDTO-001 DTO フィールド定義

以下は Wails binding で利用する DTO の定義である。型表記は TypeScript 互換のイメージで記載する。

共通

- ModeToken
  - "Contractor" | "Vendor"

- CompanyToken
  - "Contractor" | "Vendor"

- StatusToken
  - "Open" | "Working" | "Inquiry" | "Hold" | "Feedback" | "Resolved" | "Closed" | "Rejected"

- PriorityToken
  - "High" | "Medium" | "Low"

BootstrapDTO

- has_config: boolean
  - config.json が存在し読み取れたか（読取失敗でも false で継続）
- last_project_root_path: string | null
- ui_page_size: number
  - 既定 20（config.json に保存する場合はそこから読む）
- log_level: "info" | "debug"
- has_contractor_auth_file: boolean
  - auth/contractor.json が存在するか（存在する場合は ContractorPasswordDialog を要求する）

ValidationResultDTO

- is_valid: boolean
- normalized_path: string
  - 正規化後パス（存在する場合）
- message: string
  - ユーザ向け短文
- details: string | null
  - 失敗理由の詳細（任意）

ModeDTO

- mode: ModeToken
  - ContractorPasswordDialog による照合完了までは常に "Vendor" を返す
  - 照合成功後に "Contractor" を返す（照合失敗時はエラーを返す）
- requires_password: boolean
  - true の場合、VerifyContractorPassword を呼ぶ必要がある

CategoryDTO

- name: string
- is_read_only: boolean
  - true の場合、当該カテゴリは読み取り専用（編集系 API は E_CONFLICT）
- path: string
  - 通常: <PROJECT_ROOT>/<category> の絶対パス
  - 読み取り専用カテゴリ: <PROJECT_ROOT>/.tmp_rename/<category> の絶対パス
- issue_count: number 

CategoryListDTO

- categories: CategoryDTO[]
- errors: number
  - カテゴリ走査に関するエラー件数（任意）

IssueListQueryDTO

- page: number
  - 1 始まり
- page_size: number
  - 既定 20
- sort_by: "updated_at" | "due_date" | "priority" | "status" | "title"
- sort_order: "asc" | "desc"
- filter_status: StatusToken | null
- filter_priority: PriorityToken | null
- filter_due_date_from: string | null
  - YYYY-MM-DD
- filter_due_date_to: string | null
  - YYYY-MM-DD

IssueSummaryDTO

- issue_id: string
- category: string
- title: string
- status: StatusToken
- priority: PriorityToken
- origin_company: CompanyToken
- updated_at: string
  - ISO 8601 with TZ
- due_date: string
  - YYYY-MM-DD
- is_schema_invalid: boolean
  - true の場合、警告付きで表示し、更新系操作は不可

IssueListDTO

- category: string
- total: number
- page: number
- page_size: number
- issues: IssueSummaryDTO[]

IssueCreateDTO

- title: string
- description: string
- due_date: string
  - YYYY-MM-DD
- priority: PriorityToken
- assignee: string | null

IssueUpdateDTO

- title: string
- description: string
- due_date: string
- priority: PriorityToken
- status: StatusToken
- assignee: string | null

AttachmentUploadDTO

- source_path: string
  - 端末上の選択ファイルの絶対パス（Frontend はファイル選択 UI で取得する）
- original_file_name: string
  - 選択時のファイル名（source_path から取得できる場合は省略可）
- mime_type: string | null
  - 推定できる場合のみ

CommentCreateDTO

- body: string
- author_name: string
- attachments: AttachmentUploadDTO[]
  - 添付なしの場合は空配列

AttachmentRefDTO

- attachment_id: string
- file_name: string
- stored_name: string
- relative_path: string
- mime_type: string | null
- size_bytes: number | null

CommentDTO

- comment_id: string
- body: string
- author_name: string
- author_company: CompanyToken
- created_at: string
  - ISO 8601 with TZ
- attachments: AttachmentRefDTO[]

IssueDetailDTO

- is_schema_invalid: boolean
- version: number
  - Issue JSON の version をそのまま返す（必須）
- issue_id: string
- category: string
- title: string
- description: string
- status: StatusToken
- priority: PriorityToken
- origin_company: CompanyToken
- assignee: string | null
- created_at: string
- updated_at: string
- due_date: string
- comments: CommentDTO[]

ErrorScopeDTO

- scope: "all" | "category"
- category: string | null

ErrorListDTO

- generated_at: string
  - ISO 8601 with TZ
- items: ErrorDTO[]
  - ErrorDTO の定義は DD-BE-004 に従う

### DD-BE-004 共通エラー DTO

* `error_code`（例: `E_IO_READ`, `E_SCHEMA_INVALID`, `E_PERMISSION`, `E_VALIDATION`）
* `message`（ユーザ向け短文）
* `detail`（開発者向け、スタック等。表示は折りたたみ）
* `target_path`（対象ファイルやフォルダ）
* `hint`（復旧の指針。例: git のマージ結果を確認）

#### エラーコード定義（列挙）
- E_IO_READ
- E_IO_WRITE
- E_IO_RENAME
- E_SCHEMA_INVALID
- E_JSON_PARSE
- E_VALIDATION
- E_PERMISSION
- E_NOT_FOUND
- E_CONFLICT（カテゴリ削除、名称重複など）
- E_CRYPTO（contractor.json 破損、復号失敗など）

### DD-BE-005 モード判定ロジック

* `auth/contractor.json` が存在しない場合

  * Vendor モードで起動
* 存在する場合

  * パスワード入力 UI を出し、照合成功で Contractor モード
  * 照合失敗は照合失敗メッセージを表示して終了

### DD-BE-006 JSON Schema 検証（実装方針）

* 使用ライブラリ
  * `github.com/santhosh-tekuri/jsonschema/v5`
* 対象スキーマ
  * `schemas/issue.schema.json`
  * `schemas/config.schema.json`
  * `schemas/contractor.schema.json`
* ドラフト方針
  * 各 schema ファイルに `$schema` を必ず明記する（ライブラリの「$schema 未指定時は実装済み最新ドラフト扱い」を避けるため）
* 参照（$ref）の取り扱い
  * schema のロード元はローカルファイル（schemas/ 配下）のみとし、HTTP 等の外部参照は許可しない
* エラー粒度（UI/ログへの出し方）
  * 課題 JSON の検証に失敗した場合
    * `is_schema_invalid=true` を付与する
    * エラー一覧に `E_SCHEMA_INVALID` として登録する
    * `detail` にはバリデーションエラーの要点（どのフィールドがなぜ不一致か）を格納する

---

## DD-CLI-001 CLI 初期化コマンド設計

### DD-CLI-002 コマンド

* `ratta.exe init contractor`
* オプション

  * `--force`（既存 `auth/contractor.json` を上書き）

### DD-CLI-003 入力

* 共有パスワード
* 確認入力（2回目一致必須）
* 端末のコンソールで入力を隠す

### DD-CLI-004 出力（auth/contractor.json）

* `auth/` フォルダがなければ作成
* 成功時に `auth/contractor.json` を生成して終了
* 失敗時は非0終了

### DD-CLI-005 パスワード保護方式（確定）

* 平文として固定文字列 `contractor-mode` を用意
* 入力パスワードから PBKDF2-HMAC-SHA256 で 32 bytes 鍵を導出

  * iteration: 200,000
  * salt: 16 バイト（ランダム）
* AES-256-GCM で固定文字列を暗号化し、復号できれば正しいパスワードと判定

  * nonce: 16 バイト（ランダム）
* `format_version` の開始値

  * 1

保存フィールド例

* `format_version: 1`
* `kdf: "pbkdf2-hmac-sha256"`
* `kdf_iterations: 200000`
* `salt_b64: <base64>`
* `nonce_b64: <base64>`
* `ciphertext_b64: <base64>`（GCM の tag 含む）
* `mode: "contractor"`

---

## DD-CONF-001 設定ファイル設計（config.json）

### DD-CONF-002 配置

* `ratta.exe` と同階層

### DD-CONF-003 最小フィールド

* `format_version: 1`
* `last_project_root_path: string`
* `log: { level: "info" | "debug" }`
* `ui: { page_size: 20 }`

### DD-CONF-004 更新ルール

* プロジェクト選択確定時に `last_project_root_path` を更新
* JSON 更新はアトミック更新方式を適用（tmp→rename）

---

## DD-DATA-001 データ仕様（課題JSON、コメント、添付）

### DD-DATA-002 JSON 共通

* 文字コード UTF-8
* インデント 2 スペース
* 改行 LF
* キー順序固定
  * 要件D-005に従い順序の統一を行う。具体的な順序（order list）の決定・設定は実装時に行う（本書では列挙しない）
* TZ は、OSのTimeZoneとする

### DD-DATA-003 Issue JSON（1課題）

* `version: int`（必須、1 で開始）
* `issue_id: string`（必須、nanoid 9桁）
* `category: string`（必須、ディレクトリ名と一致）
* `title: string`（必須、最大 255 文字）
* `description: string`（必須、最大 255 文字）
* `status: string`（必須、内部表現は英語トークン）
* `priority: string`（必須、`High|Medium|Low`）
* `origin_company: string`（必須、`Contractor|Vendor`）
* `assignee: string`（任意）
* `created_at: string`（必須、ISO 8601 with TZ、秒精度）
* `updated_at: string`（必須、ISO 8601 with TZ、秒精度）
* `due_date: string`（必須、`YYYY-MM-DD`）
* `comments: Comment[]`（必須、空配列可）

### DD-DATA-004 Comment

* `comment_id: string`（必須、UUID v7）
* `body: string`（必須、Markdown、UTF-8 bytes <= 100KB）
* `author_name: string`（必須、最大 255 文字）
* `author_company: string`（必須、`Contractor|Vendor`）
* `created_at: string`（必須、ISO 8601 with TZ、秒精度）
* `attachments: AttachmentRef[]`（必須、空配列可）

コメントは編集・削除しない（追記のみ）。

### DD-DATA-005 AttachmentRef とファイル保存名

AttachmentRef（JSON 側）

* `attachment_id: string`（必須、nanoid 9桁）
* `file_name: string`（必須、元ファイル名、最大 255 文字）
* `stored_name: string`（必須、保存ファイル名）
* `relative_path: string`（必須、`<issue_id>.files/<stored_name>`）
* `mime_type: string`（任意）
* `size_bytes: int`（任意）

保存先（実体）

* `<PROJECT_ROOT>/<category>/<issue_id>.files/<attachment_id>_<sanitized_original_name>`

サニタイズ仕様（Windows 禁止文字対策）

* `\ / : * ? " < > |` を `_` に置換
* 末尾の `.` と空白を削除
* サニタイズ後のファイル名部分は最大 255 文字までに切り詰める
* 衝突が起きた場合は `_<n>` サフィックスで回避

添付の必須性

* 添付は必須ではない（attachments は空配列可）

---

## DD-STAT-001 ステータスと権限制御

### DD-STAT-002 内部ステータス（英語トークン固定）

JSON には内部値（英語）を保存し、UI で日本語表示する。

* 新規: `Open`
* 対応中: `Working`
* 問い合わせ中: `Inquiry`
* 保留: `Hold`
* 差し戻し: `Feedback`
* 完了: `Resolved`
* クローズ: `Closed`
* 却下: `Rejected`

### DD-STAT-003 遷移制御

* 終了状態: `Closed`, `Rejected`

  * 終了状態に入った課題は以後変更不可（Contractor/Vendor とも）
* Vendor モード

  * `Closed` と `Rejected` へ遷移不可
* Contractor モード

  * オープン状態間遷移可
  * オープン状態から `Resolved|Closed|Rejected` へ遷移可

実装方針

* Backend にモード別許可ステータスと終了状態判定を持たせる
* Frontend ではセलेकタを絞るが、最終チェックは Backend で必須

---

## DD-LOAD-001 課題データ走査・ロード・キャッシュ

### DD-LOAD-002 カテゴリ走査

* `<PROJECT_ROOT>` 直下のディレクトリをカテゴリとする（フラットのみ）
* 例外: `<PROJECT_ROOT>/.tmp_rename/<category>` が存在する場合、その `<category>` を読み取り専用カテゴリとしてカテゴリ一覧に含める
* 除外対象

  * `.git`
  * 先頭 `.` のディレクトリ（ただし `.tmp_rename` は上記例外として扱う）
* サブフォルダはカテゴリとして扱わない（`<PROJECT_ROOT>/<category>/<subdir>` を再帰走査しない）

### DD-LOAD-003 課題走査

* カテゴリ配下の `*.json` を対象
* 読み取り専用カテゴリの場合、カテゴリ配下は `<PROJECT_ROOT>/.tmp_rename/<category>` を指す
* `<issue_id>.files/` などフォルダは除外
* JSON を読む際に以下に分類

  * パース不可: 破損ファイル
  * パース可だが必須欠落／型不整合: スキーマ不整合
  * 正常

### DD-LOAD-004 破損・不整合の扱い

* パース不可

  * 課題一覧には出さない
  * エラー一覧には必ず出す（ファイル名、パス、メッセージ）
* スキーマ不整合

  * 課題一覧に警告付きで出す
  * DTO に `is_schema_invalid: true` を付与し、UI で警告表示
  * 詳細表示は可能な範囲で行うが、更新系操作は拒否する

    * 理由: 既存の不整合データを悪化させないため
    * UI はエラー詳細ダイアログへ誘導する

### DD-LOAD-005 キャッシュ方針

* 一覧はキャッシュし、UI 操作を軽くする
* 課題詳細表示時は毎回ディスク再ロードする（最新化優先）
* 保存後は対象課題と一覧キャッシュを更新し整合を取る

---

## DD-PERSIST-001 永続化（アトミック更新）

### DD-PERSIST-002 更新手順

* 書き込み先と同じディレクトリに tmp を作る
* tmp に全内容を書き込む
* rename で本来のファイル名に置換

tmp 名

* `<issue_id>.json.tmp.<pid>.<timestamp>`

### DD-PERSIST-003 fsync

* 実施しない

### DD-PERSIST-004 tmp 残骸の扱い

* 起動時に `*.tmp.*` を検出
* 最終更新時刻（mtime）から経過時間を算出し、以下で統一する
  * 24時間未満: 削除する
    * 削除失敗は E_IO_WRITE としてエラー一覧に載せる
  * 24時間以上: 削除しない
    * エラー一覧に載せる（target_path、message、hint を含む）

---

## DD-UI-001 画面設計（Vue + Vuetify）

### DD-UI-002 画面一覧

* ProjectSelectDialog
* MainView（カテゴリ一覧 + 課題一覧）
* IssueDetailDialog
* ErrorDetailDialog
* ContractorPasswordDialog

### DD-UI-003 状態管理（推奨）

* `stores/app`（モード、プロジェクトパス、起動状態）
* `stores/categories`（カテゴリ一覧、選択中カテゴリ）
* `stores/issues`（一覧キャッシュ、フィルタ、ソート、ページ）
* `stores/errors`（ロードエラー一覧）

### DD-UI-004 ProjectSelectDialog

要素

* 前回の `<PROJECT_ROOT>` を初期値表示
* 参照ボタン（ディレクトリ選択）
* 新規作成ボタン（指定パスにフォルダ作成）
* 開くボタン（Validate → Save config → 次画面へ）
* キャンセル（終了）

Backend 呼び出し

* `GetAppBootstrap`
* `ValidateProjectRoot`
* `CreateProjectRoot`
* `SaveLastProjectRoot`

### DD-UI-005 MainView（カテゴリと課題一覧）

カテゴリ（左）

* フラットなリスト表示
* Contractor のみ: 追加、名称変更、削除
* 読み取り専用カテゴリ（is_read_only=true）は、一覧上で識別できる表示とし、当該カテゴリに対する編集操作は無効化する

課題一覧（右）

* 表示列（基本設計準拠）

  * 課題ID、タイトル、ステータス、優先度、作成元会社、更新日時、回答希望期限
  * `is_schema_invalid: true` の場合、警告表示を付与
* ソート

  * 列ヘッダで切替
* フィルタ

  * ステータス、優先度、期限
* ページング

  * 20件固定（将来設定化可）

行クリック

* IssueDetailDialog を開く

終了状態表示

* `Closed|Rejected` はグレーアウト

### DD-UI-006 IssueDetailDialog

表示

* タイトル、説明、期限、優先度、ステータス、作成元、作成/更新日時
* コメント一覧（古い順）
* Markdown 表示は markdown-it

編集

* 初期は閲覧モード
* 編集ボタンで編集モード
* 必須入力（タイトル、説明、期限、優先度）をフロントで検証
* 保存時に Backend の `UpdateIssue` を呼ぶ
* スキーマ不整合課題（`is_schema_invalid: true`）は更新操作を禁止し、エラー詳細ダイアログへ誘導する
* 選択カテゴリが読み取り専用（is_read_only=true）の場合、更新・コメント追加を禁止し、読み取り専用である旨を表示する

ステータス

* モード別に候補を絞る
* Backend でも拒否

コメント追加（添付含む）

* 本文（Markdown）、投稿者名、添付選択（任意）
* 送信で `AddComment` → 表示更新
* 本文サイズ 100KB（UTF-8 bytes）を超えたら送信不可
* コメント投稿後に添付を追加する機能は持たない

  * 追加したくなった場合は新たにコメント投稿する運用とする

### DD-UI-007 ErrorDetailDialog

表示項目

* 対象パス
* エラー種別（コード）
* メッセージ
* 詳細（折りたたみ）
* コピーボタン（パス、メッセージ）

起動導線

* メニューまたはヘッダのエラーアイコンから開く
* カテゴリ単位の絞り込み

### DD-UI-008 ContractorPasswordDialog

* パスワード入力
* 失敗時は「照合失敗」を表示し、ユーザが閉じたらアプリを終了

---

## DD-STORE-001 Pinia ストア設計

### DD-STORE-002 目的と責務

- UI の状態（選択中カテゴリ、一覧キャッシュ、ソート、フィルタ、ページ、詳細表示中の課題など）を画面間で共有する
- Backend 呼び出し（Wails binding）を actions に閉じ、画面は store の API を呼ぶだけにする
- 一覧はキャッシュし、課題詳細は毎回ディスク再ロードする
- スキーマ不整合の課題は一覧に警告付きで出すが、更新系は拒否しエラー詳細へ誘導する
- エラーは stores/errors に集約し、画面は errors を単一の参照点として表示する

### DD-STORE-003 ストア一覧

- stores/app
  - 起動、Project Root、モード（Contractor/Vendor）、共通設定
- stores/categories
  - カテゴリ一覧、選択中カテゴリ
- stores/issues
  - 課題一覧キャッシュ（カテゴリ別）、ソート、フィルタ、ページング
- stores/issueDetail
  - 選択中課題の詳細、編集状態、保存、コメント投稿
- stores/errors
  - エラーの集約と表示用データ

### DD-STORE-004 型定義（共通）

- Backend DTO を基本とし、UI 専用の状態（isDirty 等）は別フィールドで保持する

TypeScript 型（例）

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
    status: IssueStatus[];         // 空なら全て
    priority: IssuePriority[];     // 空なら全て
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

### DD-STORE-010 actions 定義（共通方針）

- すべての actions は Backend 呼び出しを actions に閉じる
- 失敗時は必ず stores/errors に登録する（DD-STORE-016）

### DD-STORE-011 stores/errors actions

- capture(e, ctx)
  - 概要: 例外やエラーを UiErrorItem に正規化して items に追加する
- captureApiError(apiError, ctx)
  - 概要: Backend の ApiErrorDTO を受け取り登録する
- captureMany(list, ctx)
  - 概要: ロード時に発生した複数エラー（破損一覧など）をまとめて登録する
- markRead(id), markAllRead()
  - 概要: 既読管理
- clearAll(), clearBySource(source)
  - 概要: エラー一覧のクリア
- loadFromBackend(scope)
  - 概要: Backend の GetLoadErrors(scope) を呼び、結果を captureMany で取り込む（必要に応じて使用）

### DD-STORE-012 stores/app actions

- bootstrap()
  - 概要: GetAppBootstrap を呼び、pageSize、last project root、auth/contractor.json 有無などを state に反映する
- selectProjectRoot(path)
  - 概要: ValidateProjectRoot → SaveLastProjectRoot を行い、projectRoot を更新する
- createProjectRoot(path)
  - 概要: CreateProjectRoot → SaveLastProjectRoot を行い、projectRoot を更新する
- detectMode()
  - 概要: DetectMode を呼び、contractorAuthRequired と mode を決定する
- verifyContractorPassword(password)
  - 概要: VerifyContractorPassword を呼び Contractor モードを確定する（失敗時は errors に登録し、UI 側で終了導線へ）

### DD-STORE-013 stores/categories actions

- loadCategories()
  - 概要: ListCategories を呼び、カテゴリ一覧を更新する
- selectCategory(name)
  - 概要: selectedCategory を更新し、stores/issues.loadIssues(name) を呼ぶ
- createCategory(name)
  - 概要: Contractor のみ。CreateCategory を呼び一覧を更新する
- renameCategory(oldName, newName)
  - 概要: Contractor のみ。RenameCategory を呼び selectedCategory と issuesByCategory のキー整合を更新する
- deleteCategory(name)
  - 概要: Contractor のみ。DeleteCategory を呼び、カテゴリ一覧と該当キャッシュを更新する

権限制御
- mode が Vendor の場合は実行せず、errors に E_PERMISSION 相当を登録する

### DD-STORE-014 stores/issues actions

- loadIssues(category, opts)
  - 概要: ListIssues を呼び、issuesByCategory[category] のキャッシュを更新する
  - 補足: Backend で opts を受けて sort/filter/page に合わせて加工・整形を行う。
- refreshIssues(category)
  - 概要: loadIssues を強制実行する
- setSort(category, sort)
  - 概要: queryByCategory を更新し、表示用の並び替えに反映する
- setFilter(category, filter)
  - 概要: queryByCategory を更新し、表示用の絞り込みに反映する
- setPage(category, page)
  - 概要: queryByCategory を更新し、ページングに反映する
- invalidateCategory(category)
  - 概要: 該当カテゴリの一覧キャッシュを破棄し、次回 loadIssues を必須にする
- applyIssueUpdatedToCache(issueDetail)
  - 概要: saveIssue/addComment 成功後に、一覧キャッシュの該当課題の summary を更新する

### DD-STORE-015 stores/issueDetail actions

- openIssue(category, issueId)
  - 概要: GetIssue を呼び、毎回ディスク再ロードした最新の IssueDetailDTO を current に設定する
- reloadCurrent()
  - 概要: current が存在する場合に GetIssue を再実行する
- saveIssue(update)
  - 概要: UpdateIssue を呼び current を更新する
  - 制約: current.is_schema_invalid が true の場合は保存を拒否し、errors に登録して終了する
- addComment(payload)
  - 概要: AddComment を呼び current を更新する
  - 制約: 添付は任意で、コメント投稿と同時のみ

### DD-STORE-016 エラー集約ルール（stores/errors に集約）

- すべての store の actions は、Backend 呼び出しを try/catch で囲み、失敗時に必ず stores/errors に登録する
- 画面へ出すエラー表示は以下に統一する
  - 一時的な通知（トースト等）を出す場合でも、登録を先に行う
  - ErrorDetailDialog は stores/errors.items を表示する
- 画面内バリデーション（必須未入力、文字数超過など）は errors へ登録しない
  - ただし Backend が返す E_VALIDATION は errors に登録する
- どの操作で失敗したか追跡できるよう、ctx に source と action を必ず入れる
  - 例: { source: "issues", action: "loadIssues", category }

### DD-STORE-017 追記規則

- 新規の state や action を追加する場合は、該当ストアの末尾に追記し、既存の名前と意味を変更しない
- 表示仕様の変更により state が増える場合でも、削除やリネームは避け、非推奨化して段階移行する

## DD-VALID-001 入力バリデーション仕様

### DD-VALID-002 課題

必須

* title, description, due_date, priority

上限（255 文字）

* title 最大 255 文字
* description 最大 255 文字

due_date

* `YYYY-MM-DD`
* ローカル日付、時刻なし

category 名（フォルダ名）

* 最大 255 文字
* Windows 禁止文字を含まない
* 末尾 `.` や空白不可

### DD-VALID-003 コメント

* body: UTF-8 bytes <= 100KB
* author_name: 必須、最大 255 文字
* 添付

  * 必須ではない
  * 添付ファイル名（サニタイズ後）最大 255 文字

---

## DD-LOG-001 ログ設計

### DD-LOG-002 出力先

* `logs/ratta.log`

### DD-LOG-003 ローテーション

* 1ファイル最大 1MB
* 最大 3 世代
* サイズ到達時にローテート
* 設定ファイル(config.json)のlog.leveldで出力を制御

### DD-LOG-004 記録内容

* 例外、I/O エラー、検証エラー
* 入力値は最小限（長文や機微情報をログへ残さない）
* JSONによる構造化ログを用いる

---

## DD-TEST-001 テスト設計

### DD-TEST-002 Go（unit）

対象

* ID 採番（nanoid, uuid v7）
* JSON ロード、スキーマ検証、破損分類（破損と不整合の分類）
* アトミック更新（tmp→rename）
* モード判定、権限制御（ステータス遷移）
* 添付ファイル保存（サニタイズ、相対パス生成、255 文字切り詰め）

### DD-TEST-003 Vue（unit）

対象

* 各ダイアログの必須入力、エラー表示
* ソート・フィルタ・ページング（一覧キャッシュ前提）
* Markdown 表示（最小限）
* スキーマ不整合課題に対する更新禁止の UI 動作

### DD-TEST-004 E2E（Playwright）

代表シナリオ

* 課題新規登録 → JSON 生成 → 一覧反映
* コメント追加（添付なし、添付あり）→ JSON 更新 → 詳細反映
* Vendor モードで `Closed|Rejected` 遷移不可
* スキーマ不整合 JSON を置いた状態で起動 → 一覧に警告付きで表示、エラー詳細にも表示、更新操作は不可
* 破損 JSON を置いた状態で起動 → 一覧に出さず、エラー詳細に表示
* 課題詳細は毎回ディスク再ロードされること（外部変更の反映確認）
* 一覧キャッシュが更新操作後に正しく更新されること
