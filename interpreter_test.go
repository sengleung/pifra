package pifra

import (
	"bytes"
	"reflect"
	"testing"
)

func TestTrans(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output []byte
	}{
		"input": {
			input: []byte(`
a(b).0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- 0
1 1* -> {(1,&b_0)} ¦- 0
`),
		},
		"inp2_name_in_P": {
			input: []byte(`
a(b).a(b).0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- #1(&b_1).0
1 2* -> {(1,#1),(2,&b_0)} ¦- #1(&b_1).0
`),
		},
		"output": {
			input: []byte(`
a'<b>.0
`),
			output: []byte(`
1'2  -> {(1,#1),(2,#2)} ¦- 0
`),
		},
		"open": {
			input: []byte(`
$a.b'<c>.0
`),
			output: []byte(`
1'2  -> {(1,#1),(2,#2)} ¦- $&a_0.0
`),
		},
		"open_name_in_P": {
			input: []byte(`
$a.b'<c>.b'<c>.0
`),
			output: []byte(`
1'2  -> {(1,#1),(2,#2)} ¦- $&a_0.#1'<#2>.0
`),
		},
		"multiple_inputs": {
			input: []byte(`
a(b).c(d).e(f).0
`),
			output: []byte(`
1 1  -> {(1,#1),(2,#2),(3,#3)} ¦- #2(&d_1).#3(&f_2).0
1 2  -> {(1,#1),(2,#2),(3,#3)} ¦- #2(&d_1).#3(&f_2).0
1 3  -> {(1,#1),(2,#2),(3,#3)} ¦- #2(&d_1).#3(&f_2).0
1 1* -> {(1,&b_0),(2,#2),(3,#3)} ¦- #2(&d_1).#3(&f_2).0
`),
		},
		"res_comm": {
			input: []byte(`
$a.b'<c>.0 | b(x).0
`),
			output: []byte(`
1'2  -> {(1,#1),(2,#2)} ¦- ($&a_0.0 | #1(&x_1).0)
1 1  -> {(1,#1),(2,#2)} ¦- ($&a_0.#1'<#2>.0 | 0)
1 2  -> {(1,#1),(2,#2)} ¦- ($&a_0.#1'<#2>.0 | 0)
1 3* -> {(1,#1),(2,#2),(3,&x_1)} ¦- ($&a_0.#1'<#2>.0 | 0)
t    -> {(1,#1),(2,#2)} ¦- ($&a_0.0 | 0)
`),
		},
		"close_left": {
			input: []byte(`
$a.b'<a>.0 | b(a).a'<a>.0
`),
			output: []byte(`
1'2^ -> {(1,#1),(2,&a_0)} ¦- (0 | #1(&a_1).&a_1'<&a_1>.0)
1 1  -> {(1,#1)} ¦- ($&a_0.#1'<&a_0>.0 | #1'<#1>.0)
1 2* -> {(1,#1),(2,&a_1)} ¦- ($&a_0.#1'<&a_0>.0 | &a_1'<&a_1>.0)
t    -> {(1,#1)} ¦- $&a_0.(0 | &a_0'<&a_0>.0)
`),
		},
		"close_right": {
			input: []byte(`
b(a).a'<a>.0 | $a.b'<a>.0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- (#1'<#1>.0 | $&a_1.#1'<&a_1>.0)
1 2* -> {(1,#1),(2,&a_0)} ¦- (&a_0'<&a_0>.0 | $&a_1.#1'<&a_1>.0)
1'2^ -> {(1,#1),(2,&a_1)} ¦- (#1(&a_0).&a_0'<&a_0>.0 | 0)
t    -> {(1,#1)} ¦- $&a_1.(&a_1'<&a_1>.0 | 0)
`),
		},
		"sum": {
			input: []byte(`
a(b).b<b>.0 + a(b).a<b>.0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- #1'<#1>.0
1 1* -> {(1,&b_0)} ¦- &b_0'<&b_0>.0
1 1  -> {(1,#1)} ¦- #1'<#1>.0
1 2* -> {(1,#1),(2,&b_1)} ¦- #1'<&b_1>.0
`),
		},
		"match": {
			input: []byte(`
[a=a]a(b).0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- 0
1 1* -> {(1,&b_0)} ¦- 0
`),
		},
		"no_match": {
			input: []byte(`
[a=b]a(b).0
`),
			output: []byte(`
`),
		},
		"rec": {
			input: []byte(`
P(b) = b'<b>.0
			
a(b).P(b)
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- P(#1)
1 1* -> {(1,&b_0)} ¦- P(&b_0)
`),
		},
		"par2_inp": {
			input: []byte(`
a(b).0 | 0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- (0 | 0)
1 1* -> {(1,&b_0)} ¦- (0 | 0)
`),
		},
		"par2_inp_name_in_Q": {
			input: []byte(`
a(b).0 | a<a>.0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- (0 | #1'<#1>.0)
1 2* -> {(1,#1),(2,&b_0)} ¦- (0 | #1'<#1>.0)
1'1  -> {(1,#1)} ¦- (#1(&b_0).0 | 0)
t    -> {(1,#1)} ¦- (0 | 0)
`),
		},
		"par2_sym_inp": {
			input: []byte(`
0 | a(b).0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- (0 | 0)
1 1* -> {(1,&b_0)} ¦- (0 | 0)
`),
		},
		"par2_sym_inp_name_in_P": {
			input: []byte(`
a<a>.0 | a(b).0
`),
			output: []byte(`
1'1  -> {(1,#1)} ¦- (0 | #1(&b_0).0)
1 1  -> {(1,#1)} ¦- (#1'<#1>.0 | 0)
1 2* -> {(1,#1),(2,&b_0)} ¦- (#1'<#1>.0 | 0)
t    -> {(1,#1)} ¦- (0 | 0)
`),
		},
		"par2_out": {
			input: []byte(`
$x.a<a>.0 | 0
`),
			output: []byte(`
1'1  -> {(1,#1)} ¦- ($&x_0.0 | 0)
`),
		},
		"par2_sym_out": {
			input: []byte(`
0 | $x.a<a>.0
`),
			output: []byte(`
1'1  -> {(1,#1)} ¦- (0 | $&x_0.0)
`),
		},
		"par2_out_name_in_Q": {
			input: []byte(`
a(b).0 | $x.a<a>.0
`),
			output: []byte(`
1 1  -> {(1,#1)} ¦- (0 | $&x_1.#1'<#1>.0)
1 2* -> {(1,#1),(2,&b_0)} ¦- (0 | $&x_1.#1'<#1>.0)
1'1  -> {(1,#1)} ¦- (#1(&b_0).0 | $&x_1.0)
t    -> {(1,#1)} ¦- (0 | $&x_1.0)
`),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			proc, _ := InitProgram(tc.input)
			root := newRootConf(proc)
			confs := trans(root)
			var output bytes.Buffer
			output.WriteString("\n")
			for _, conf := range confs {
				output.WriteString(PrettyPrintConfiguration(conf) + "\n")
			}
			if !reflect.DeepEqual(tc.output, output.Bytes()) {
				t.Error(name)
			}
		})
	}
}
