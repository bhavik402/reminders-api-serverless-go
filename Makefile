SHELL=/bin/bash

test: 
	go test -v ./...

build: 
	go build -v ./...

format: 
	gofmt -w */*/*.go

clean: 
	if [ -f bootstrap ]; then rm bootstrap; fi

build-executable: clean
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap cmd/main/main.go

artifact: build-executable
	zip RemindersFunction.zip bootstrap

cdk-synth: artifact 
	export REMINDERS_ENV=dev && cdk synth

cdk-bootstrap: artifact 
	export REMINDERS_ENV=dev && cdk bootstrap

cdk-deploy: cdk-bootstrap
	export REMINDERS_ENV=dev && cdk deploy "RemindersAPIStack"

cdk-destroy: artifact
	export REMINDERS_ENV=dev && cdk destroy