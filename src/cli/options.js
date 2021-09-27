const path = require('path')
const {Option} = require('commander')

const configFile = () => {
  return new Option('-c, --config <config-file>', 'path to the config file')
    .default(path.join(process.cwd(), 'error-pages.yml'))
    .env('CONFIG_FILE')
}

module.exports = {
  configFile,
}
