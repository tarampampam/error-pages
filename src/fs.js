const fs = require("fs")

/**
 * @param {string} filePath
 * @return {Promise<string>}
 */
const readFile = async (filePath) => {
  return new Promise((resolve, reject) => {
    if (typeof filePath !== 'string' || filePath.length === 0) {
      return reject(new Error('file path was not provided'))
    }

    fs.stat(filePath, (err, stat) => {
      if (err === null) {
        if (!stat.isFile()) {
          return reject(new Error(`"${filePath}" is not a regular file`))
        }

        fs.readFile(filePath, {flag: 'r', encoding: 'utf8'}, (err, data) => {
          if (err !== null) {
            return reject(err)
          }

          resolve(data)
        })
      } else if (err.code === 'ENOENT') {
        reject(new Error(`file "${filePath}" does not exist`))
      } else {
        reject(err)
      }
    })
  })
}

/**
 * @param {string} dirPath
 * @return {Promise<void>}
 */
const mkdirForWriting = async (dirPath) => {
  return new Promise((resolve, reject) => {
    fs.stat(dirPath, (err, stat) => {
      if (err === null) {
        if (!stat.isDirectory()) {
          return reject(new Error(`"${dirPath}" already exists and it's not a directory`))
        }

        fs.access(dirPath, fs.constants.W_OK, (err) => {
          if (err !== null) {
            return reject(new Error(`directory "${dirPath}" is not writable`))
          }

          resolve()
        })
      } else if (err.code === 'ENOENT') {
        fs.mkdir(dirPath, {recursive: true, mode: 0o775}, (err) => {
          if (err !== null) {
            return reject(err)
          }

          resolve()
        })
      } else {
        reject(err)
      }
    })
  })
}

/**
 * @param {string} filePath
 * @param {string} content
 * @return {Promise<void>}
 */
const writeFile = async (filePath, content) => {
  return new Promise((resolve, reject) => {
    fs.open(filePath, 'w', (err, fd) => {
      if (err !== null) {
        return reject(new Error(`cannot open file "${filePath}" for writing`))
      }

      fs.write(fd, content, 0, 'utf8', (err) => {
        fs.close(fd)

        if (err === null) {
          return resolve()
        }

        reject(err)
      })
    })
  })
}

module.exports = {
  readFile,
  mkdirForWriting,
  writeFile,
}
