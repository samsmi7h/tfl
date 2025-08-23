deps:
	brew install railway

auth:
	railway login

init:
	if [ -z "$$TFL_APP_KEY" ]; then echo "TFL_APP_KEY is not set"; exit 1; fi
	railway init
	railway environment
	railway add --service tfl
	railway variables -s tfl --set="TFL_APP_KEY=$$TFL_APP_KEY"

deploy:
	railway up

status:
	railway status

