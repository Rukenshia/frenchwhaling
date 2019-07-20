.PHONY: build clean deploy gomodgen

FUNCTION_NAME ?= refresh

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/refresh functions/refresh/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login functions/login/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose

logs:
	sls logs -f $(FUNCTION_NAME) -t

invoke: build
	./invoke.sh $(FUNCTION_NAME)