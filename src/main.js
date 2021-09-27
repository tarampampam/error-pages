#!/usr/bin/env node

const {Command} = require('commander')
const {BuildCommand} = require('./cli/build')
const {InitCommand} = require('./cli/init')

async function run() {
  await new Command()
    .addCommand(new BuildCommand().cmd)
    .addCommand(new InitCommand().cmd)
    .showHelpAfterError()
    .parseAsync(process.argv)
}

try {
  run()
} catch (e) {
  console.error(e)
  process.exit(1)
}
