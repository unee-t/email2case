const crypto = require('crypto')

const secret = 'foobar'
const hash = crypto.createHmac('sha256', secret)
  .update('12345')
  .digest('base64')
console.log(hash)
