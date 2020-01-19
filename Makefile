.PHONY: generate

generate:
	@rm -rf lex.go parser.go parser.output
	@ragel -Z -G2 -o lex.go lex.rl
	@goyacc -o parser.go -v parser.output parser.y

clean:
	@rm -rf lex.go parser.go parser.output
