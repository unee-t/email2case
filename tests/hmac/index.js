const crypto = require('crypto')

const secret = process.env.API_ACCESS_TOKEN
console.log('Secret is', secret)

function genreply (bugID, secret) {
  const hash = crypto.createHmac('sha256', secret)
    .update(bugID)
    .digest('hex')

  return `reply+${bugID}-${hash}@dev.unee-t.com`
}

console.log(genreply('87654', secret))
