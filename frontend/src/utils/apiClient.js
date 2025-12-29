import * as App from '../../wailsjs/go/main/App.js'

// ApiError は DD-STORE-004 の ApiErrorDTO を保持するエラーを表す。
// 目的: フロントエンド側で統一されたエラー情報を扱う。
// 入力: message は表示用メッセージ、payload は詳細情報。
// 出力: ApiError インスタンス。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: payload の各フィールドをそのまま保持する。
// 関連DD: DD-STORE-004
export class ApiError extends Error {
  constructor(message, payload) {
    super(message)
    this.name = 'ApiError'
    this.errorCode = payload?.error_code ?? 'E_INTERNAL'
    this.detail = payload?.detail ?? ''
    this.targetPath = payload?.target_path ?? ''
    this.hint = payload?.hint ?? ''
    this.action = payload?.action ?? ''
  }
}

// unwrapResponse は DD-BE-003 の ResponseDTO を正規化し、成功時は data を返す。
// 目的: ok/data/error の分岐を統一し、ストアで扱いやすくする。
// 入力: response はバックエンドのレスポンス、action は呼び出し名。
// 出力: response.data。
// エラー: ok=false またはレスポンス不正時に ApiError を送出する。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export function unwrapResponse(response, action) {
  if (!response || typeof response.ok !== 'boolean') {
    throw new ApiError('backend response is invalid', { action })
  }
  if (!response.ok) {
    const payload = response.error ?? { error_code: 'E_INTERNAL', message: 'backend error' }
    payload.action = action
    throw new ApiError(payload.message ?? 'backend error', payload)
  }
  return response.data
}

// getAppBootstrap は DD-BE-003 の起動時情報取得を行う。
// 目的: 起動時に必要な設定情報を取得する。
// 入力: なし。
// 出力: BootstrapDTO。
// エラー: 取得失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function getAppBootstrap() {
  const response = await App.GetAppBootstrap()
  return unwrapResponse(response, 'GetAppBootstrap')
}

// validateProjectRoot は DD-BE-003 の Project Root 検証を行う。
// 目的: プロジェクトルートの妥当性を検証する。
// 入力: path は対象パス。
// 出力: ValidationResultDTO。
// エラー: 検証失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function validateProjectRoot(path) {
  const response = await App.ValidateProjectRoot(path)
  return unwrapResponse(response, 'ValidateProjectRoot')
}

// createProjectRoot は DD-BE-003 の Project Root 作成を行う。
// 目的: プロジェクトルートを作成する。
// 入力: path は作成対象パス。
// 出力: ValidationResultDTO。
// エラー: 作成失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function createProjectRoot(path) {
  const response = await App.CreateProjectRoot(path)
  return unwrapResponse(response, 'CreateProjectRoot')
}

// saveLastProjectRoot は DD-BE-003 の設定更新を行う。
// 目的: 最終プロジェクトルートを保存する。
// 入力: path は保存対象パス。
// 出力: なし（data は未使用）。
// エラー: 保存失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ成功とみなす。
// 関連DD: DD-BE-003
export async function saveLastProjectRoot(path) {
  const response = await App.SaveLastProjectRoot(path)
  return unwrapResponse(response, 'SaveLastProjectRoot')
}

// detectMode は DD-BE-003 の起動時モード判定を行う。
// 目的: Vendor/Contractor モードとパスワード要求有無を取得する。
// 入力: なし。
// 出力: ModeDTO。
// エラー: 判定失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function detectMode() {
  const response = await App.DetectMode()
  return unwrapResponse(response, 'DetectMode')
}

// verifyContractorPassword は DD-BE-003/DD-CLI-005 のパスワード検証を行う。
// 目的: Contractor パスワードの検証結果を取得する。
// 入力: password は入力パスワード。
// 出力: ModeDTO。
// エラー: 検証失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003, DD-CLI-005
export async function verifyContractorPassword(password) {
  const response = await App.VerifyContractorPassword(password)
  return unwrapResponse(response, 'VerifyContractorPassword')
}

// listCategories は DD-BE-003 のカテゴリ一覧取得を行う。
// 目的: カテゴリ一覧を取得する。
// 入力: なし。
// 出力: CategoryListDTO。
// エラー: 取得失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function listCategories() {
  const response = await App.ListCategories()
  return unwrapResponse(response, 'ListCategories')
}

// createCategory は DD-BE-003 のカテゴリ作成を行う。
// 目的: 新規カテゴリを作成する。
// 入力: name はカテゴリ名。
// 出力: CategoryDTO。
// エラー: 作成失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function createCategory(name) {
  const response = await App.CreateCategory(name)
  return unwrapResponse(response, 'CreateCategory')
}

// renameCategory は DD-BE-003 のカテゴリ名変更を行う。
// 目的: 既存カテゴリの名称を変更する。
// 入力: oldName は旧名、newName は新名。
// 出力: CategoryDTO。
// エラー: 変更失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function renameCategory(oldName, newName) {
  const response = await App.RenameCategory(oldName, newName)
  return unwrapResponse(response, 'RenameCategory')
}

// deleteCategory は DD-BE-003 のカテゴリ削除を行う。
// 目的: カテゴリを削除する。
// 入力: name はカテゴリ名。
// 出力: なし（data は未使用）。
// エラー: 削除失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ成功とみなす。
// 関連DD: DD-BE-003
export async function deleteCategory(name) {
  const response = await App.DeleteCategory(name)
  return unwrapResponse(response, 'DeleteCategory')
}

// listIssues は DD-BE-003 の課題一覧取得を行う。
// 目的: 課題一覧を取得する。
// 入力: category はカテゴリ名、query は一覧条件。
// 出力: IssueListDTO。
// エラー: 取得失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function listIssues(category, query) {
  const response = await App.ListIssues(category, query)
  return unwrapResponse(response, 'ListIssues')
}

// getIssue は DD-BE-003 の課題取得を行う。
// 目的: 課題詳細を取得する。
// 入力: category はカテゴリ名、issueId は課題ID。
// 出力: IssueDetailDTO。
// エラー: 取得失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function getIssue(category, issueId) {
  const response = await App.GetIssue(category, issueId)
  return unwrapResponse(response, 'GetIssue')
}

// createIssue は DD-BE-003 の課題作成を行う。
// 目的: 新規課題を作成する。
// 入力: category はカテゴリ名、input は課題作成DTO。
// 出力: IssueDetailDTO。
// エラー: 作成失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function createIssue(category, input) {
  const response = await App.CreateIssue(category, input)
  return unwrapResponse(response, 'CreateIssue')
}

// updateIssue は DD-BE-003 の課題更新を行う。
// 目的: 既存課題を更新する。
// 入力: category はカテゴリ名、issueId は課題ID、input は更新DTO。
// 出力: IssueDetailDTO。
// エラー: 更新失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function updateIssue(category, issueId, input) {
  const response = await App.UpdateIssue(category, issueId, input)
  return unwrapResponse(response, 'UpdateIssue')
}

// addComment は DD-BE-003 のコメント追加を行う。
// 目的: 課題にコメントを追加する。
// 入力: category はカテゴリ名、issueId は課題ID、input はコメント作成DTO。
// 出力: IssueDetailDTO。
// エラー: 追加失敗時に ApiError を送出する。
// 副作用: バックエンド呼び出しを行う。
// 並行性: スレッドセーフ。
// 不変条件: ok=true の場合のみ data を返す。
// 関連DD: DD-BE-003
export async function addComment(category, issueId, input) {
  const response = await App.AddComment(category, issueId, input)
  return unwrapResponse(response, 'AddComment')
}
