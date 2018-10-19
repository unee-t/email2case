To get summaries of emails sent to reply@\*.unee-t.com * test@\*.unee-t.com, subscribe to arn:aws:sns:us-west-2:\*:incomingreply SNS.

# AWS SES receving setup in dev account

<img src=https://media.dev.unee-t.com/2018-09-13/reply.png>

# Testing

Please refer to `test/setup-mail-to.sh` which generates JS files to run for each account.

# Deploy

	make {dev,demo,prod}
