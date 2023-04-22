build:
	docker build -t ytrip .
run:
	docker run --rm --name ytrip -de TG_BOT_TOKEN=${TG_BOT_TOKEN} ytrip
stop:
	docker stop ytrip
clean:
	docker rm ytrip
logs:
	docker logs -f ytrip
