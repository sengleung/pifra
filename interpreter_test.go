package pifra

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
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
1 1  -> {(1,a)} ¦- 0
1 1* -> {(1,b_0)} ¦- 0
`),
		},
		"output": {
			input: []byte(`
a'<b>.0
`),
			output: []byte(`
1'1  -> {(1,a),(2,b)} ¦- 0
1'2  -> {(1,a),(2,b)} ¦- 0
`),
		},
		"restriction": {
			input: []byte(`
$a.b'<c>.0
`),
			output: []byte(`
1'1^ -> {(1,a_0),(2,c)} ¦- 0
1'1^ -> {(1,a_0),(2,c)} ¦- 0
`),
		},
		"multiple_inputs": {
			input: []byte(`
a(b).c(d).e(f).0
`),
			output: []byte(`
1 1  -> {(1,a),(2,c),(3,e)} ¦- c(d_1).e(f_2).0
1 2  -> {(1,a),(2,c),(3,e)} ¦- c(d_1).e(f_2).0
1 3  -> {(1,a),(2,c),(3,e)} ¦- c(d_1).e(f_2).0
1 1* -> {(1,b_0),(2,c),(3,e)} ¦- c(d_1).e(f_2).0
`),
		},
		"res_comm": {
			input: []byte(`
$a.b'<c>.0 | b(x).0
`),
			output: []byte(`
1'2^ -> {(1,b),(2,a_0)} ¦- (0 | b(x_1).0)
1'2^ -> {(1,b),(2,a_0)} ¦- (0 | b(x_1).0)
1 1  -> {(1,b),(2,c)} ¦- ($a_0.b'<c>.0 | 0)
1 2  -> {(1,b),(2,c)} ¦- ($a_0.b'<c>.0 | 0)
1 3* -> {(1,b),(2,c),(3,x_1)} ¦- ($a_0.b'<c>.0 | 0)
t    -> {(1,b),(2,c)} ¦- $a_0.(0 | 0)
t    -> {(1,b),(2,c)} ¦- $a_0.(0 | 0)
t    -> {(1,b),(2,c)} ¦- $a_0.(0 | 0)
`),
		},
		"close_left": {
			input: []byte(`
$a.b'<a>.0 | b(a).a'<a>.0
`),
			output: []byte(`
1'2^ -> {(1,b),(2,a_0)} ¦- (0 | b(a_1).a_1'<a_1>.0)
1 1  -> {(1,b)} ¦- ($a_0.b'<a_0>.0 | b'<b>.0)
1 2* -> {(1,b),(2,a_1)} ¦- ($a_0.b'<a_0>.0 | a_1'<a_1>.0)
t    -> {(1,b)} ¦- $a_0.(0 | a_0'<a_0>.0)
t    -> {(1,b)} ¦- $a_0.(0 | a_0'<a_0>.0)
`),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			proc, _ := InitProgram(tc.input)
			root := newTransitionStateRoot(proc)
			confs := trans(root.Configuration)
			var output bytes.Buffer
			output.WriteString("\n")
			for _, conf := range confs {
				output.WriteString(PrettyPrintConfiguration(conf) + "\n")
			}
			if err := deep.Equal(tc.output, output.Bytes()); err != nil {
				t.Error(err)
			}
		})
	}
}
