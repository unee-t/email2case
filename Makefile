REGION:=us-west-2

demo:
	apex deploy -r $(REGION) --env demo

demologs:
	apex logs ses -r $(REGION) --env demo

prod:
	apex deploy -r ap-southeast-1 --env prod

prodlogs:
	apex logs ses -r ap-southeast-1 --env prod

dev:
	apex deploy -r $(REGION) --env dev

devlogs:
	apex logs ses -r $(REGION) --env dev

test:
	apex --env dev -r $(REGION) invoke ses < functions/ses/sns.json

testprod:
	apex --env prod -r ap-southeast-1 invoke ses < functions/ses/sns.json
