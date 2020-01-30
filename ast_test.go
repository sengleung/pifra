package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
)

func TestAlphaConversion(t *testing.T) {
	log = false
	tests := map[string]struct {
		input           []byte
		declaredProcs   map[string]Element
		undeclaredProcs []Element
		procParams      map[string][]string
		err             error
	}{
		"parallel_restriction": {
			input: []byte(`
$a.b(a).$a.(b'<a>.0 | $b.(a(b).0 | c(d).0))
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemRestriction{
					Name: "bn_0",
					Next: &ElemInput{
						Channel: Name{
							Name: "b",
						},
						Input: Name{
							Name: "bn_0",
						},
						Next: &ElemRestriction{
							Name: "bn_1",
							Next: &ElemParallel{
								ProcessL: &ElemOutput{
									Channel: Name{
										Name: "b",
									},
									Output: Name{
										Name: "bn_1",
									},
									Next: &ElemNil{},
								},
								ProcessR: &ElemRestriction{
									Name: "bn_2",
									Next: &ElemParallel{
										ProcessL: &ElemInput{
											Channel: Name{
												Name: "bn_1",
											},
											Input: Name{
												Name: "bn_2",
											},
											Next: &ElemNil{},
										},
										ProcessR: &ElemInput{
											Channel: Name{
												Name: "c",
											},
											Input: Name{
												Name: "d",
											},
											Next: &ElemNil{},
										},
									},
								},
							},
						},
					},
				},
			},
			procParams: map[string][]string{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			initParser()
			lex := newLexer(tc.input)
			yyParse(lex)
			for _, elem := range declaredProcs {
				doAlphaConversion(elem)
			}
			for _, elem := range undeclaredProcs {
				doAlphaConversion(elem)
			}
			if err := deep.Equal(tc.declaredProcs, declaredProcs); err != nil {
				spew.Dump(declaredProcs, undeclaredProcs, procParams)
				t.Error(err)
			}
			if err := deep.Equal(tc.procParams, procParams); err != nil {
				spew.Dump(declaredProcs, undeclaredProcs, procParams)
				t.Error(err)
			}
			if err := deep.Equal(tc.undeclaredProcs, undeclaredProcs); err != nil {
				spew.Dump(declaredProcs, undeclaredProcs, procParams)
				t.Error(err)
			}
		})
	}
}
