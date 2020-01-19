package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
)

func TestParser(t *testing.T) {
	log = false
	tests := map[string]struct {
		input           []byte
		declaredProcs   map[string]Element
		undeclaredProcs []Element
		procParams      map[string][]string
		err             error
	}{
		"nil": {
			input: []byte(`
0
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemNil{},
			},
			procParams: map[string][]string{},
		},
		"output": {
			input: []byte(`
a'<b>.P
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemOutput{
					Name:     "a",
					Variable: "b",
					Next: &ElemProcess{
						Name: "P",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"input": {
			input: []byte(`
a(b).P
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemInput{
					Name:     "a",
					Variable: "b",
					Next: &ElemProcess{
						Name: "P",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"match": {
			input: []byte(`
[a=b]P
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemMatch{
					NameL: "a",
					NameR: "b",
					Next: &ElemProcess{
						Name: "P",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"restriction": {
			input: []byte(`
$a.P
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemRestriction{
					Name: "a",
					Next: &ElemProcess{
						Name: "P",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"sum": {
			input: []byte(`
P + Q
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemSum{
					ProcessL: &ElemProcess{
						Name: "P",
					},
					ProcessR: &ElemProcess{
						Name: "Q",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"parallel": {
			input: []byte(`
P | Q
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcess{
						Name: "P",
					},
					ProcessR: &ElemProcess{
						Name: "Q",
					},
				},
			},
			procParams: map[string][]string{},
		},
		"process": {
			input: []byte(`
P
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemProcess{
					Name: "P",
				},
			},
			procParams: map[string][]string{},
		},
		"process_constants": {
			input: []byte(`
P(a,b,c)
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemProcessConstants{
					Name:       "P",
					Parameters: []string{"c", "b", "a"},
				},
			},
			procParams: map[string][]string{},
		},
		"declared_process": {
			input: []byte(`
P = a'<b>.c(d).0
			`),
			declaredProcs: map[string]Element{
				"P": &ElemOutput{
					Name:     "a",
					Variable: "b",
					Next: &ElemInput{
						Name:     "c",
						Variable: "d",
						Next:     &ElemNil{},
					},
				},
			},
			undeclaredProcs: []Element{},
			procParams:      map[string][]string{},
		},
		"declared_process_constants": {
			input: []byte(`
Q(x,y,z) = $x.[y=z]P
			`),
			declaredProcs: map[string]Element{
				"Q": &ElemRestriction{
					Name: "x",
					Next: &ElemMatch{
						NameL: "y",
						NameR: "z",
						Next: &ElemProcess{
							Name: "P",
						},
					},
				},
			},
			undeclaredProcs: []Element{},
			procParams: map[string][]string{
				"Q": []string{"z", "y", "x"},
			},
		},
		"undecl_decl_processes": {
			input: []byte(`
P = a'<b>.c(d).0
Q(x,y,z) = $x.[y=z]P
i(j).k'<l>.0
			`),
			declaredProcs: map[string]Element{
				"P": &ElemOutput{
					Name:     "a",
					Variable: "b",
					Next: &ElemInput{
						Name:     "c",
						Variable: "d",
						Next:     &ElemNil{},
					},
				},
				"Q": &ElemRestriction{
					Name: "x",
					Next: &ElemMatch{
						NameL: "y",
						NameR: "z",
						Next: &ElemProcess{
							Name: "P",
						},
					},
				},
			},
			undeclaredProcs: []Element{
				&ElemInput{
					Name:     "i",
					Variable: "j",
					Next: &ElemOutput{
						Name:     "k",
						Variable: "l",
						Next:     &ElemNil{},
					},
				},
			},
			procParams: map[string][]string{
				"Q": []string{"z", "y", "x"},
			},
		},
		"processes_parallel": {
			input: []byte(`
R(i,j,k) = a(b).0 | (c'<d>.0 | e'<f>.0) | g(h).P(a,b,c,d) | i(j).Proc1
			`),
			declaredProcs: map[string]Element{
				"R": &ElemParallel{
					ProcessL: &ElemInput{
						Name:     "a",
						Variable: "b",
						Next:     &ElemNil{},
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemParallel{
							ProcessL: &ElemOutput{
								Name:     "c",
								Variable: "d",
								Next:     &ElemNil{},
							},
							ProcessR: &ElemOutput{
								Name:     "e",
								Variable: "f",
								Next:     &ElemNil{},
							},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemInput{
								Name:     "g",
								Variable: "h",
								Next: &ElemProcessConstants{
									Name:       "P",
									Parameters: []string{"d", "c", "b", "a"},
								},
							},
							ProcessR: &ElemInput{
								Name:     "i",
								Variable: "j",
								Next: &ElemProcess{
									Name: "Proc1",
								},
							},
						},
					},
				},
			},
			undeclaredProcs: []Element{},
			procParams: map[string][]string{
				"R": []string{"k", "j", "i"},
			},
		},
		"parallel_brackets": {
			input: []byte(`
((A | B) | (((C | D))) | E)
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemParallel{
						ProcessL: &ElemProcess{
							Name: "A",
						},
						ProcessR: &ElemProcess{
							Name: "B",
						},
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemParallel{
							ProcessL: &ElemProcess{
								Name: "C",
							},
							ProcessR: &ElemProcess{
								Name: "D",
							},
						},
						ProcessR: &ElemProcess{
							Name: "E",
						},
					},
				},
			},
			procParams: map[string][]string{},
		},
		"sum_parallel": {
			input: []byte(`
A | B + C | D | E + (F + G) + H
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcess{
						Name: "A",
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemSum{
							ProcessL: &ElemProcess{
								Name: "B",
							},
							ProcessR: &ElemProcess{
								Name: "C",
							},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemProcess{
								Name: "D",
							},
							ProcessR: &ElemSum{
								ProcessL: &ElemProcess{
									Name: "E",
								},
								ProcessR: &ElemSum{
									ProcessL: &ElemSum{
										ProcessL: &ElemProcess{
											Name: "F",
										},
										ProcessR: &ElemProcess{
											Name: "G",
										},
									},
									ProcessR: &ElemProcess{
										Name: "H",
									},
								},
							},
						},
					},
				},
			},
			procParams: map[string][]string{},
		},
		"sum_parallel_2": {
			input: []byte(`
A | B + C | D | E + (F + G | (P + R | Q)) + H
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcess{
						Name: "A",
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemSum{
							ProcessL: &ElemProcess{
								Name: "B",
							},
							ProcessR: &ElemProcess{
								Name: "C",
							},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemProcess{
								Name: "D",
							},
							ProcessR: &ElemSum{
								ProcessL: &ElemProcess{
									Name: "E",
								},
								ProcessR: &ElemSum{
									ProcessL: &ElemParallel{
										ProcessL: &ElemSum{
											ProcessL: &ElemProcess{
												Name: "F",
											},
											ProcessR: &ElemProcess{
												Name: "G",
											},
										},
										ProcessR: &ElemParallel{
											ProcessL: &ElemSum{
												ProcessL: &ElemProcess{
													Name: "P",
												},
												ProcessR: &ElemProcess{
													Name: "R",
												},
											},
											ProcessR: &ElemProcess{
												Name: "Q",
											},
										},
									},
									ProcessR: &ElemProcess{
										Name: "H",
									},
								},
							},
						},
					},
				},
			},
			procParams: map[string][]string{},
		},
		"parallel_restriction": {
			input: []byte(`
$a.b(a).$a.(b'<a>.0 | $b.(a(b).0 | c(d).0))
			`),
			declaredProcs: map[string]Element{},
			undeclaredProcs: []Element{
				&ElemRestriction{
					Name: "a",
					Next: &ElemInput{
						Name:     "b",
						Variable: "a",
						Next: &ElemRestriction{
							Name: "a",
							Next: &ElemParallel{
								ProcessL: &ElemOutput{
									Name:     "b",
									Variable: "a",
									Next:     &ElemNil{},
								},
								ProcessR: &ElemRestriction{
									Name: "b",
									Next: &ElemParallel{
										ProcessL: &ElemInput{
											Name:     "a",
											Variable: "b",
											Next:     &ElemNil{},
										},
										ProcessR: &ElemInput{
											Name:     "c",
											Variable: "d",
											Next:     &ElemNil{},
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
