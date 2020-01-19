.PHONY: generate

generate:
	@rm -rf lex.go parser.go parser.output coverage.out coverage.html
	@ragel -Z -G2 -o lex.go lex.rl
	@goyacc -o parser.go -v parser.output parser.y

cover:
	@go test -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html


clean:
	@rm -rf lex.go parser.go parser.output coverage.out coverage.html
