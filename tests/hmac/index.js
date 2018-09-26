const crypto = require('crypto')

const secret = process.env.API_ACCESS_TOKEN

function genreply (bugID, userid, secret) {
  const hash = crypto.createHmac('sha256', secret)
    .update(`${bugID}${userid}`)
    .digest('hex')

  if (process.env.STAGE) {
    return `reply+${bugID}-${userid}-${hash}@${process.env.STAGE}.unee-t.com`
  }
  return `reply+${bugID}-${userid}-${hash}@unee-t.com`
}

console.log(genreply('61825', 107, secret))
