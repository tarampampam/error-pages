const yaml = require('yaml')
const fs = require('./fs')

class ConfigTemplate {
  /** @type {string} */
  name
  /** @type {string} */
  path
}

class ConfigPage {
  /** @type {string} */
  code
  /** @type {string} */
  message
  /** @type {string} */
  description
}

class Config {
  /** @type {ConfigTemplate[]} */
  templates = []
  /** @type {ConfigPage[]} */
  pages = []

  /**
   * @param {{templates: {name, path: any}[]|undefined|null, pages: Object.<any, {message, description: any}>|undefined|null}|null} input
   * @return {this}
   */
  static fromObject(input) {
    const cfg = new Config

    if (typeof input !== 'object' || input === null) { // fast break for the empty input
      return cfg
    }

    if (Object.prototype.hasOwnProperty.call(input, 'templates') && Array.isArray(input.templates)) {
      input.templates.forEach(p => {
        if (typeof p === 'object' && p !== null) {
          const template = new ConfigTemplate

          template.name = p.name !== undefined ? String(p.name).trim() : ''
          template.path = p.path !== undefined ? String(p.path).trim() : ''

          if (template.name.length + template.path.length > 0) { // skip skip empty entries
            cfg.templates.push(template)
          }
        }
      })
    }

    if (Object.prototype.hasOwnProperty.call(input, 'pages') && typeof input.pages === 'object') {
      for (const [code, props] of Object.entries(input.pages)) {
        const page = new ConfigPage

        page.code = typeof code === 'string' ? code.trim() : String(code).trim()

        if (typeof props === 'object' && props !== null) {
          page.message = props.message !== undefined ? String(props.message).trim() : ''
          page.description = props.description !== undefined ? String(props.description).trim() : ''
        } else {
          page.message = ''
          page.description = ''
        }

        cfg.pages.push(page)
      }
    }

    return cfg
  }
}

/**
 * @param {string} configFilePath
 * @return {Promise<Config>}
 */
const readConfig = async (configFilePath) => {
  return new Promise((resolve, reject) => {
    fs.readFile(configFilePath)
      .then(content => {
        try {
          const y = yaml.parse(content) // can throw an error

          resolve(Config.fromObject(y))
        } catch (e) {
          reject(e)
        }
      })
      .catch(reject)
  })
}

module.exports = {
  readConfig,
}
