.PHONY: generate

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

generate:
	@rm -rf pifra/lex.go pifra/parser.go pifra/parser.output coverage.out coverage.html
	@cd pifra && ragel -Z -G2 -o lex.go lex.rl
	@cd pifra && goyacc -o parser.go -v parser.output parser.y
	@go build -o $$GOPATH/bin/pifra cmd/pifra/main.go

cover:
	@go test ./pifra -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

clean:
	@rm -rf pifra/lex.go pifra/parser.go pifra/parser.output coverage.out coverage.html
