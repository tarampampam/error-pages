const {build} = require('./template')

describe('build', () => {
  [
    {
      name: 'empty',
      giveTemplate: '',
      giveCode: 'foo',
      giveMessage: undefined,
      giveDescription: undefined,
      wantResult: '',
    },
    {
      name: 'common template',
      giveTemplate: '{{ code}}_{{code }}. {{message}}:\t {{ description }}  ',
      giveCode: '1',
      giveMessage: ' foo',
      giveDescription: 'bar\t ',
      wantResult: '1_1. foo:\t bar  ',
    },
    {
      name: 'alpha and underline in the code',
      giveTemplate: '\t{{ code }}\t',
      giveCode: '  qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_  ',
      wantResult: '\tqwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_\t',
    },
    {
      name: 'undefined code',
      giveTemplate: '{{ code }}',
      giveCode: undefined,
      wantException: 'invalid code type',
    },
    {
      name: 'wrong code',
      giveTemplate: '{{ code }}',
      giveCode: '~`!@#$%^&*()+[]{}\\|\';:<>?/',
      wantException: 'empty or invalid code',
    },
  ].forEach((tt) => {
    test(tt.name, () => {
      if (typeof tt.wantException === 'string') {
        expect(() => {
          build(tt.giveTemplate, tt.giveCode, tt.giveMessage, tt.giveDescription)
        }).toThrow(tt.wantException)
      } else {
        expect(build(tt.giveTemplate, tt.giveCode, tt.giveMessage, tt.giveDescription)).toBe(tt.wantResult)
      }
    })
  })
})
