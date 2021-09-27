const fs = require('fs')
const path = require('path')
const {readConfig} = require('./config')

describe('readConfig', () => {
  const configFilePath = path.join(__dirname, 'test_readConfig_' + Date.now().toString())

  beforeEach(() => {
    if (fs.existsSync(configFilePath)) {
      fs.rmSync(configFilePath)
    }
  })

  afterEach(() => {
    fs.rmSync(configFilePath) // comment this line for the test debugging
  });

  [
    {
      name: 'common usage',
      giveFileContent: `
templates:
  - name: foo
    path: foo path
  - path: bar
  - {}

pages:
  1:
    message: foo
    description: bar
  2:
  3: {}
`,
      wantResult: {
        templates: [
          {name: 'foo', path: 'foo path'},
          {name: '', path: 'bar'},
        ],
        pages: [
          {code: '1', message: 'foo', description: 'bar'},
          {code: '2', message: '', description: ''},
          {code: '3', message: '', description: ''},
        ],
      },
    },
    {
      name: 'empty',
      giveFileContent: ``,
      wantResult: {
        templates: [],
        pages: [],
      },
    },
    {
      name: 'wrong key types',
      giveFileContent: `
templates: foo
pages: 123`,
      wantResult: {
        templates: [],
        pages: [],
      },
    },
  ].forEach(tt => {
    test(tt.name, () => {
      fs.writeFileSync(configFilePath, tt.giveFileContent)

      expect.assertions(1)

      return readConfig(configFilePath).then(cfg => {
        expect(cfg).toEqual(tt.wantResult)
      })
    })
  })
})
