// errors.js はエラーの集約と表示用データの生成を担い、UIの描画は扱わない。
// バックエンドからのレスポンス正規化は utils に委ねる。
import { defineStore } from 'pinia'

import { ApiError } from '../utils/apiClient'

// useErrorsStore は DD-STORE-009/011 のエラー集約ストアを提供する。
// 目的: エラー情報をUiErrorItemとして蓄積し、読み取り状態を管理する。
// 入力: Pinia の内部状態。
// 出力: errors ストア。
// エラー: なし。
// 副作用: なし。
// 並行性: Pinia のリアクティブ更新に従う。
// 不変条件: items は新しい順に追加する。
// 関連DD: DD-STORE-009, DD-STORE-011
export const useErrorsStore = defineStore('errors', {
  state: () => ({
    items: []
  }),
  actions: {
    // capture は未知の例外を UiErrorItem に変換して追加する。
    // 目的: 例外を統一形式に変換する。
    // 入力: e は例外、ctx は付帯情報。
    // 出力: 追加した UiErrorItem。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: action/source は必ず設定される。
    // 関連DD: DD-STORE-011, DD-STORE-016
    capture(e, ctx = {}) {
      if (e instanceof ApiError) {
        return this.captureApiError(e, ctx)
      }
      const item = buildErrorItem({
        source: ctx.source ?? 'backend',
        action: ctx.action ?? 'unknown',
        category: ctx.category,
        issue_id: ctx.issue_id,
        severity: 'error',
        user_message: e?.message ?? 'unexpected error',
        raw: e
      })
      this.items.unshift(item)
      return item
    },
    // captureApiError は ApiError を UiErrorItem に変換して追加する。
    // 目的: バックエンド由来のエラーを一貫した形式で保存する。
    // 入力: apiError は ApiError、ctx は付帯情報。
    // 出力: 追加した UiErrorItem。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: api が設定される。
    // 関連DD: DD-STORE-011, DD-STORE-016
    captureApiError(apiError, ctx = {}) {
      const api = {
        error_code: apiError.errorCode,
        message: apiError.message,
        detail: apiError.detail,
        target_path: apiError.targetPath,
        hint: apiError.hint
      }
      const item = buildErrorItem({
        source: ctx.source ?? 'backend',
        action: apiError.action ?? ctx.action ?? 'unknown',
        category: ctx.category,
        issue_id: ctx.issue_id,
        severity: severityFromCode(api.error_code),
        user_message: api.message ?? 'backend error',
        api,
        raw: apiError
      })
      this.items.unshift(item)
      return item
    },
    // captureMany は複数のエラーをまとめて追加する。
    // 目的: 複数エラーを一括で取り込む。
    // 入力: list はエラー配列、ctx は付帯情報。
    // 出力: 追加した件数。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: 追加順はリスト順。
    // 関連DD: DD-STORE-011
    captureMany(list, ctx = {}) {
      if (!Array.isArray(list)) {
        return 0
      }
      list.forEach((entry) => {
        this.capture(entry, ctx)
      })
      return list.length
    },
    // markRead は指定IDのエラーを既読にする。
    // 目的: 読み取り状態を更新する。
    // 入力: id は対象ID。
    // 出力: なし。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: ID が一致するもののみ更新する。
    // 関連DD: DD-STORE-011
    markRead(id) {
      const target = this.items.find((item) => item.id === id)
      if (target) {
        target.is_read = true
      }
    },
    // markAllRead は全てのエラーを既読にする。
    // 目的: 既読状態を一括更新する。
    // 入力: なし。
    // 出力: なし。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: すべての items が is_read=true。
    // 関連DD: DD-STORE-011
    markAllRead() {
      this.items.forEach((item) => {
        item.is_read = true
      })
    },
    // clearAll はエラー一覧を空にする。
    // 目的: エラー一覧を初期化する。
    // 入力: なし。
    // 出力: なし。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: items の長さは 0。
    // 関連DD: DD-STORE-011
    clearAll() {
      this.items = []
    },
    // clearBySource は指定ソースのエラーを削除する。
    // 目的: ソース別にエラーを整理する。
    // 入力: source は削除対象ソース。
    // 出力: なし。
    // エラー: なし。
    // 副作用: items を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: source が一致する items は残らない。
    // 関連DD: DD-STORE-011
    clearBySource(source) {
      this.items = this.items.filter((item) => item.source !== source)
    }
  }
})

// buildErrorItem は UiErrorItem の共通生成を行う。
// 目的: 生成ロジックを一箇所にまとめる。
// 入力: payload はUiErrorItemの構成要素。
// 出力: UiErrorItem。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: id と occurred_at は常に設定される。
// 関連DD: DD-STORE-009
function buildErrorItem(payload) {
  return {
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    occurred_at: new Date().toISOString(),
    source: payload.source ?? 'backend',
    action: payload.action ?? 'unknown',
    category: payload.category,
    issue_id: payload.issue_id,
    severity: payload.severity ?? 'error',
    api: payload.api,
    raw: payload.raw,
    user_message: payload.user_message ?? 'unexpected error',
    is_read: false
  }
}

// severityFromCode はエラーコードから重要度を決定する。
// 目的: エラー表示の重要度を統一する。
// 入力: code はApiErrorDTO.error_code。
// 出力: "info" | "warn" | "error"。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 不明なコードは error にする。
// 関連DD: DD-STORE-011
function severityFromCode(code) {
  switch (code) {
    case 'E_VALIDATION':
    case 'E_CONFLICT':
      return 'warn'
    case 'E_PERMISSION':
      return 'info'
    default:
      return 'error'
  }
}
