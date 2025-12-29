package present

// Response は DD-BE-003 の標準レスポンス形式を表す。
type Response struct {
	Ok    bool         `json:"ok"`
	Data  any          `json:"data,omitempty"`
	Error *ApiErrorDTO `json:"error,omitempty"`
}

// ApiErrorDTO は DD-BE-003 の共通エラーを表す。
type ApiErrorDTO struct {
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	TargetPath string `json:"target_path,omitempty"`
	Hint       string `json:"hint,omitempty"`
}

// BootstrapDTO は DD-BE-003 の起動時情報を表す。
type BootstrapDTO struct {
	HasConfig             bool    `json:"has_config"`
	LastProjectRootPath   *string `json:"last_project_root_path"`
	UIPageSize            int     `json:"ui_page_size"`
	LogLevel              string  `json:"log_level"`
	HasContractorAuthFile bool    `json:"has_contractor_auth_file"`
}

// ValidationResultDTO は DD-BE-003 の検証結果を表す。
type ValidationResultDTO struct {
	IsValid        bool    `json:"is_valid"`
	NormalizedPath string  `json:"normalized_path,omitempty"`
	Message        string  `json:"message"`
	Details        *string `json:"details,omitempty"`
}

// ModeDTO は DD-BE-003 のモード情報を表す。
type ModeDTO struct {
	Mode             string `json:"mode"`
	RequiresPassword bool   `json:"requires_password"`
}

// CategoryDTO は DD-BE-003 のカテゴリ情報を表す。
type CategoryDTO struct {
	Name       string `json:"name"`
	IsReadOnly bool   `json:"is_read_only"`
	Path       string `json:"path"`
	IssueCount int    `json:"issue_count"`
}

// CategoryListDTO は DD-BE-003 のカテゴリ一覧を表す。
type CategoryListDTO struct {
	Categories []CategoryDTO `json:"categories"`
	Errors     int           `json:"errors"`
}

// IssueSummaryDTO は DD-LOAD-004 の課題一覧項目を表す。
type IssueSummaryDTO struct {
	IssueID         string `json:"issue_id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	Priority        string `json:"priority"`
	OriginCompany   string `json:"origin_company"`
	UpdatedAt       string `json:"updated_at"`
	DueDate         string `json:"due_date"`
	IsSchemaInvalid bool   `json:"is_schema_invalid"`
}

// IssueListDTO は DD-BE-003 の課題一覧結果を表す。
type IssueListDTO struct {
	Category string            `json:"category"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Issues   []IssueSummaryDTO `json:"issues"`
}

// IssueListQueryDTO は DD-BE-003 の一覧条件を表す。
type IssueListQueryDTO struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// IssueCreateDTO は DD-BE-003 の課題作成入力を表す。
type IssueCreateDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Priority    string `json:"priority"`
	Assignee    string `json:"assignee"`
}

// IssueUpdateDTO は DD-BE-003 の課題更新入力を表す。
type IssueUpdateDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
}

// AttachmentUploadDTO は DD-DATA-005 の添付入力を表す。
type AttachmentUploadDTO struct {
	SourcePath       string `json:"source_path"`
	OriginalFileName string `json:"original_file_name"`
	MimeType         string `json:"mime_type"`
}

// CommentCreateDTO は DD-DATA-004 のコメント作成入力を表す。
type CommentCreateDTO struct {
	Body        string                `json:"body"`
	AuthorName  string                `json:"author_name"`
	Attachments []AttachmentUploadDTO `json:"attachments"`
}

// AttachmentRefDTO は DD-DATA-005 の添付参照を表す。
type AttachmentRefDTO struct {
	AttachmentID string `json:"attachment_id"`
	FileName     string `json:"file_name"`
	StoredName   string `json:"stored_name"`
	RelativePath string `json:"relative_path"`
	MimeType     string `json:"mime_type,omitempty"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
}

// CommentDTO は DD-DATA-004 のコメント情報を表す。
type CommentDTO struct {
	CommentID     string             `json:"comment_id"`
	Body          string             `json:"body"`
	AuthorName    string             `json:"author_name"`
	AuthorCompany string             `json:"author_company"`
	CreatedAt     string             `json:"created_at"`
	Attachments   []AttachmentRefDTO `json:"attachments"`
}

// IssueDetailDTO は DD-DATA-003/004 の課題詳細を表す。
type IssueDetailDTO struct {
	IsSchemaInvalid bool         `json:"is_schema_invalid"`
	Version         int          `json:"version"`
	IssueID         string       `json:"issue_id"`
	Category        string       `json:"category"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	Priority        string       `json:"priority"`
	OriginCompany   string       `json:"origin_company"`
	Assignee        string       `json:"assignee"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
	DueDate         string       `json:"due_date"`
	Comments        []CommentDTO `json:"comments"`
}
