gen:
	API_ACCESS_TOKEN=$(shell aws --profile uneet-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text) node index.js
