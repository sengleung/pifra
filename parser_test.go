package pifra

import (
	"reflect"
	"testing"
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
					Next: &ElemProcess{
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
					Next: &ElemProcess{
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
				&ElemEquality{
					NameL: Name{
						Name: "a",
					},
					NameR: Name{
						Name: "b",
					},
					Next: &ElemProcess{
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
					Next: &ElemProcess{
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
					ProcessL: &ElemProcess{
						Name: "P",
					},
					ProcessR: &ElemProcess{
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
					ProcessL: &ElemProcess{
						Name: "P",
					},
					ProcessR: &ElemProcess{
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
				&ElemProcess{
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
				&ElemProcess{
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
						Next: &ElemEquality{
							NameL: Name{
								Name: "y",
							},
							NameR: Name{
								Name: "z",
							},
							Next: &ElemProcess{
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
						Next: &ElemEquality{
							NameL: Name{
								Name: "y",
							},
							NameR: Name{
								Name: "z",
							},
							Next: &ElemProcess{
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
									Next: &ElemProcess{
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
									Next: &ElemProcess{
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
		},
		"sum_parallel": {
			input: []byte(`
A | B + C | D | E + (F + G) + H
			`),
			declaredProcs: map[string]DeclaredProcess{},
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
		},
		"sum_parallel_2": {
			input: []byte(`
A | B + C | D | E + (F + G | (P + R | Q)) + H
			`),
			declaredProcs: map[string]DeclaredProcess{},
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
			if !reflect.DeepEqual(tc.declaredProcs, DeclaredProcs) {
				t.Error(name)
			}
			if !reflect.DeepEqual(tc.undeclaredProcs, undeclaredProcs) {
				t.Error(name)
			}
		})
	}
}
