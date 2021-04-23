COV = coverage.out

.PHONY: dep
dep:
	go mod tidy && go mod vendor

.PHONY: build
build:
	# TODO go build

.PHONY: client
client: 
	# gin \

.PHONY: server
server: 
	# TODO

.PHONY: test 
test: unittest gosec trufflehog

.PHONY: unittest
unittest:
	go test -v -race -coverprofile=$(COV) ./... \
		&& go tool cover -func $(COV)

.PHONY: gosec
gosec:
	gosec ./...

.PHONY: trufflehog
trufflehog:
	trufflehog3 .