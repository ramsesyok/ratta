// issueDetail.js は課題詳細の状態管理を担い、UIの描画は扱わない。
// 詳細データは常にバックエンドから再取得する。
import { defineStore } from 'pinia'

import { addComment, getIssue, updateIssue } from '../utils/apiClient'
import { useErrorsStore } from './errors'
import { useIssuesStore } from './issues'

// useIssueDetailStore は DD-STORE-008/015 の課題詳細ストアを提供する。
// 目的: 課題詳細の取得・更新・コメント追加を管理する。
// 入力: Pinia の内部状態。
// 出力: issueDetail ストア。
// エラー: なし。
// 副作用: なし。
// 並行性: Pinia の更新に従う。
// 不変条件: current は最新の取得結果のみ保持する。
// 関連DD: DD-STORE-008, DD-STORE-015
export const useIssueDetailStore = defineStore('issueDetail', {
  state: () => ({
    current: null,
    currentCategory: null,
    isLoading: false,
    isDirty: false,
    lastLoadedAt: null
  }),
  actions: {
    // openIssue は課題詳細を読み込む。
    // 目的: 画面表示用に最新の課題詳細を取得する。
    // 入力: category はカテゴリ名、issueId は課題ID。
    // 出力: IssueDetailDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: current と currentCategory を更新する。
    // 関連DD: DD-STORE-015
    async openIssue(category, issueId) {
      const errors = useErrorsStore()
      this.isLoading = true
      try {
        const data = await getIssue(category, issueId)
        this.current = data
        this.currentCategory = category
        this.isDirty = false
        this.lastLoadedAt = new Date().toISOString()
        return data
      } catch (e) {
        errors.capture(e, { source: 'issueDetail', action: 'openIssue', category, issue_id: issueId })
        return null
      } finally {
        this.isLoading = false
      }
    },
    // reloadCurrent は current がある場合に再読み込みする。
    // 目的: 最新状態を再取得する。
    // 入力: なし。
    // 出力: IssueDetailDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: current が null の場合は何もしない。
    // 関連DD: DD-STORE-015
    async reloadCurrent() {
      if (!this.current || !this.currentCategory) {
        return null
      }
      return this.openIssue(this.currentCategory, this.current.issue_id)
    },
    // saveIssue は課題更新を行い current を更新する。
    // 目的: 課題の更新結果を反映する。
    // 入力: update は IssueUpdateDTO。
    // 出力: IssueDetailDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しと issues キャッシュ更新を行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: schema invalid の課題は更新しない。
    // 関連DD: DD-STORE-015
    async saveIssue(update) {
      const errors = useErrorsStore()
      if (!this.current || !this.currentCategory) {
        return null
      }
      if (this.current.is_schema_invalid) {
        errors.capture(
          new Error('schema invalid issue is read-only'),
          { source: 'issueDetail', action: 'saveIssue', category: this.currentCategory, issue_id: this.current.issue_id }
        )
        return null
      }
      this.isLoading = true
      try {
        const data = await updateIssue(this.currentCategory, this.current.issue_id, update)
        this.current = data
        this.isDirty = false
        this.lastLoadedAt = new Date().toISOString()
        const issues = useIssuesStore()
        issues.applyIssueUpdatedToCache(data)
        return data
      } catch (e) {
        errors.capture(e, { source: 'issueDetail', action: 'saveIssue', category: this.currentCategory, issue_id: this.current.issue_id })
        return null
      } finally {
        this.isLoading = false
      }
    },
    // addComment はコメント追加を行い current を更新する。
    // 目的: コメント追加結果を反映する。
    // 入力: payload は CommentCreateDTO。
    // 出力: IssueDetailDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しと issues キャッシュ更新を行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: current が null の場合は何もしない。
    // 関連DD: DD-STORE-015
    async addComment(payload) {
      const errors = useErrorsStore()
      if (!this.current || !this.currentCategory) {
        return null
      }
      this.isLoading = true
      try {
        const data = await addComment(this.currentCategory, this.current.issue_id, payload)
        this.current = data
        this.lastLoadedAt = new Date().toISOString()
        const issues = useIssuesStore()
        issues.applyIssueUpdatedToCache(data)
        return data
      } catch (e) {
        errors.capture(e, { source: 'comments', action: 'addComment', category: this.currentCategory, issue_id: this.current.issue_id })
        return null
      } finally {
        this.isLoading = false
      }
    },
    // markDirty は編集状態を更新する。
    // 目的: 編集中フラグを設定する。
    // 入力: value は真偽値。
    // 出力: なし。
    // エラー: なし。
    // 副作用: 状態を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: isDirty は value と一致する。
    // 関連DD: DD-STORE-008
    markDirty(value) {
      this.isDirty = value
    }
  }
})
