const path = require('path')
const {Command, Argument, Option} = require('commander')
const {configFile} = require('./options')
const {readConfig} = require('../config')
const {readFile, mkdirForWriting, writeFile} = require('../fs')
const {build: buildTemplate} = require('../template')

class BuildCommand {
  /** @type Command */
  cmd

  constructor() {
    this.cmd = new Command('build')
      .showHelpAfterError()
      .description('build the templates')
      .addArgument(new Argument('output-dir', 'output directory').argRequired())
      .addOption(configFile())
      .addOption(new Option('-i, --index', 'generate index page'))
      .action(this.exec)
  }

  /**
   * @param {string} outputDir
   * @param {{config: string; index: ?boolean}} opt
   * @return {Promise<void>}
   */
  async exec(outputDir, opt) {
    return new Promise((resolve, reject) => {
      readConfig(opt.config)
        .then(async cfg => {
          if (cfg.templates.length === 0) {
            throw new Error('no templates in the config file')
          }

          if (cfg.pages.length === 0) {
            throw new Error('no error pages in the config file')
          }

          /** @type {Promise<string>[]} */
          const tplJobs = []

          const totalTimerName = 'total templates building time'
          console.time(totalTimerName)

          /** @type {Object<string, Array.<{code, message, path: string}>>} */
          const buildHistory = {}

          cfg.templates.forEach(tpl => {
            if (tpl.name.length === 0 && tpl.path.length !== 0) { // set the name based on file path, if needed
              tpl.name = path.parse(tpl.path).name
            }

            if (tpl.path.length === 0) { // skip templates without path
              return
            }

            const timerName = `template "${tpl.name}" (${tpl.path}) built in`
            console.time(timerName)

            tplJobs.push(new Promise((tplJobResolve, tplJobReject) => {
              readFile(tpl.path)
                .then(tplContent => {
                  /** @type {Promise<void>[]} */
                  const pageJobs = []

                  cfg.pages.forEach(page => {
                    pageJobs.push(new Promise((pageJobResolve, pageJobReject) => {
                      const pageContent = buildTemplate(tplContent, page.code, page.message, page.description)

                      mkdirForWriting(path.join(outputDir, tpl.name))
                        .then(() => {
                          const errorFileName = `${page.code}.html`

                          writeFile(path.join(outputDir, tpl.name, errorFileName), pageContent)
                            .then(() => {
                              if (!Object.prototype.hasOwnProperty.call(buildHistory, tpl.name)) {
                                buildHistory[tpl.name] = []
                              }

                              buildHistory[tpl.name].push({
                                code: page.code,
                                message: page.message,
                                path: path.join(tpl.name, errorFileName),
                              })

                              pageJobResolve()
                            })
                            .catch(pageJobReject)
                        })
                        .catch(pageJobReject)
                    }))
                  })

                  Promise.all(pageJobs)
                    .then(() => tplJobResolve(timerName))
                    .catch(tplJobReject)
                })
                .catch(tplJobReject)
            }))
          })

          Promise
            .all(tplJobs)
            .then(async results => {
              results.forEach(timerName => console.timeEnd(timerName))
              console.timeEnd(totalTimerName)

              if (opt.index === true) {
                for (const [tplName] of Object.entries(buildHistory)) { // make sort
                  buildHistory[tplName].sort((a, b) => {
                    return a.code - b.code
                  })
                }

                const indexTimer = 'index page building time'
                console.time(indexTimer)

                await BuildCommand.generateIndexPage(path.join(outputDir, 'index.html'), buildHistory)

                console.timeEnd(indexTimer)
              }

              resolve()
            })
            .catch(reject)
        })
        .catch(reject)
    })
  }

  /**
   * @param {string} indexPagePath
   * @param {Object<string, Array.<{code, message, path: string}>>} buildHistory
   * @return {Promise<void>}
   */
  static async generateIndexPage(indexPagePath, buildHistory) {
    return new Promise((resolve, reject) => {
      let lines = []

      for (const [tplName, data] of Object.entries(buildHistory)) {
        lines.push(`    <h2 class="mb-3">Template name: <code>${tplName}</code></h2>`)
        lines.push(`    <ul class="mb-5">`)

        data.forEach(props => {
          lines.push(`      <li><a href="${props.path}"><strong>${props.code}</strong>: ${props.message}</a></li>`)
        })

        lines.push(`    </ul>`)
      }

      writeFile(indexPagePath, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
  <title>Error pages list</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/5.1.1/css/bootstrap.min.css"
    integrity="sha512-6KY5s6UI5J7SVYuZB4S/CZMyPylqyyNZco376NM2Z8Sb8OxEdp02e1jkKk/wZxIEmjQ6DRCEBhni+gpr9c4tvA=="
    crossorigin="anonymous" referrerpolicy="no-referrer" />
</head>
<body class="bg-light">
<div class="container">
  <main>
    <div class="py-5 text-center">
      <img class="d-block mx-auto mb-4" src="https://hsto.org/webt/rm/9y/ww/rm9ywwx3gjv9agwkcmllhsuyo7k.png" alt="" width="94">
      <h2>Error pages index</h2>
    </div>
${lines.join('\n')}
  </main>
</div>
<footer class="footer">
  <div class="container text-center text-muted mt-3 mb-3">
    For online documentation and support please refer to the <a href="https://github.com/tarampampam/error-pages">project repository</a>.
  </div>
</footer>
</body>
</html>`)
        .then(resolve)
        .catch(reject)
    })
  }
}

module.exports = {
  BuildCommand,
}
