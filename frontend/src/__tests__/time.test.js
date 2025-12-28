import { describe, expect, it } from 'vitest'
import { formatJapaneseDateTime } from '../utils/time'

describe('formatJapaneseDateTime', () => {
  it('DD-DATA-002 の表示形式で日時を整形する', () => {
    // ローカル時刻の組み立てでも表示フォーマットが維持されることを確認する。
    const value = new Date(2024, 0, 2, 3, 4, 5)
    expect(formatJapaneseDateTime(value)).toBe('2024年01月02日 03時04分05秒')
  })
})
