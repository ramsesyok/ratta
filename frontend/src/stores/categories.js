// categories.js はカテゴリ一覧と選択状態を管理し、UIの描画は扱わない。
// 読み込み結果の整形はバックエンドDTOに従う。
import { defineStore } from 'pinia'

import { createCategory, deleteCategory, listCategories, renameCategory } from '../utils/apiClient'
import { useAppStore } from './app'
import { useErrorsStore } from './errors'
import { useIssuesStore } from './issues'

// useCategoriesStore は DD-STORE-006/013 のカテゴリストアを提供する。
// 目的: カテゴリ一覧の取得と選択状態を管理する。
// 入力: Pinia の内部状態。
// 出力: categories ストア。
// エラー: なし。
// 副作用: なし。
// 並行性: Pinia の更新に従う。
// 不変条件: selectedCategory は items 内の name か null。
// 関連DD: DD-STORE-006, DD-STORE-013
export const useCategoriesStore = defineStore('categories', {
  state: () => ({
    items: [],
    selectedCategory: null,
    isLoading: false,
    lastLoadedAt: null
  }),
  actions: {
    // loadCategories はカテゴリ一覧を読み込む。
    // 目的: バックエンドからカテゴリ一覧を取得する。
    // 入力: なし。
    // 出力: CategoryListDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: items と lastLoadedAt を更新する。
    // 関連DD: DD-STORE-013
    async loadCategories() {
      const errors = useErrorsStore()
      this.isLoading = true
      try {
        const data = await listCategories()
        this.items = data.categories ?? []
        this.lastLoadedAt = new Date().toISOString()
        return data
      } catch (e) {
        errors.capture(e, { source: 'categories', action: 'loadCategories' })
        return null
      } finally {
        this.isLoading = false
      }
    },
    // selectCategory は選択カテゴリを更新し、課題一覧を読み込む。
    // 目的: 選択変更と課題一覧の同期を行う。
    // 入力: name はカテゴリ名。
    // 出力: なし。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: issues ストアの読み込みを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: selectedCategory が更新される。
    // 関連DD: DD-STORE-013
    async selectCategory(name) {
      const errors = useErrorsStore()
      this.selectedCategory = name
      try {
        const issues = useIssuesStore()
        await issues.loadIssues(name)
      } catch (e) {
        errors.capture(e, { source: 'categories', action: 'selectCategory', category: name })
      }
    },
    // createCategory はカテゴリを作成して一覧を更新する。
    // 目的: Contractor 操作でカテゴリを新規作成する。
    // 入力: name はカテゴリ名。
    // 出力: CategoryDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: 成功時に items が更新される。
    // 関連DD: DD-STORE-013
    async createCategory(name) {
      const errors = useErrorsStore()
      const app = useAppStore()
      if (app.mode !== 'Contractor') {
        errors.captureApiError(
          new PermissionError('Vendor cannot create category'),
          { source: 'categories', action: 'createCategory' }
        )
        return null
      }
      try {
        const data = await createCategory(name)
        await this.loadCategories()
        return data
      } catch (e) {
        errors.capture(e, { source: 'categories', action: 'createCategory' })
        return null
      }
    },
    // renameCategory はカテゴリ名を変更し、選択状態とキャッシュを更新する。
    // 目的: Contractor 操作でカテゴリ名を変更する。
    // 入力: oldName は旧名、newName は新名。
    // 出力: CategoryDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しと issues キャッシュの更新を行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: selectedCategory が更新される。
    // 関連DD: DD-STORE-013
    async renameCategory(oldName, newName) {
      const errors = useErrorsStore()
      const app = useAppStore()
      if (app.mode !== 'Contractor') {
        errors.captureApiError(
          new PermissionError('Vendor cannot rename category'),
          { source: 'categories', action: 'renameCategory' }
        )
        return null
      }
      try {
        const data = await renameCategory(oldName, newName)
        const issues = useIssuesStore()
        issues.renameCategoryKey(oldName, newName)
        if (this.selectedCategory === oldName) {
          this.selectedCategory = newName
        }
        await this.loadCategories()
        return data
      } catch (e) {
        errors.capture(e, { source: 'categories', action: 'renameCategory', category: oldName })
        return null
      }
    },
    // deleteCategory はカテゴリを削除して一覧を更新する。
    // 目的: Contractor 操作でカテゴリを削除する。
    // 入力: name はカテゴリ名。
    // 出力: なし。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しと issues キャッシュの更新を行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: selectedCategory が削除対象なら null にする。
    // 関連DD: DD-STORE-013
    async deleteCategory(name) {
      const errors = useErrorsStore()
      const app = useAppStore()
      if (app.mode !== 'Contractor') {
        errors.captureApiError(
          new PermissionError('Vendor cannot delete category'),
          { source: 'categories', action: 'deleteCategory' }
        )
        return null
      }
      try {
        const data = await deleteCategory(name)
        const issues = useIssuesStore()
        issues.invalidateCategory(name)
        if (this.selectedCategory === name) {
          this.selectedCategory = null
        }
        await this.loadCategories()
        return data
      } catch (e) {
        errors.capture(e, { source: 'categories', action: 'deleteCategory', category: name })
        return null
      }
    }
  }
})

// PermissionError は Vendor 操作を拒否するための疑似 API エラーを表す。
// 目的: 権限不足を ApiErrorDTO 相当として登録する。
// 入力: message は表示用メッセージ。
// 出力: PermissionError。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: errorCode は E_PERMISSION。
// 関連DD: DD-STORE-013
class PermissionError {
  constructor(message) {
    this.errorCode = 'E_PERMISSION'
    this.message = message
    this.detail = ''
    this.targetPath = ''
    this.hint = ''
    this.action = 'permission'
  }
}
