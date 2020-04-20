
//line lex.rl:1
package pifra

var parseError string


//line lex.go:9
const parser_start int = 2
const parser_first_final int = 2
const parser_error int = 0

const parser_en_main int = 2


//line lex.rl:11


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
    
//line lex.go:32
	{
	 lex.cs = parser_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line lex.rl:25
    return lex
}

func (lex *lexer) Lex(out *yySymType) int {
    eof := lex.pe
    tok := 0

    
//line lex.go:49
	{
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
	switch  lex.cs {
	case 2:
		goto st_case_2
	case 0:
		goto st_case_0
	case 3:
		goto st_case_3
	case 1:
		goto st_case_1
	}
	goto st_out
tr2:
//line lex.rl:50
 lex.te = ( lex.p)+1

	goto st2
tr3:
//line lex.rl:45
 lex.te = ( lex.p)+1
{ tok = EXCLAMATION; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr4:
//line lex.rl:38
 lex.te = ( lex.p)+1
{ tok = DOLLARSIGN; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr5:
//line lex.rl:35
 lex.te = ( lex.p)+1
{ tok =  APOSTROPHE; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr6:
//line lex.rl:40
 lex.te = ( lex.p)+1
{ tok = LBRACKET; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr7:
//line lex.rl:41
 lex.te = ( lex.p)+1
{ tok = RBRACKET; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr8:
//line lex.rl:39
 lex.te = ( lex.p)+1
{ tok = PLUS; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr9:
//line lex.rl:44
 lex.te = ( lex.p)+1
{ tok = COMMA; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr10:
//line lex.rl:48
 lex.te = ( lex.p)+1
{ tok = DOT; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr12:
//line lex.rl:42
 lex.te = ( lex.p)+1
{ tok = LANGLE; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr13:
//line lex.rl:46
 lex.te = ( lex.p)+1
{ tok = EQUAL; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr14:
//line lex.rl:43
 lex.te = ( lex.p)+1
{ tok = RANGLE; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr15:
//line lex.rl:36
 lex.te = ( lex.p)+1
{ tok =  LSQBRACKET; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr16:
//line lex.rl:37
 lex.te = ( lex.p)+1
{ tok =  RSQBRACKET; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr18:
//line lex.rl:47
 lex.te = ( lex.p)+1
{ tok = VERTBAR; {( lex.p)++;  lex.cs = 2; goto _out } }
	goto st2
tr19:
//line NONE:1
	switch  lex.act {
	case 1:
	{( lex.p) = ( lex.te) - 1
 tok =  ZERO; {( lex.p)++;  lex.cs = 2; goto _out } }
	case 16:
	{( lex.p) = ( lex.te) - 1
 out.name = string(lex.data[lex.ts:lex.te]); tok = NAME; {( lex.p)++;  lex.cs = 2; goto _out } }
	}
	
	goto st2
	st2:
//line NONE:1
 lex.ts = 0

		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof2
		}
	st_case_2:
//line NONE:1
 lex.ts = ( lex.p)

//line lex.go:163
		switch  lex.data[( lex.p)] {
		case 32:
			goto tr2
		case 33:
			goto tr3
		case 36:
			goto tr4
		case 39:
			goto tr5
		case 40:
			goto tr6
		case 41:
			goto tr7
		case 43:
			goto tr8
		case 44:
			goto tr9
		case 46:
			goto tr10
		case 48:
			goto tr11
		case 60:
			goto tr12
		case 61:
			goto tr13
		case 62:
			goto tr14
		case 91:
			goto tr15
		case 93:
			goto tr16
		case 95:
			goto st1
		case 124:
			goto tr18
		}
		switch {
		case  lex.data[( lex.p)] < 49:
			if 9 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 13 {
				goto tr2
			}
		case  lex.data[( lex.p)] > 57:
			switch {
			case  lex.data[( lex.p)] > 90:
				if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
					goto tr0
				}
			case  lex.data[( lex.p)] >= 65:
				goto tr0
			}
		default:
			goto tr0
		}
		goto st0
st_case_0:
	st0:
		 lex.cs = 0
		goto _out
tr0:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:49
 lex.act = 16;
	goto st3
tr11:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:34
 lex.act = 1;
	goto st3
	st3:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof3
		}
	st_case_3:
//line lex.go:241
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr0
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr0
			}
		default:
			goto tr0
		}
		goto tr19
	st1:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof1
		}
	st_case_1:
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr0
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr0
			}
		default:
			goto tr0
		}
		goto st0
	st_out:
	_test_eof2:  lex.cs = 2; goto _test_eof
	_test_eof3:  lex.cs = 3; goto _test_eof
	_test_eof1:  lex.cs = 1; goto _test_eof

	_test_eof: {}
	if ( lex.p) == eof {
		switch  lex.cs {
		case 3:
			goto tr19
		}
	}

	_out: {}
	}

//line lex.rl:53


    return tok;
}

func (lex *lexer) Error(err string) {
    parseError = err
}
