// issues.js は課題一覧のキャッシュとクエリ状態を管理し、UI描画は扱わない。
// フィルタリングの実適用はUI側で実施する。
import { defineStore } from 'pinia'

import { listIssues } from '../utils/apiClient'
import { useAppStore } from './app'
import { useErrorsStore } from './errors'

const DEFAULT_QUERY = {
  sort: { key: 'updated_at', dir: 'desc' },
  filter: {
    text: '',
    status: [],
    priority: [],
    dueDateFrom: null,
    dueDateTo: null,
    schemaInvalidOnly: false
  },
  page: 1
}

// useIssuesStore は DD-STORE-007/014 の課題一覧ストアを提供する。
// 目的: 課題一覧キャッシュと検索条件を管理する。
// 入力: Pinia の内部状態。
// 出力: issues ストア。
// エラー: なし。
// 副作用: なし。
// 並行性: Pinia の更新に従う。
// 不変条件: defaultQuery は不変として扱う。
// 関連DD: DD-STORE-007, DD-STORE-014
export const useIssuesStore = defineStore('issues', {
  state: () => ({
    issuesByCategory: {},
    queryByCategory: {},
    defaultQuery: DEFAULT_QUERY
  }),
  actions: {
    // loadIssues はカテゴリの課題一覧を読み込む。
    // 目的: バックエンドから一覧を取得しキャッシュに保存する。
    // 入力: category はカテゴリ名、opts はクエリ上書き。
    // 出力: IssueListDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: issuesByCategory[category] が更新される。
    // 関連DD: DD-STORE-014
    async loadIssues(category, opts = {}) {
      const errors = useErrorsStore()
      const app = useAppStore()
      const query = this.getQuery(category)
      const nextQuery = {
        ...query,
        ...opts,
        sort: { ...query.sort, ...(opts.sort ?? {}) },
        filter: { ...query.filter, ...(opts.filter ?? {}) }
      }
      this.queryByCategory[category] = nextQuery
      this.ensureCache(category)
      this.issuesByCategory[category].isLoading = true
      try {
        const request = {
          page: nextQuery.page,
          page_size: app.pageSize,
          sort_by: nextQuery.sort.key,
          sort_order: nextQuery.sort.dir
        }
        const data = await listIssues(category, request)
        this.issuesByCategory[category] = {
          items: data.issues ?? [],
          total: data.total ?? 0,
          lastLoadedAt: new Date().toISOString(),
          isLoading: false
        }
        return data
      } catch (e) {
        errors.capture(e, { source: 'issues', action: 'loadIssues', category })
        this.issuesByCategory[category].isLoading = false
        return null
      }
    },
    // refreshIssues はキャッシュを無視して再取得する。
    // 目的: 最新状態を取得する。
    // 入力: category はカテゴリ名。
    // 出力: IssueListDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: loadIssues を強制実行する。
    // 関連DD: DD-STORE-014
    async refreshIssues(category) {
      return this.loadIssues(category, {})
    },
    // setSort はソート条件を更新して一覧を再読み込みする。
    // 目的: ソート条件を適用する。
    // 入力: category はカテゴリ名、sort はソート条件。
    // 出力: IssueListDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: queryByCategory[category].sort が更新される。
    // 関連DD: DD-STORE-014
    async setSort(category, sort) {
      return this.loadIssues(category, { sort, page: 1 })
    },
    // setFilter はフィルタ条件を更新する。
    // 目的: UI側のフィルタ条件を保持する。
    // 入力: category はカテゴリ名、filter は条件。
    // 出力: なし。
    // エラー: なし。
    // 副作用: 状態を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: queryByCategory[category].filter が更新される。
    // 関連DD: DD-STORE-014
    setFilter(category, filter) {
      const query = this.getQuery(category)
      this.queryByCategory[category] = {
        ...query,
        filter: { ...query.filter, ...filter }
      }
    },
    // setPage はページ番号を更新して一覧を再読み込みする。
    // 目的: ページングを適用する。
    // 入力: category はカテゴリ名、page はページ番号。
    // 出力: IssueListDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: queryByCategory[category].page が更新される。
    // 関連DD: DD-STORE-014
    async setPage(category, page) {
      return this.loadIssues(category, { page })
    },
    // invalidateCategory は指定カテゴリのキャッシュを破棄する。
    // 目的: 再読み込みを促す。
    // 入力: category はカテゴリ名。
    // 出力: なし。
    // エラー: なし。
    // 副作用: issuesByCategory を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: キャッシュが削除される。
    // 関連DD: DD-STORE-014
    invalidateCategory(category) {
      delete this.issuesByCategory[category]
      delete this.queryByCategory[category]
    },
    // applyIssueUpdatedToCache は詳細更新後に一覧キャッシュを更新する。
    // 目的: 既存キャッシュを最新内容に同期する。
    // 入力: issueDetail は課題詳細DTO。
    // 出力: なし。
    // エラー: なし。
    // 副作用: issuesByCategory を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: 同一 issue_id の項目のみ更新する。
    // 関連DD: DD-STORE-014
    applyIssueUpdatedToCache(issueDetail) {
      const category = issueDetail.category
      const entry = this.issuesByCategory[category]
      if (!entry) {
        return
      }
      const target = entry.items.find((item) => item.issue_id === issueDetail.issue_id)
      if (!target) {
        return
      }
      target.title = issueDetail.title
      target.status = issueDetail.status
      target.priority = issueDetail.priority
      target.origin_company = issueDetail.origin_company
      target.updated_at = issueDetail.updated_at
      target.due_date = issueDetail.due_date
      target.is_schema_invalid = issueDetail.is_schema_invalid
    },
    // renameCategoryKey はカテゴリ名変更に合わせてキャッシュキーを移動する。
    // 目的: カテゴリ名変更後もキャッシュを引き継ぐ。
    // 入力: oldName は旧名、newName は新名。
    // 出力: なし。
    // エラー: なし。
    // 副作用: issuesByCategory/queryByCategory を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: oldName のエントリは削除される。
    // 関連DD: DD-STORE-013
    renameCategoryKey(oldName, newName) {
      if (this.issuesByCategory[oldName]) {
        this.issuesByCategory[newName] = this.issuesByCategory[oldName]
        delete this.issuesByCategory[oldName]
      }
      if (this.queryByCategory[oldName]) {
        this.queryByCategory[newName] = this.queryByCategory[oldName]
        delete this.queryByCategory[oldName]
      }
    },
    // ensureCache はカテゴリキャッシュの初期形を用意する。
    // 目的: キャッシュ参照時の undefined を防ぐ。
    // 入力: category はカテゴリ名。
    // 出力: なし。
    // エラー: なし。
    // 副作用: issuesByCategory を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: キャッシュエントリが必ず存在する。
    // 関連DD: DD-STORE-014
    ensureCache(category) {
      if (!this.issuesByCategory[category]) {
        this.issuesByCategory[category] = {
          items: [],
          total: 0,
          lastLoadedAt: null,
          isLoading: false
        }
      }
    },
    // getQuery はカテゴリごとのクエリ状態を取得する。
    // 目的: クエリの初期化と取得を行う。
    // 入力: category はカテゴリ名。
    // 出力: IssuesQueryState。
    // エラー: なし。
    // 副作用: queryByCategory を更新する。
    // 並行性: Pinia の更新に従う。
    // 不変条件: 返却値は defaultQuery から生成される。
    // 関連DD: DD-STORE-014
    getQuery(category) {
      if (!this.queryByCategory[category]) {
        this.queryByCategory[category] = cloneQuery(this.defaultQuery)
      }
      return this.queryByCategory[category]
    }
  }
})

// cloneQuery はクエリ状態のディープコピーを作る。
// 目的: defaultQuery の参照共有を避ける。
// 入力: query はコピー元。
// 出力: 新しいクエリオブジェクト。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 返却値は query と独立する。
// 関連DD: DD-STORE-014
function cloneQuery(query) {
  return JSON.parse(JSON.stringify(query))
}
