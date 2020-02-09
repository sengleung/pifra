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
		declaredProcs   map[string]DeclaredProcess
		undeclaredProcs []Element
		err             error
	}{
		"nil": {
			input: []byte(`
0
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemNil{},
			},
		},
		"output": {
			input: []byte(`
a'<b>.P
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemOutput{
					Channel: Name{
						Name: "a",
					},
					Output: Name{
						Name: "b",
					},
					Next: &ElemProcessConstants{
						Name: "P",
					},
				},
			},
		},
		"input": {
			input: []byte(`
a(b).P
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemInput{
					Channel: Name{
						Name: "a",
					},
					Input: Name{
						Name: "b",
					},
					Next: &ElemProcessConstants{
						Name: "P",
					},
				},
			},
		},
		"match": {
			input: []byte(`
[a=b]P
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemMatch{
					NameL: Name{
						Name: "a",
					},
					NameR: Name{
						Name: "b",
					},
					Next: &ElemProcessConstants{
						Name: "P",
					},
				},
			},
		},
		"restriction": {
			input: []byte(`
$a.P
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemRestriction{
					Restrict: Name{
						Name: "a",
					},
					Next: &ElemProcessConstants{
						Name: "P",
					},
				},
			},
		},
		"sum": {
			input: []byte(`
P + Q
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemSum{
					ProcessL: &ElemProcessConstants{
						Name: "P",
					},
					ProcessR: &ElemProcessConstants{
						Name: "Q",
					},
				},
			},
		},
		"parallel": {
			input: []byte(`
P | Q
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcessConstants{
						Name: "P",
					},
					ProcessR: &ElemProcessConstants{
						Name: "Q",
					},
				},
			},
		},
		"process": {
			input: []byte(`
P
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemProcessConstants{
					Name: "P",
				},
			},
		},
		"process_constants": {
			input: []byte(`
P(a,b,c)
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemProcessConstants{
					Name: "P",
					Parameters: []Name{
						{Name: "a"},
						{Name: "b"},
						{Name: "c"},
					},
				},
			},
		},
		"declared_process": {
			input: []byte(`
P = a'<b>.c(d).0
			`),
			declaredProcs: map[string]DeclaredProcess{
				"P": DeclaredProcess{
					Process: &ElemOutput{
						Channel: Name{
							Name: "a",
						},
						Output: Name{
							Name: "b",
						},
						Next: &ElemInput{
							Channel: Name{
								Name: "c",
							},
							Input: Name{
								Name: "d",
							},
							Next: &ElemNil{},
						},
					},
					Parameters: []string{},
				},
			},
			undeclaredProcs: []Element{},
		},
		"declared_process_constants": {
			input: []byte(`
Q(x,y,z) = $x.[y=z]P
			`),
			declaredProcs: map[string]DeclaredProcess{
				"Q": DeclaredProcess{
					Process: &ElemRestriction{
						Restrict: Name{
							Name: "x",
						},
						Next: &ElemMatch{
							NameL: Name{
								Name: "y",
							},
							NameR: Name{
								Name: "z",
							},
							Next: &ElemProcessConstants{
								Name: "P",
							},
						},
					},
					Parameters: []string{"x", "y", "z"},
				},
			},
			undeclaredProcs: []Element{},
		},
		"undecl_decl_processes": {
			input: []byte(`
P = a'<b>.c(d).0
Q(x,y,z) = $x.[y=z]P
i(j).k'<l>.0
			`),
			declaredProcs: map[string]DeclaredProcess{
				"P": DeclaredProcess{
					Process: &ElemOutput{
						Channel: Name{
							Name: "a",
						},
						Output: Name{
							Name: "b",
						},
						Next: &ElemInput{
							Channel: Name{
								Name: "c",
							},
							Input: Name{
								Name: "d",
							},
							Next: &ElemNil{},
						},
					},
					Parameters: []string{},
				},
				"Q": DeclaredProcess{
					Process: &ElemRestriction{
						Restrict: Name{
							Name: "x",
						},
						Next: &ElemMatch{
							NameL: Name{
								Name: "y",
							},
							NameR: Name{
								Name: "z",
							},
							Next: &ElemProcessConstants{
								Name: "P",
							},
						},
					},
					Parameters: []string{"x", "y", "z"},
				},
			},
			undeclaredProcs: []Element{
				&ElemInput{
					Channel: Name{
						Name: "i",
					},
					Input: Name{
						Name: "j",
					},
					Next: &ElemOutput{
						Channel: Name{
							Name: "k",
						},
						Output: Name{
							Name: "l",
						},
						Next: &ElemNil{},
					},
				},
			},
		},
		"processes_parallel": {
			input: []byte(`
R(i,j,k) = a(b).0 | (c'<d>.0 | e'<f>.0) | g(h).P(a,b,c,d) | i(j).Proc1
			`),
			declaredProcs: map[string]DeclaredProcess{
				"R": DeclaredProcess{
					Process: &ElemParallel{
						ProcessL: &ElemInput{
							Channel: Name{
								Name: "a",
							},
							Input: Name{
								Name: "b",
							},
							Next: &ElemNil{},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemParallel{
								ProcessL: &ElemOutput{
									Channel: Name{
										Name: "c",
									},
									Output: Name{
										Name: "d",
									},
									Next: &ElemNil{},
								},
								ProcessR: &ElemOutput{
									Channel: Name{
										Name: "e",
									},
									Output: Name{
										Name: "f",
									},
									Next: &ElemNil{},
								},
							},
							ProcessR: &ElemParallel{
								ProcessL: &ElemInput{
									Channel: Name{
										Name: "g",
									},
									Input: Name{
										Name: "h",
									},
									Next: &ElemProcessConstants{
										Name: "P",
										Parameters: []Name{
											{Name: "a"},
											{Name: "b"},
											{Name: "c"},
											{Name: "d"},
										},
									},
								},
								ProcessR: &ElemInput{
									Channel: Name{
										Name: "i",
									},
									Input: Name{
										Name: "j",
									},
									Next: &ElemProcessConstants{
										Name: "Proc1",
									},
								},
							},
						},
					},
					Parameters: []string{"i", "j", "k"},
				},
			},
			undeclaredProcs: []Element{},
		},
		"parallel_brackets": {
			input: []byte(`
((A | B) | (((C | D))) | E)
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemParallel{
						ProcessL: &ElemProcessConstants{
							Name: "A",
						},
						ProcessR: &ElemProcessConstants{
							Name: "B",
						},
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemParallel{
							ProcessL: &ElemProcessConstants{
								Name: "C",
							},
							ProcessR: &ElemProcessConstants{
								Name: "D",
							},
						},
						ProcessR: &ElemProcessConstants{
							Name: "E",
						},
					},
				},
			},
		},
		"sum_parallel": {
			input: []byte(`
A | B + C | D | E + (F + G) + H
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcessConstants{
						Name: "A",
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemSum{
							ProcessL: &ElemProcessConstants{
								Name: "B",
							},
							ProcessR: &ElemProcessConstants{
								Name: "C",
							},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemProcessConstants{
								Name: "D",
							},
							ProcessR: &ElemSum{
								ProcessL: &ElemProcessConstants{
									Name: "E",
								},
								ProcessR: &ElemSum{
									ProcessL: &ElemSum{
										ProcessL: &ElemProcessConstants{
											Name: "F",
										},
										ProcessR: &ElemProcessConstants{
											Name: "G",
										},
									},
									ProcessR: &ElemProcessConstants{
										Name: "H",
									},
								},
							},
						},
					},
				},
			},
		},
		"sum_parallel_2": {
			input: []byte(`
A | B + C | D | E + (F + G | (P + R | Q)) + H
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemParallel{
					ProcessL: &ElemProcessConstants{
						Name: "A",
					},
					ProcessR: &ElemParallel{
						ProcessL: &ElemSum{
							ProcessL: &ElemProcessConstants{
								Name: "B",
							},
							ProcessR: &ElemProcessConstants{
								Name: "C",
							},
						},
						ProcessR: &ElemParallel{
							ProcessL: &ElemProcessConstants{
								Name: "D",
							},
							ProcessR: &ElemSum{
								ProcessL: &ElemProcessConstants{
									Name: "E",
								},
								ProcessR: &ElemSum{
									ProcessL: &ElemParallel{
										ProcessL: &ElemSum{
											ProcessL: &ElemProcessConstants{
												Name: "F",
											},
											ProcessR: &ElemProcessConstants{
												Name: "G",
											},
										},
										ProcessR: &ElemParallel{
											ProcessL: &ElemSum{
												ProcessL: &ElemProcessConstants{
													Name: "P",
												},
												ProcessR: &ElemProcessConstants{
													Name: "R",
												},
											},
											ProcessR: &ElemProcessConstants{
												Name: "Q",
											},
										},
									},
									ProcessR: &ElemProcessConstants{
										Name: "H",
									},
								},
							},
						},
					},
				},
			},
		},
		"parallel_restriction": {
			input: []byte(`
$a.b(a).$a.(b'<a>.0 | $b.(a(b).0 | c(d).0))
			`),
			declaredProcs: map[string]DeclaredProcess{},
			undeclaredProcs: []Element{
				&ElemRestriction{
					Restrict: Name{
						Name: "a",
					},
					Next: &ElemInput{
						Channel: Name{
							Name: "b",
						},
						Input: Name{
							Name: "a",
						},
						Next: &ElemRestriction{
							Restrict: Name{
								Name: "a",
							},
							Next: &ElemParallel{
								ProcessL: &ElemOutput{
									Channel: Name{
										Name: "b",
									},
									Output: Name{
										Name: "a",
									},
									Next: &ElemNil{},
								},
								ProcessR: &ElemRestriction{
									Restrict: Name{
										Name: "b",
									},
									Next: &ElemParallel{
										ProcessL: &ElemInput{
											Channel: Name{
												Name: "a",
											},
											Input: Name{
												Name: "b",
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
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			initParser()
			lex := newLexer(tc.input)
			yyParse(lex)
			if err := deep.Equal(tc.declaredProcs, DeclaredProcs); err != nil {
				spew.Dump(DeclaredProcs, undeclaredProcs)
				t.Error(err)
			}
			if err := deep.Equal(tc.undeclaredProcs, undeclaredProcs); err != nil {
				spew.Dump(DeclaredProcs, undeclaredProcs)
				t.Error(err)
			}
		})
	}
}
