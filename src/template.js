/**
 * @param {string} templateContent
 * @param {string} code
 * @param {?string} message
 * @param {?string} description
 * @return {string}
 * @throws {Error} On empty code or invalid code type
 */
const build = (templateContent, code, message, description) => {
  if (typeof code !== 'string') {
    throw new Error('invalid code type')
  }

  const clearCode = code.replace(/([^a-zA-Z0-9_]+)/g, '')
  message = typeof message === 'string' ? message.trim() : ''
  description = typeof description === 'string' ? description.trim() : ''

  if (clearCode.length > 0) {
    return templateContent
      .replace(/{{\s?code\s?}}/g, clearCode)
      .replace(/{{\s?message\s?}}/g, message)
      .replace(/{{\s?description\s?}}/g, description)
  }

  throw new Error(`empty or invalid code: "${code}"`)
}

module.exports = {
  build,
}
