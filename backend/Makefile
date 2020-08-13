.PHONY: build clean deploy gomodgen

FUNCTION_NAME ?= schedule

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/refresh functions/refresh/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/requestRefresh functions/requestRefresh/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/schedule functions/schedule/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login functions/login/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/click functions/click/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	npx sls deploy --verbose

logs:
	npx sls logs -f $(FUNCTION_NAME) -t

invoke:
	npx sls invoke -f $(FUNCTION_NAME) -l --path functions/$(FUNCTION_NAME)/resources/request.json

invoke-local: build
	./invoke.sh $(FUNCTION_NAME)