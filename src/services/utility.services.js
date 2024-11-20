/** @param {string} text */
export const extractMentions = (text) => {
  const matches = text.match(/(?<=@)\w+/g)
  return matches ? [...new Set(matches)] : []
}

/** @param {string} text */
export const extractHashtags = (text) => {
  const matches = text.match(/(?<=#)\w+/g)
  return matches ? [...new Set(matches)] : []
}