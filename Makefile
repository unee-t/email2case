REGION:=us-west-2

dev:
	apex deploy -r $(REGION) --env dev

devlogs:
	apex logs ses -r $(REGION) --env dev

test:
	apex --env dev -r $(REGION) invoke ses < functions/ses/sns.json
