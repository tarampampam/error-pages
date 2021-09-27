const {Command, Option} = require('commander')
const {configFile} = require('./options')
const spawn = require('child_process').spawn
const fs = require('fs')

class InitCommand {
  /** @type Command */
  cmd

  constructor() {
    this.cmd = new Command('init')
      .showHelpAfterError()
      .description('run the initialization process (for usage as a docker image entrypoint)')
      .addOption(configFile())
      .addOption(
        new Option(
          '-t, --template-name <template-name>',
          'use defined template name (set "random" to use the randomized template)',
        )
          .env('TEMPLATE_NAME')
      )
      .addOption(
        new Option('--templates-dir <templates-dir>', 'path to the directory with error page templates')
          .env('TEMPLATES_DIR')
      )
      .addOption(
        new Option('--custom-template <template-body>', 'use this as a custom template')
          .env('CUSTOM_TEMPLATE')
      )
      .addOption(new Option('-j, --log-json', 'logs in json format',).env('LOG_JSON'))
      .action(this.exec)
  }

  /**
   * @param {{config: string, templateName: string|undefined, logJson: boolean, templatesDir: string|undefined, customTemplate: string|undefined}} opt
   * @param {Command} cmd
   * @return {Promise<void>}
   */
  exec(opt, cmd) {
    let templatesList = []

    console.log('WIP - implement this', opt.customTemplate)

    if (typeof opt.templatesDir === 'string') {
      try {
        templatesList = fs.readdirSync(opt.templatesDir, {withFileTypes: true})
          .filter(dirent => dirent.isDirectory())
          .map(dirent => dirent.name)
      } catch (e) {
        InitCommand.log(opt.logJson, 'templates list fetching failed', e)
      }
    }

    if (typeof opt.templateName === 'string') {
      if (templatesList.length > 0) {
        if (opt.templateName.toLowerCase().trim() === 'random') {
          const randomTpl = templatesList[Math.floor(Math.random()*templatesList.length)]

          InitCommand.log(opt.logJson, 'use randomly selected template', {template: randomTpl})
          process.env['TEMPLATE_NAME'] = randomTpl
        } else {
          if (templatesList.includes(opt.templateName)) {
            InitCommand.log(opt.logJson, 'use requested template', {template: opt.templateName})
            process.env['TEMPLATE_NAME'] = opt.templateName
          } else {
            InitCommand.log(opt.logJson, 'requested nonexistent template', {template: opt.templateName})
          }
        }
      } else {
        InitCommand.log(opt.logJson, 'cannot set default template', {template: opt.templateName})
      }
    }

    if (cmd.args.length > 0) {
      const commandToRun = cmd.args[0], commandArgs = cmd.args.slice(1, cmd.args.length)

      let childProcessIsAlive = true

      const childProc = spawn(commandToRun, commandArgs, {
        cwd: process.cwd(),
        env: process.env,
        stdio: 'inherit',
        shell: false,
      })

      InitCommand.log(opt.logJson, 'child process started', {cmd: commandToRun, args: commandArgs, pid: childProc.pid})

      childProc.on('exit', (code, signal) => {
        childProcessIsAlive = false
        InitCommand.log(opt.logJson, 'child process ends', {pid: childProc.pid, signal: signal, code: code})

        process.exit(code) // force the init process exit
      })

      /** @param {string} signal */
      const killChildProc = (signal) => {
        if (childProcessIsAlive === true) {
          InitCommand.log(opt.logJson, 'killing the child process', {signal: signal})
          childProc.kill(signal)
        }
      }

      // subscribe for the termination signals and proxypass them to the child process
      ['SIGINT', 'SIGTERM'].forEach(signal => process.on(signal, signal => {
        InitCommand.log(opt.logJson, 'termination signal received', {signal: signal})
        killChildProc(signal)
      }))

      // not sure it's necessary, but...
      process.on('exit', () => killChildProc('SIGTERM'))
    }
  }

  /**
   * @param {boolean} json
   * @param {string} msg
   * @param {object|null} payload
   */
  static log(json, msg, payload = null) {
    const memUsage = `${Math.round(process.memoryUsage().heapUsed / 1024 / 1024 * 100) / 100}mb`

    /** @type {any[]} */
    let output = []

    if (json === true) {
      output.push(JSON.stringify({
        init: msg,
        payload: payload,
        main_proc_pid: process.pid,
        init_mem_usage: memUsage,
      }))
    } else {
      output.push(msg, payload, {main_proc_pid: process.pid, init_mem_usage: memUsage})
    }

    console.log(...output)
  }
}

module.exports = {
  InitCommand,
}
