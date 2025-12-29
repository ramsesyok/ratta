package issue

// Status は DD-DATA-003 のステータス種別を表す。
type Status string

const (
	StatusOpen     Status = "Open"
	StatusWorking  Status = "Working"
	StatusInquiry  Status = "Inquiry"
	StatusHold     Status = "Hold"
	StatusFeedback Status = "Feedback"
	StatusResolved Status = "Resolved"
	StatusClosed   Status = "Closed"
	StatusRejected Status = "Rejected"
)

// IsValid は DD-DATA-003 のステータス一覧に含まれるかを判定する。
func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusWorking, StatusInquiry, StatusHold, StatusFeedback, StatusResolved, StatusClosed, StatusRejected:
		return true
	default:
		return false
	}
}

// IsEndState は DD-DATA-003 の終状態かどうかを判定する。
func (s Status) IsEndState() bool {
	return s == StatusClosed || s == StatusRejected
}

// Priority は DD-DATA-003 の優先度種別を表す。
type Priority string

const (
	PriorityHigh   Priority = "High"
	PriorityMedium Priority = "Medium"
	PriorityLow    Priority = "Low"
)

// IsValid は DD-DATA-003 の優先度一覧に含まれるかを判定する。
func (p Priority) IsValid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}

// Company は DD-DATA-003/DD-DATA-004 の会社種別を表す。
type Company string

const (
	CompanyContractor Company = "Contractor"
	CompanyVendor     Company = "Vendor"
)

// IsValid は DD-DATA-003/DD-DATA-004 の会社種別に含まれるかを判定する。
func (c Company) IsValid() bool {
	return c == CompanyContractor || c == CompanyVendor
}

// Issue は DD-DATA-003 の課題データを表す。
type Issue struct {
	Version       int       `json:"version"`
	IssueID       string    `json:"issue_id"`
	Category      string    `json:"category"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        Status    `json:"status"`
	Priority      Priority  `json:"priority"`
	OriginCompany Company   `json:"origin_company"`
	Assignee      string    `json:"assignee,omitempty"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	DueDate       string    `json:"due_date"`
	Comments      []Comment `json:"comments"`
}

// Comment は DD-DATA-004 のコメントデータを表す。
type Comment struct {
	CommentID     string          `json:"comment_id"`
	Body          string          `json:"body"`
	AuthorName    string          `json:"author_name"`
	AuthorCompany Company         `json:"author_company"`
	CreatedAt     string          `json:"created_at"`
	Attachments   []AttachmentRef `json:"attachments"`
}

// AttachmentRef は DD-DATA-005 の添付参照を表す。
type AttachmentRef struct {
	AttachmentID string `json:"attachment_id"`
	FileName     string `json:"file_name"`
	StoredName   string `json:"stored_name"`
	RelativePath string `json:"relative_path"`
	MimeType     string `json:"mime_type,omitempty"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
}
