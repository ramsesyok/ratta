// app.go は Wails のアプリケーション層ブリッジを担い、UI からの操作をユースケースに接続する。
// UI 表示やドメイン詳細はここで扱わない。
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"ratta/internal/app/categoryops"
	"ratta/internal/app/categoryscan"
	"ratta/internal/app/issueops"
	"ratta/internal/app/modedetect"
	"ratta/internal/app/projectroot"
	"ratta/internal/domain/issue"
	"ratta/internal/infra/configrepo"
	"ratta/internal/infra/schema"
	"ratta/internal/present"

	mod "ratta/internal/domain/mode"
)

// App は DD-BE-002 の Wails バインド対象を表す。
type App struct {
	ctx     context.Context
	exePath string
	mode    mod.Mode
	root    string

	configRepo *configrepo.Repository
	validator  *schema.Validator
}

// NewApp は DD-BE-002 の初期化を行う。
// 目的: Wails 起動時に必要な状態を初期化する。
// 入力: なし。
// 出力: 初期化済み App。
// エラー: 返却値で表現しない。実行ファイルパスや設定読み込み失敗時は空文字のまま保持する。
// 副作用: config.json を読み取る。
// 並行性: 呼び出し側が単一スレッドで実行する前提。
// 不変条件: mode は Vendor を初期値とし、root は設定があれば復元する。
// 関連DD: DD-BE-002
func NewApp() *App {
	exePath, exeErr := os.Executable()
	if exeErr != nil {
		exePath = ""
	}
	configRepo := configrepo.NewRepository(exePath)
	root := ""
	if cfg, hasConfig, err := configRepo.Load(); err == nil && hasConfig {
		if cfg.LastProjectRootPath != "" {
			root = cfg.LastProjectRootPath
		}
	}
	validator := loadValidator(exePath)
	return &App{
		exePath:    exePath,
		mode:       mod.ModeVendor,
		root:       root,
		configRepo: configRepo,
		validator:  validator,
	}
}

// startup は起動時に context を保存する。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetAppBootstrap は DD-BE-003 の起動時情報を返す。
// 目的: UI 初期表示に必要な設定値と状態を返す。
// 入力: なし。
// 出力: BootstrapDTO を含む Response。
// エラー: 設定読み込みに失敗した場合はデフォルト設定で続行する。
// 副作用: 設定リポジトリから読み取りを行う。
// 並行性: App はスレッドセーフではないため同時呼び出しは想定しない。
// 不変条件: 返却する DTO は nil の代わりに空値を使う。
// 関連DD: DD-BE-003
func (a *App) GetAppBootstrap() present.Response {
	cfg, hasConfig, err := a.configRepo.Load()
	if err != nil {
		cfg = configrepo.DefaultConfig()
		hasConfig = false
	}

	var lastPath *string
	if cfg.LastProjectRootPath != "" {
		value := cfg.LastProjectRootPath
		lastPath = &value
	}

	hasAuth := false
	if a.exePath != "" {
		if _, statErr := os.Stat(filepath.Join(filepath.Dir(a.exePath), "auth", "contractor.json")); statErr == nil {
			hasAuth = true
		}
	}

	dto := present.BootstrapDTO{
		HasConfig:             hasConfig,
		LastProjectRootPath:   lastPath,
		UIPageSize:            cfg.UI.PageSize,
		LogLevel:              cfg.Log.Level,
		HasContractorAuthFile: hasAuth,
	}
	return present.Ok(dto)
}

// ValidateProjectRoot は DD-BE-003 の Project Root 検証を行う。
func (a *App) ValidateProjectRoot(path string) present.Response {
	service := projectroot.NewService(a.configRepo)
	result, err := service.ValidateProjectRoot(path)
	if err != nil {
		return present.Fail(err)
	}
	dto := present.ValidationResultDTO{
		IsValid:        result.IsValid,
		NormalizedPath: result.NormalizedPath,
		Message:        result.Message,
	}
	if result.Details != "" {
		value := result.Details
		dto.Details = &value
	}
	return present.Ok(dto)
}

// CreateProjectRoot は DD-BE-003 の Project Root 作成を行う。
func (a *App) CreateProjectRoot(path string) present.Response {
	service := projectroot.NewService(a.configRepo)
	result, err := service.CreateProjectRoot(path)
	if err != nil {
		return present.Fail(err)
	}
	dto := present.ValidationResultDTO{
		IsValid:        result.IsValid,
		NormalizedPath: result.NormalizedPath,
		Message:        result.Message,
	}
	if result.Details != "" {
		value := result.Details
		dto.Details = &value
	}
	return present.Ok(dto)
}

// SaveLastProjectRoot は DD-BE-003 の last_project_root_path 更新を行う。
func (a *App) SaveLastProjectRoot(path string) present.Response {
	service := projectroot.NewService(a.configRepo)
	if err := service.SaveLastProjectRoot(path); err != nil {
		return present.Fail(err)
	}
	a.root = path
	return present.Ok(nil)
}

// DetectMode は DD-BE-003 のモード判定を行う。
func (a *App) DetectMode() present.Response {
	service := modedetect.NewService(a.exePath, a.validator)
	modeValue, requiresPassword, err := service.DetectMode()
	if err != nil {
		return present.Fail(err)
	}
	dto := present.ModeDTO{Mode: string(modeValue), RequiresPassword: requiresPassword}
	return present.Ok(dto)
}

// VerifyContractorPassword は DD-BE-003 のパスワード検証を行う。
func (a *App) VerifyContractorPassword(password string) present.Response {
	service := modedetect.NewService(a.exePath, a.validator)
	modeValue, err := service.VerifyContractorPassword(password)
	if err != nil {
		return present.Fail(err)
	}
	a.mode = modeValue
	dto := present.ModeDTO{Mode: string(modeValue), RequiresPassword: false}
	return present.Ok(dto)
}

// ListCategories は DD-LOAD-002 のカテゴリ一覧を返す。
func (a *App) ListCategories() present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	result, err := categoryscan.Scan(a.root)
	if err != nil {
		return present.Fail(err)
	}
	categories := make([]present.CategoryDTO, 0, len(result.Categories))
	for _, category := range result.Categories {
		categories = append(categories, present.ToCategoryDTO(category))
	}
	dto := present.CategoryListDTO{
		Categories: categories,
		Errors:     result.ErrorCount,
	}
	return present.Ok(dto)
}

// CreateCategory は DD-BE-003 のカテゴリ作成を行う。
func (a *App) CreateCategory(name string) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := categoryops.NewService(a.root)
	category, err := service.CreateCategory(name, a.mode)
	if err != nil {
		return present.Fail(err)
	}
	dto := present.CategoryDTO{
		Name:       category.Name,
		IsReadOnly: category.IsReadOnly,
		Path:       category.Path,
		IssueCount: 0,
	}
	return present.Ok(dto)
}

// RenameCategory は DD-BE-003 のカテゴリ名変更を行う。
func (a *App) RenameCategory(oldName, newName string) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := categoryops.NewService(a.root)
	category, err := service.RenameCategory(oldName, newName, a.mode)
	if err != nil {
		return present.Fail(err)
	}
	dto := present.CategoryDTO{
		Name:       category.Name,
		IsReadOnly: category.IsReadOnly,
		Path:       category.Path,
		IssueCount: 0,
	}
	return present.Ok(dto)
}

// DeleteCategory は DD-BE-003 のカテゴリ削除を行う。
func (a *App) DeleteCategory(name string) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := categoryops.NewService(a.root)
	if err := service.DeleteCategory(name, a.mode); err != nil {
		return present.Fail(err)
	}
	return present.Ok(nil)
}

// ListIssues は DD-BE-003 の課題一覧を返す。
func (a *App) ListIssues(category string, query present.IssueListQueryDTO) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := issueops.NewService(a.root, a.validator)
	result, err := service.ListIssues(category, issueops.IssueListQuery{
		Page:      query.Page,
		PageSize:  query.PageSize,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
	})
	if err != nil {
		return present.Fail(err)
	}
	items := make([]present.IssueSummaryDTO, 0, len(result.Issues))
	for _, item := range result.Issues {
		items = append(items, present.ToIssueSummaryDTO(item))
	}
	dto := present.IssueListDTO{
		Category: result.Category,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Issues:   items,
	}
	return present.Ok(dto)
}

// GetIssue は DD-BE-003 の課題詳細を取得する。
func (a *App) GetIssue(category, issueID string) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := issueops.NewService(a.root, a.validator)
	detail, err := service.GetIssue(category, issueID)
	if err != nil {
		return present.Fail(err)
	}
	return present.Ok(present.ToIssueDetailDTO(detail))
}

// CreateIssue は DD-BE-003 の課題作成を行う。
func (a *App) CreateIssue(category string, dto present.IssueCreateDTO) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := issueops.NewService(a.root, a.validator)
	detail, err := service.CreateIssue(category, a.mode, issueops.IssueCreateInput{
		Title:       dto.Title,
		Description: dto.Description,
		DueDate:     dto.DueDate,
		Priority:    issue.Priority(dto.Priority),
		Assignee:    dto.Assignee,
	})
	if err != nil {
		return present.Fail(err)
	}
	return present.Ok(present.ToIssueDetailDTO(detail))
}

// UpdateIssue は DD-BE-003 の課題更新を行う。
func (a *App) UpdateIssue(category, issueID string, dto present.IssueUpdateDTO) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := issueops.NewService(a.root, a.validator)
	detail, err := service.UpdateIssue(category, issueID, a.mode, issueops.IssueUpdateInput{
		Title:       dto.Title,
		Description: dto.Description,
		DueDate:     dto.DueDate,
		Priority:    issue.Priority(dto.Priority),
		Status:      issue.Status(dto.Status),
		Assignee:    dto.Assignee,
	})
	if err != nil {
		return present.Fail(err)
	}
	return present.Ok(present.ToIssueDetailDTO(detail))
}

// AddComment は DD-BE-003 のコメント追加を行う。
func (a *App) AddComment(category, issueID string, dto present.CommentCreateDTO) present.Response {
	if a.root == "" {
		return present.Fail(errors.New("project root is not set"))
	}
	service := issueops.NewService(a.root, a.validator)
	attachments := make([]issueops.CommentAttachmentInput, 0, len(dto.Attachments))
	for _, attachment := range dto.Attachments {
		data, err := os.ReadFile(attachment.SourcePath)
		if err != nil {
			return present.Fail(err)
		}
		original := attachment.OriginalFileName
		if original == "" {
			original = filepath.Base(attachment.SourcePath)
		}
		attachments = append(attachments, issueops.CommentAttachmentInput{
			OriginalName: original,
			Data:         data,
			MimeType:     attachment.MimeType,
		})
	}
	detail, err := service.AddComment(category, issueID, a.mode, issueops.CommentCreateInput{
		Body:        dto.Body,
		AuthorName:  dto.AuthorName,
		Attachments: attachments,
	})
	if err != nil {
		return present.Fail(err)
	}
	return present.Ok(present.ToIssueDetailDTO(detail))
}

func loadValidator(exePath string) *schema.Validator {
	if exePath != "" {
		dir := filepath.Join(filepath.Dir(exePath), "schemas")
		if validator, err := schema.NewValidatorFromDir(dir); err == nil {
			return validator
		}
	}
	if validator, err := schema.NewValidatorFromDir("schemas"); err == nil {
		return validator
	}
	return nil
}
