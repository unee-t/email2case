deploy:
	apex deploy -r us-west-2

logs:
	apex logs ses -r us-west-2

test:
	apex -r us-west-2 invoke ses < functions/ses/sns.json
