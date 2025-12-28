const pad2 = (value) => String(value).padStart(2, '0')

// formatJapaneseDateTime は DD-DATA-002 の表示形式で日時を整形する。
export const formatJapaneseDateTime = (value) => {
  const date = value instanceof Date ? value : new Date(value)
  return (
    `${date.getFullYear()}年` +
    `${pad2(date.getMonth() + 1)}月` +
    `${pad2(date.getDate())}日 ` +
    `${pad2(date.getHours())}時` +
    `${pad2(date.getMinutes())}分` +
    `${pad2(date.getSeconds())}秒`
  )
}
