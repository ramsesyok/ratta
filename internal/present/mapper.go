package present

import (
	"ratta/internal/app/categoryscan"
	"ratta/internal/app/issueops"
	"ratta/internal/domain/issue"
)

// ToCategoryDTO は DD-BE-003 のカテゴリ DTO に変換する。
func ToCategoryDTO(category categoryscan.Category) CategoryDTO {
	return CategoryDTO{
		Name:       category.Name,
		IsReadOnly: category.IsReadOnly,
		Path:       category.Path,
		IssueCount: 0,
	}
}

// ToIssueDetailDTO は DD-DATA-003/004 の課題詳細 DTO に変換する。
func ToIssueDetailDTO(detail issueops.IssueDetail) IssueDetailDTO {
	issueValue := detail.Issue
	return IssueDetailDTO{
		IsSchemaInvalid: detail.IsSchemaInvalid,
		Version:         issueValue.Version,
		IssueID:         issueValue.IssueID,
		Category:        issueValue.Category,
		Title:           issueValue.Title,
		Description:     issueValue.Description,
		Status:          string(issueValue.Status),
		Priority:        string(issueValue.Priority),
		OriginCompany:   string(issueValue.OriginCompany),
		Assignee:        issueValue.Assignee,
		CreatedAt:       issueValue.CreatedAt,
		UpdatedAt:       issueValue.UpdatedAt,
		DueDate:         issueValue.DueDate,
		Comments:        toCommentDTOs(issueValue.Comments),
	}
}

// ToIssueSummaryDTO は DD-LOAD-004 の課題一覧 DTO に変換する。
func ToIssueSummaryDTO(summary issueops.IssueSummary) IssueSummaryDTO {
	return IssueSummaryDTO{
		IssueID:         summary.IssueID,
		Title:           summary.Title,
		Status:          summary.Status,
		Priority:        summary.Priority,
		OriginCompany:   summary.OriginCompany,
		UpdatedAt:       summary.UpdatedAt,
		DueDate:         summary.DueDate,
		IsSchemaInvalid: summary.IsSchemaInvalid,
	}
}

func toCommentDTOs(comments []issue.Comment) []CommentDTO {
	if len(comments) == 0 {
		return []CommentDTO{}
	}
	dtos := make([]CommentDTO, 0, len(comments))
	for _, comment := range comments {
		dtos = append(dtos, CommentDTO{
			CommentID:     comment.CommentID,
			Body:          comment.Body,
			AuthorName:    comment.AuthorName,
			AuthorCompany: string(comment.AuthorCompany),
			CreatedAt:     comment.CreatedAt,
			Attachments:   toAttachmentDTOs(comment.Attachments),
		})
	}
	return dtos
}

func toAttachmentDTOs(attachments []issue.AttachmentRef) []AttachmentRefDTO {
	if len(attachments) == 0 {
		return []AttachmentRefDTO{}
	}
	dtos := make([]AttachmentRefDTO, 0, len(attachments))
	for _, attachment := range attachments {
		dtos = append(dtos, AttachmentRefDTO{
			AttachmentID: attachment.AttachmentID,
			FileName:     attachment.FileName,
			StoredName:   attachment.StoredName,
			RelativePath: attachment.RelativePath,
			MimeType:     attachment.MimeType,
			SizeBytes:    attachment.SizeBytes,
		})
	}
	return dtos
}
