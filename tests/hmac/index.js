const crypto = require('crypto')

const secret = process.env.API_ACCESS_TOKEN

function genreply (bugID, secret) {
  const hash = crypto.createHmac('sha256', secret)
    .update(bugID)
    .digest('hex')

  if (process.env.STAGE) {
    return `reply+${bugID}-${hash}@${process.env.STAGE}.unee-t.com`
  }
  return `reply+${bugID}-${hash}@unee-t.com`
}

console.log(genreply('61825', secret))
