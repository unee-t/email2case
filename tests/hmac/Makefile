all: dev demo prod

dev:
	export STAGE=dev; \
	export API_ACCESS_TOKEN=$(shell aws --profile uneet-dev ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text); \
	node index.js

demo:
	export STAGE=demo; \
	export API_ACCESS_TOKEN=$(shell aws --profile uneet-demo ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text); \
	node index.js

prod:
	export STAGE=""; \
	export API_ACCESS_TOKEN=$(shell aws --profile uneet-prod ssm get-parameters --names API_ACCESS_TOKEN --with-decryption --query Parameters[0].Value --output text); \
	node index.js
