const crypto = require('crypto')

const secret = 'foobar'
const hash = crypto.createHmac('sha256', secret)
  .update('12345')
  .digest('hex')
console.log(hash)

function genreply (bugID, secret) {
  const hash = crypto.createHmac('sha256', secret)
    .update(bugID)
    .digest('hex')

  return `reply+${bugID}-${hash}@dev.unee-t.com`
}

console.log(genreply('12345', 'foobar'))
