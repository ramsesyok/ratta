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
	Version       int
	IssueID       string
	Category      string
	Title         string
	Description   string
	Status        Status
	Priority      Priority
	OriginCompany Company
	Assignee      string
	CreatedAt     string
	UpdatedAt     string
	DueDate       string
	Comments      []Comment
}

// Comment は DD-DATA-004 のコメントデータを表す。
type Comment struct {
	CommentID     string
	Body          string
	AuthorName    string
	AuthorCompany Company
	CreatedAt     string
	Attachments   []AttachmentRef
}

// AttachmentRef は DD-DATA-005 の添付参照を表す。
type AttachmentRef struct {
	AttachmentID string
	FileName     string
	StoredName   string
	RelativePath string
	MimeType     string
	SizeBytes    int64
}
