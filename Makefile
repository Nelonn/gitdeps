
all: test build

build:
	go build gitdeps.go

test:
	go test gitdeps.go gitdeps_test.go

run: build
	./gitdeps
