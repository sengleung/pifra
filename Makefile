.PHONY: generate

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

generate:
	@rm -rf lex.go parser.go parser.output coverage.out coverage.html
	@ragel -Z -G2 -o lex.go lex.rl
	@goyacc -o parser.go -v parser.output parser.y
	@go build -o $$GOPATH/bin/pifra pifra/main.go

cover:
	@go test -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

clean:
	@rm -rf lex.go parser.go parser.output coverage.out coverage.html
