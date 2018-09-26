#!/bin/bash

for STAGE in dev demo prod
do

FROM="prod Unee-T case <test@unee-t.com>"

case $STAGE in
	"dev")
		bugid=62321
		;;
	"demo")
		bugid=68
		;;
	*)
		echo Unknown $STAGE
		exit
		;;
esac

CASE=https://case.unee-t.com/case/$bugid

if test "$STAGE" != "prod"
then
	FROM="$STAGE Unee-T case <test@$STAGE.unee-t.com>"
	CASE="https://case.$STAGE.unee-t.com/case/$bugid"
fi

MAIL_URL=$(aws --profile uneet-${STAGE} ssm get-parameters --names MAIL_URL --with-decryption --query Parameters[0].Value --output text)
SECRET=$(aws --profile uneet-${STAGE} ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text)


REPLY=reply+$bugid-$(echo -n $bugid | hmac256 $SECRET)@$STAGE.unee-t.com
if test $STAGE == "prod"
then
	REPLY=reply+$bugid-$(echo -n $bugid | hmac256 $SECRET)@unee-t.com
fi

DATE=$(date)

cat << END > index.$STAGE.js
var nodemailer = require('nodemailer')

// create reusable transporter object using the default SMTP transport
var transporter = nodemailer.createTransport('${MAIL_URL}')

// setup e-mail data with unicode symbols
var mailOptions = {
	from: '${FROM}', // sender address must be whitelisted for this to work
	to: '${REPLY}',
	subject: 'Testing ${DATE}', // Subject line
	text: 'Text message ${DATE}', // plaintext body
	html: '<b>Html message ${DATE}</b>' // html body
};

// send mail with defined transport object
transporter.sendMail(mailOptions, function(error, info){
    if(error){
        return console.log(error);
    }
    console.log('Message $DATE sent: ' + info.response);
    console.log('Verify on $CASE with $FROM');
});
END

done
