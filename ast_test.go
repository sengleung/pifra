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
a(b).$a.b(a).$a.(b'<a>.0 | $b.(a(b).0 | c(d).0))
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemInput{
					Channel: Name{
						Name: "a",
					},
					Input: Name{
						Name: "b_0",
						Type: Bound,
					},
					Next: &ElemRestriction{
						Restrict: Name{
							Name: "a_1",
							Type: Bound,
						},
						Next: &ElemInput{
							Channel: Name{
								Name: "b_0",
								Type: Bound,
							},
							Input: Name{
								Name: "a_1",
								Type: Bound,
							},
							Next: &ElemRestriction{
								Restrict: Name{
									Name: "a_2",
									Type: Bound,
								},
								Next: &ElemParallel{
									ProcessL: &ElemOutput{
										Channel: Name{
											Name: "b_0",
											Type: Bound,
										},
										Output: Name{
											Name: "a_2",
											Type: Bound,
										},
										Next: &ElemNil{},
									},
									ProcessR: &ElemRestriction{
										Restrict: Name{
											Name: "b_3",
											Type: Bound,
										},
										Next: &ElemParallel{
											ProcessL: &ElemInput{
												Channel: Name{
													Name: "a_2",
													Type: Bound,
												},
												Input: Name{
													Name: "b_3",
													Type: Bound,
												},
												Next: &ElemNil{},
											},
											ProcessR: &ElemInput{
												Channel: Name{
													Name: "c",
												},
												Input: Name{
													Name: "d_4",
													Type: Bound,
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
