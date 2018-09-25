#!/bin/bash

for STAGE in dev demo prod
do

	FROM="prod Unee-T case <test@unee-t.com>"

if test "$STAGE" != "prod"
then
	FROM="$STAGE Unee-T case <test@$STAGE.unee-t.com>"
fi

MAIL_URL=$(aws --profile uneet-${STAGE} ssm get-parameters --names MAIL_URL --with-decryption --query Parameters[0].Value --output text)

cat << END > index.$STAGE.js
var nodemailer = require('nodemailer');

// create reusable transporter object using the default SMTP transport
var transporter = nodemailer.createTransport('${MAIL_URL}');

// setup e-mail data with unicode symbols
var mailOptions = {
    from: '${FROM}>', // sender address
    to: 'kai.hendry@gmail.com', // list of receivers
    subject: 'Hello ✔', // Subject line
    text: 'Hello world ?', // plaintext body
    html: '<b>Hello world ?</b>' // html body
};

// send mail with defined transport object
transporter.sendMail(mailOptions, function(error, info){
    if(error){
        return console.log(error);
    }
    console.log('Message sent: ' + info.response);
});
END

done
