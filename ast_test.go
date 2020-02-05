package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
)

func TestSubstituteName(t *testing.T) {
	log = false
	tests := map[string]struct {
		input   Element
		output  Element
		oldName Name
		newName Name
		err     error
	}{
		"output_input_bound_name": {
			input: &ElemOutput{
				Channel: Name{
					Name: "a",
				},
				Output: Name{
					Name: "b",
				},
				Next: &ElemInput{
					Channel: Name{
						Name: "a",
					},
					Input: Name{
						Name: "d",
					},
					Next: &ElemNil{},
				},
			},
			output: &ElemOutput{
				Channel: Name{
					Name: "b",
					Type: Bound,
				},
				Output: Name{
					Name: "b",
				},
				Next: &ElemInput{
					Channel: Name{
						Name: "b",
						Type: Bound,
					},
					Input: Name{
						Name: "d",
					},
					Next: &ElemNil{},
				},
			},
			oldName: Name{
				Name: "a",
			},
			newName: Name{
				Name: "b",
				Type: Bound,
			},
		},
		"par_match_free_name": {
			input: &ElemOutput{
				Channel: Name{
					Name: "a",
					Type: Bound,
				},
				Output: Name{
					Name: "b",
				},
				Next: &ElemParallel{
					ProcessL: &ElemInput{
						Channel: Name{
							Name: "a",
							Type: Bound,
						},
						Input: Name{
							Name: "d",
						},
						Next: &ElemNil{},
					},
					ProcessR: &ElemMatch{
						NameL: Name{
							Name: "a",
							Type: Bound,
						},
						NameR: Name{
							Name: "e",
						},
						Next: &ElemNil{},
					},
				},
			},
			output: &ElemOutput{
				Channel: Name{
					Name: "b",
				},
				Output: Name{
					Name: "b",
				},
				Next: &ElemParallel{
					ProcessL: &ElemInput{
						Channel: Name{
							Name: "b",
						},
						Input: Name{
							Name: "d",
						},
						Next: &ElemNil{},
					},
					ProcessR: &ElemMatch{
						NameL: Name{
							Name: "b",
						},
						NameR: Name{
							Name: "e",
						},
						Next: &ElemNil{},
					},
				},
			},
			oldName: Name{
				Name: "a",
				Type: Bound,
			},
			newName: Name{
				Name: "b",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			substituteName(tc.input, tc.oldName, tc.newName)
			if err := deep.Equal(tc.input, tc.output); err != nil {
				spew.Dump(tc.input)
				t.Error(err)
			}
		})
	}
}

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
				DoAlphaConversion(elem)
			}
			for _, elem := range undeclaredProcs {
				DoAlphaConversion(elem)
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
