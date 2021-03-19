.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/get/article get/article/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get/tag get/tag/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/post/article post/article/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/post/tag post/tag/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
