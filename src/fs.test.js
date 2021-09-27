const fs = require('fs')
const path = require('path')
const {readFile: __readFile, mkdirForWriting, writeFile: __writeFile} = require('./fs')

describe('readFile', () => {
  [
    {
      name: 'common usage',
      giveFile: __filename,
      wantContent: fs.readFileSync(__filename).toString('utf8'),
    },
    {
      name: 'non-existent file',
      giveFile: 'foobar',
      wantError: true,
    },
    {
      name: 'empty file path',
      giveFile: '',
      wantError: true,
    },
  ].forEach(tt => {
    test(tt.name, () => {
      expect.assertions(1)

      if (tt.wantError === true) {
        return __readFile(tt.giveFile).catch(e => expect(e).toBeInstanceOf(Error))
      } else {
        return __readFile(tt.giveFile).then(data => {
          expect(data).toBe(tt.wantContent)
        })
      }
    })
  })
})

describe('mkdirForWriting', () => {
  const testDirPath = path.join(__dirname, 'test_mkdirForWriting_' + Date.now().toString())

  beforeEach(() => {
    if (fs.existsSync(testDirPath)) {
      fs.rmSync(testDirPath, {recursive: true, force: true})
    }

    fs.mkdirSync(testDirPath)
    fs.writeFileSync(path.join(testDirPath, 'exists'), 'lorem ipsum')
  })

  afterEach(() => {
    fs.rmSync(testDirPath, {recursive: true, force: true}) // comment this line for the test debugging
  });

  [
    {
      name: 'common usage',
      givePath: path.join(testDirPath, 'foo'),
      wantError: false,
    },
    {
      name: 'file exists',
      givePath: path.join(testDirPath, 'exists'),
      wantError: true,
    },
  ].forEach(tt => {
    test(tt.name, () => {
      if (tt.wantError === true) {
        expect.assertions(1)

        return mkdirForWriting(tt.givePath).catch(e => expect(e).toBeInstanceOf(Error))
      } else {
        expect.assertions(2)

        expect(fs.existsSync(tt.givePath)).toBe(false)

        return mkdirForWriting(tt.givePath).then(() => {
          expect(fs.existsSync(tt.givePath)).toBe(true)
          fs.accessSync(tt.givePath, fs.constants.R_OK | fs.constants.W_OK) // can throws an error
        })
      }
    })
  })
})

describe('writeFile', () => {
  const testDirPath = path.join(__dirname, 'test_writeFile_' + Date.now().toString())

  beforeEach(() => {
    if (fs.existsSync(testDirPath)) {
      fs.rmSync(testDirPath, {recursive: true, force: true})
    }

    fs.mkdirSync(testDirPath)
    fs.writeFileSync(path.join(testDirPath, 'exists'), '#####################################')
  })

  afterEach(() => {
    fs.rmSync(testDirPath, {recursive: true, force: true}) // comment this line for the test debugging
  });

  [
    {
      name: 'common usage',
      givePath: path.join(testDirPath, 'foo'),
      giveContent: 'foo bar\t\n',
      wantError: false,
      wantContent: 'foo bar\t\n',
    },
    {
      name: 'overwrite the existing file',
      givePath: path.join(testDirPath, 'exists'),
      giveContent: 'foo bar\t',
      wantError: false,
      wantContent: 'foo bar\t',
    },
    {
      name: 'wrong path (is directory path)',
      givePath: testDirPath,
      giveContent: 'foo',
      wantError: true,
    },
  ].forEach(tt => {
    test(tt.name, () => {
      if (tt.wantError === true) {
        expect.assertions(1)

        return __writeFile(tt.givePath, tt.giveContent).catch(e => expect(e).toBeInstanceOf(Error))
      } else {
        expect.assertions(2)

        return __writeFile(tt.givePath, tt.giveContent).then(() => {
          expect(fs.existsSync(tt.givePath)).toBe(true)
          expect(fs.readFileSync(tt.givePath).toString()).toBe(tt.wantContent)
        })
      }
    })
  })
})
