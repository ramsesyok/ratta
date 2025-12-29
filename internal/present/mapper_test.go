// mapper_test.go は DTO 変換処理のテストを行い、UI統合は扱わない。
package present

import (
	"testing"

	"ratta/internal/app/categoryscan"
	"ratta/internal/app/issueops"
	"ratta/internal/domain/issue"
)

func TestToCategoryDTO_MapsFields(t *testing.T) {
	// カテゴリ情報がDTOへ正しく写像されることを確認する。
	input := categoryscan.Category{
		Name:       "Cat",
		IsReadOnly: true,
		Path:       "C:/project/Cat",
	}

	dto := ToCategoryDTO(input)

	if dto.Name != "Cat" {
		t.Fatalf("unexpected name: %s", dto.Name)
	}
	if !dto.IsReadOnly {
		t.Fatal("expected read-only to be true")
	}
	if dto.Path != "C:/project/Cat" {
		t.Fatalf("unexpected path: %s", dto.Path)
	}
	if dto.IssueCount != 0 {
		t.Fatalf("unexpected issue count: %d", dto.IssueCount)
	}
}

func TestToIssueDetailDTO_MapsNested(t *testing.T) {
	// コメントと添付を含む課題が詳細DTOへ変換されることを確認する。
	detail := issueops.IssueDetail{
		IsSchemaInvalid: false,
		Issue: issue.Issue{
			Version:       1,
			IssueID:       "ABC123DEF",
			Category:      "Cat",
			Title:         "Title",
			Description:   "Desc",
			Status:        issue.StatusOpen,
			Priority:      issue.PriorityHigh,
			OriginCompany: issue.CompanyVendor,
			Assignee:      "assignee",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-02T00:00:00Z",
			DueDate:       "2024-01-03",
			Comments: []issue.Comment{
				{
					CommentID:     "comment-1",
					Body:          "body",
					AuthorName:    "author",
					AuthorCompany: issue.CompanyContractor,
					CreatedAt:     "2024-01-01T12:00:00Z",
					Attachments: []issue.AttachmentRef{
						{
							AttachmentID: "att-1",
							FileName:     "file.txt",
							StoredName:   "att-1_file.txt",
							RelativePath: "ABC123DEF.files/att-1_file.txt",
							MimeType:     "text/plain",
							SizeBytes:    42,
						},
					},
				},
			},
		},
	}

	dto := ToIssueDetailDTO(detail)

	if dto.IssueID != "ABC123DEF" {
		t.Fatalf("unexpected issue id: %s", dto.IssueID)
	}
	if len(dto.Comments) != 1 {
		t.Fatalf("unexpected comment count: %d", len(dto.Comments))
	}
	if len(dto.Comments[0].Attachments) != 1 {
		t.Fatalf("unexpected attachment count: %d", len(dto.Comments[0].Attachments))
	}
	if dto.Comments[0].Attachments[0].StoredName != "att-1_file.txt" {
		t.Fatalf("unexpected stored name: %s", dto.Comments[0].Attachments[0].StoredName)
	}
}

func TestToIssueSummaryDTO_MapsFields(t *testing.T) {
	// 一覧要約が DTO へ正しく写像されることを確認する。
	summary := issueops.IssueSummary{
		IssueID:         "ABC123DEF",
		Title:           "Title",
		Status:          "Open",
		Priority:        "High",
		OriginCompany:   "Vendor",
		UpdatedAt:       "2024-01-02T00:00:00Z",
		DueDate:         "2024-01-03",
		IsSchemaInvalid: true,
	}

	dto := ToIssueSummaryDTO(summary)

	if dto.IssueID != "ABC123DEF" {
		t.Fatalf("unexpected issue id: %s", dto.IssueID)
	}
	if !dto.IsSchemaInvalid {
		t.Fatal("expected schema invalid to be true")
	}
}
