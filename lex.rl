package pifra

var parseError string

%%{ 
    machine parser;
    write data;
    access lex.;
    variable p lex.p;
    variable pe lex.pe;
}%%

type lexer struct {
    data []byte
    p, pe, cs int
    ts, te, act int
}

func newLexer(data []byte) *lexer {
    lex := &lexer{ 
        data: data,
        pe: len(data),
    }
    %% write init;
    return lex
}

func (lex *lexer) Lex(out *yySymType) int {
    eof := lex.pe
    tok := 0

    %%{ 
        main := |*
            '0' => { tok =  ZERO; fbreak; };
            '\'' => { tok =  APOSTROPHE; fbreak; };
            '[' => { tok =  LSQBRACKET; fbreak; };
            ']' => { tok =  RSQBRACKET; fbreak; };
            '$' => { tok = DOLLARSIGN; fbreak; };
            '+' => { tok = PLUS; fbreak; };
            '(' => { tok = LBRACKET; fbreak; };
            ')' => { tok = RBRACKET; fbreak; };
            '<' => { tok = LANGLE; fbreak; };
            '>' => { tok = RANGLE; fbreak; };
            ',' => { tok = COMMA; fbreak; };
            '!' => { tok = EXCLAMATION; fbreak; };
            '=' => { tok = EQUAL; fbreak; };
            '|' => { tok = VERTBAR; fbreak; };
            '.' => { tok = DOT; fbreak; };
            [_]?[a-zA-Z0-9]+ => { out.name = string(lex.data[lex.ts:lex.te]); tok = NAME; fbreak; };
            space;
        *|;
         write exec;
    }%%

    return tok;
}

func (lex *lexer) Error(err string) {
    parseError = err
}
