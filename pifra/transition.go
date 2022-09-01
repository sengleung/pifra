package pifra

import (
	"sort"
	"strconv"

	"github.com/mohae/deepcopy"
)

var maxStatesExplored = 1
var registerSize = 1073741824

type Configuration struct {
	Process   Element
	Registers Registers
	Label     Label
}

type SymbolType int

const (
	SymbolTypTau SymbolType = iota
	SymbolTypInput
	SymbolTypOutput
	SymbolTypFreshInput
	SymbolTypFreshOutput
	SymbolTypKnown
)

type Symbol struct {
	Type  SymbolType
	Value int
}

type Label struct {
	Symbol  Symbol
	Symbol2 Symbol
}

type Registers struct {
	Size      int
	Registers map[int]string
}

// UpdateMax adds a free name to the register at the register size + 1 and
// increments the register size.
// σ+v = σ U {(|σ|+1, v)}.
func (reg *Registers) UpdateMax(freeName string) int {
	reg.Size = reg.Size + 1
	reg.Registers[reg.Size] = freeName
	return reg.Size
}

// RemoveMax removes a free name from the register at the register size + 1
// and decrements the register size.
// Undos UpdateMax.
func (reg *Registers) RemoveMax() {
	delete(reg.Registers, reg.Size)
	reg.Size = reg.Size - 1
}

// AddEmptyName increments all labels by one while retaining mapping
// to their name and leaves an empty name (#) at label 1.
// #+o = {(1, #)} U {(i+1, v′) | (i, v′) E o}.
func (reg *Registers) AddEmptyName() {
	labels := reg.Labels()
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]
		reg.Registers[label+1] = reg.GetName(label)
	}
	delete(reg.Registers, 1)
}

// UpdateMin updates the register with a name at the minimum label
// where it does not exist in the set of free names.
func (reg *Registers) UpdateMin(name string, freshNames []string) int {
	freshNamesSet := make(map[string]bool)
	for _, freshName := range freshNames {
		freshNamesSet[freshName] = true
	}

	for label := 1; label <= reg.Size+1; label++ {
		if !freshNamesSet[reg.GetName(label)] {
			reg.Registers[label] = name
			return label
		}
	}
	return -1
}

// Labels returns register labels in sorted order.
func (reg *Registers) Labels() []int {
	var labels []int
	for k := range reg.Registers {
		labels = append(labels, k)
	}
	sort.Ints(labels)
	return labels
}

// GetName returns register name corresponding to the label.
func (reg *Registers) GetName(label int) string {
	if name, ok := reg.Registers[label]; ok {
		return name
	}
	return "NAME_NOT_FOUND"
}

// GetLabel returns register label corresponding to the name.
func (reg *Registers) GetLabel(name string) int {
	labels := reg.Labels()
	for _, label := range labels {
		n := reg.Registers[label]
		if n == name {
			return label
		}
	}
	return -1
}

func newRootConf(process Element) Configuration {
	fns := GetAllFreeNames(process)

	for _, dp := range DeclaredProcs {
		// Perform alpha conversion on the declared process
		// to determine scope.
		proc := deepcopy.Copy(dp.Process).(Element)
		bni := boundNameIndex
		DoAlphaConversion(proc)
		boundNameIndex = bni

		// Change the parameter names to bound so they are not
		// included in the free names.
		for _, oldName := range dp.Parameters {
			subName(proc, Name{
				Name: oldName,
			}, Name{
				Name: oldName,
				Type: Bound,
			})
		}

		// Gather free names in declared process.
		fns = append(fns, GetAllFreeNames(proc)...)
	}

	freshNamesSet := make(map[string]bool)
	for _, freshName := range fns {
		freshNamesSet[freshName] = true
	}

	var markedNames []string
	var freshNames []string
	for name := range freshNamesSet {
		if string(name[0]) == "_" {
			markedNames = append(markedNames, name)
		} else {
			freshNames = append(freshNames, name)
		}
	}
	sort.Strings(markedNames)
	sort.Strings(freshNames)

	// Place marked names ("_"-prefixed names) first in the register.
	register := make(map[int]string)
	regIndex := 1
	for _, name := range markedNames {
		register[regIndex] = name
		regIndex++
	}

	// Initialise the registers with generated free names.
	for i, name := range freshNames {
		fn := fnPrefix + strconv.Itoa(i+1)
		register[regIndex] = fn

		// Substitute the actual name with a generated free name.
		subName(process, Name{
			Name: name,
		}, Name{
			Name: fn,
		})

		for _, dp := range DeclaredProcs {
			// Change the parameter names to bound so they are not
			// substituted with the generated free name.
			for _, oldName := range dp.Parameters {
				subName(dp.Process, Name{
					Name: oldName,
				}, Name{
					Name: oldName,
					Type: Bound,
				})
			}

			// Substitute the actual name with a generated free name
			// in the process definition.
			subName(dp.Process, Name{
				Name: name,
			}, Name{
				Name: fn,
			})

			// Undo change of parameter names to bound so
			// unfolded processes are properly alpha-converted.
			for _, oldName := range dp.Parameters {
				subName(dp.Process, Name{
					Name: oldName,
					Type: Bound,
				}, Name{
					Name: oldName,
				})
			}
		}

		regIndex++
	}
	return Configuration{
		Process: process,
		Registers: Registers{
			Size:      registerSize,
			Registers: register,
		},
	}
}

var recVisitedProcs map[string]bool

func trans(conf Configuration) []Configuration {
	switch conf.Process.Type() {
	// DBLINP = INP1 + INP2A/INP2B
	case ElemTypInput:
		inp1Conf := conf
		inpElem := inp1Conf.Process.(*ElemInput)

		// Find the input channel label in the register.
		inpLabel := inp1Conf.Registers.GetLabel(inpElem.Channel.Name)
		inp1Conf.Label = Label{
			Symbol: Symbol{
				Type:  SymbolTypInput,
				Value: inpLabel,
			},
		}

		// INP2A
		var confs []Configuration
		for _, label := range inp1Conf.Registers.Labels() {
			inp2aConf := deepcopy.Copy(inp1Conf).(Configuration)
			inp2aElem := inp2aConf.Process.(*ElemInput)
			substituteName(inp2aElem, inp2aElem.Input, Name{
				Name: inp2aConf.Registers.GetName(label),
				Type: Free,
			})
			inp2aConf.Label.Symbol2 = Symbol{
				Type:  SymbolTypKnown,
				Value: label,
			}
			inp2aConf.Process = inp2aElem.Next
			confs = append(confs, inp2aConf)
		}

		// INP2B
		inp2bConf := inp1Conf
		inp2bElem := inp2bConf.Process.(*ElemInput)
		// Change the input bound name to a fresh name.
		substituteName(inp2bElem, inp2bElem.Input, Name{
			Name: inp2bElem.Input.Name,
			Type: Free,
		})

		name := inp2bElem.Input.Name
		freshNamesP := GetAllFreeNames(inp2bElem.Next)
		inp2bConf.Label.Symbol2 = Symbol{
			Type:  SymbolTypFreshInput,
			Value: inp2bConf.Registers.UpdateMin(name, freshNamesP),
		}
		inp2bConf.Process = inp2bElem.Next

		return append(confs, inp2bConf)

	// DBLOUT = OUT1 + OUT2
	case ElemTypOutput:
		out1Conf := conf
		outElem := out1Conf.Process.(*ElemOutput)

		outLabel := out1Conf.Registers.GetLabel(outElem.Channel.Name)
		out1Conf.Label = Label{
			Symbol: Symbol{
				Type:  SymbolTypOutput,
				Value: outLabel,
			},
		}

		// OUT2
		var confs []Configuration
		out2Conf := out1Conf
		out2Elem := out2Conf.Process.(*ElemOutput)

		label := out2Conf.Registers.GetLabel(out2Elem.Output.Name)

		out2Conf.Label.Symbol2 = Symbol{
			Type:  SymbolTypKnown,
			Value: label,
		}
		out2Conf.Process = out2Elem.Next
		confs = append(confs, out2Conf)
		return confs

	// MATCH
	case ElemTypMatch:
		var confs []Configuration

		matchElem := conf.Process.(*ElemEquality)
		// o ¦- [a=a]P
		if (!matchElem.Inequality && matchElem.NameL.Name == matchElem.NameR.Name) ||
			(matchElem.Inequality && matchElem.NameL.Name != matchElem.NameR.Name) {
			// o ¦- P
			matchConf := conf
			matchElem = matchConf.Process.(*ElemEquality)
			matchConf.Process = matchElem.Next
			// o ¦- P -t-> o ¦- P^'
			tconfs := trans(matchConf)
			// o ¦- P^'
			confs = append(confs, tconfs...)
		}

		return confs

	// RES, OPEN
	case ElemTypRestriction:
		var confs []Configuration

		// o |- $a.P^
		baseResConf := conf

		// RES
		// P^
		resConf := deepcopy.Copy(conf).(Configuration)
		resElem := resConf.Process.(*ElemRestriction)
		resName := resElem.Restrict.Name
		resConf.Process = resElem.Next
		// (o+a) ¦- P^
		resLabel := resConf.Registers.UpdateMax(resName)
		// (o+a) ¦- P^ -t-> (o'+a) ¦- P^' -t-> (o'+a) ¦- P^'
		tconfs := trans(resConf)
		// (o'+a) ¦- P^
		for _, conf := range tconfs {
			// OPEN
			if conf.Label.Symbol.Value != resLabel && conf.Label.Symbol2.Value != resLabel {
				// $a.P^'
				conf.Process = &ElemRestriction{
					Restrict: resElem.Restrict,
					Next:     conf.Process,
				}
				// o' ¦- $a.P^'
				conf.Registers.RemoveMax()

				// Convert the restriction free name to a bound name.
				subName(conf.Process, Name{
					Name: resName,
					Type: Free,
				}, Name{
					Name: resName,
					Type: Bound,
				})

				confs = append(confs, conf)
			}

			// RES
			if conf.Label.Symbol.Type == SymbolTypOutput &&
				conf.Label.Symbol2.Type == SymbolTypKnown &&
				conf.Label.Symbol.Value != resLabel &&
				conf.Label.Symbol2.Value == resLabel {
				// o
				conf.Registers = deepcopy.Copy(baseResConf.Registers).(Registers)
				// fn(P')
				freeNamesP := GetAllFreeNames(conf.Process)
				// o[j -> a], j = min{j | reg(j) !E fn(P')}
				label := conf.Registers.UpdateMin(resName, freeNamesP)
				// ij
				conf.Label.Symbol2.Value = label
				// ij^
				conf.Label.Symbol2.Type = SymbolTypFreshOutput

				// Substitute the bound name type to a fresh name type.
				subName(conf.Process, Name{
					Name: resName,
					Type: Bound,
				}, Name{
					Name: resName,
					Type: Free,
				})

				// o |- P'
				confs = append(confs, conf)
			}
		}

		return confs

	// REC
	case ElemTypProcess:
		procConf := conf
		procElem := procConf.Process.(*ElemProcess)

		processName := procElem.Name
		if _, ok := DeclaredProcs[processName]; !ok {
			return []Configuration{}
		}
		dp := DeclaredProcs[processName]
		if len(dp.Parameters) != len(procElem.Parameters) {
			return []Configuration{}
		}

		// P{a/b}
		proc := deepcopy.Copy(dp.Process).(Element)
		for i, oldName := range dp.Parameters {
			subName(proc, Name{
				Name: oldName,
			}, procElem.Parameters[i])
		}

		procConf.Process = proc
		doAlphaConversion(proc)

		// Create visited processes set.
		if recVisitedProcs == nil {
			recVisitedProcs = make(map[string]bool)
		}
		// Detects infinitely recursive processes such as P(a) = P(a).
		if recVisitedProcs[processName] {
			return []Configuration{}
		}
		recVisitedProcs[processName] = true
		tconfs := trans(procConf)
		recVisitedProcs = nil

		return tconfs

	// SUM
	case ElemTypSum:
		var confs []Configuration

		// SUM_L
		sumConf := deepcopy.Copy(conf).(Configuration)
		sumElem := sumConf.Process.(*ElemSum)
		sumConf.Process = sumElem.ProcessL
		lconfs := trans(sumConf)
		confs = append(confs, lconfs...)

		// SUM_R
		sumConf = deepcopy.Copy(conf).(Configuration)
		sumElem = sumConf.Process.(*ElemSum)
		sumConf.Process = sumElem.ProcessR
		rconfs := trans(sumConf)
		confs = append(confs, rconfs...)

		return confs

	// PAR1, PAR2, COMM, CLOSE
	case ElemTypParallel:
		var confs []Configuration
		var lconfs []Configuration
		var rconfs []Configuration
		basePar := conf

		// PAR1_L
		parConf := deepcopy.Copy(conf).(Configuration)
		parElem := parConf.Process.(*ElemParallel)
		parConf.Process = parElem.ProcessL
		tconfs := trans(parConf)

		// PAR2_L
		for _, conf := range tconfs {
			parConf = deepcopy.Copy(basePar).(Configuration)

			// When DBPINP/DBLOUT and the 2nd label is fresh input/fresh output.
			if conf.Label.Symbol2.Type == SymbolTypFreshInput ||
				conf.Label.Symbol2.Type == SymbolTypFreshOutput {
				// Find fn(P', Q).
				freeNamesP := GetAllFreeNames(conf.Process)
				freeNamesQ := GetAllFreeNames(parElem.ProcessR)
				// Get the name reg(i).
				name := conf.Registers.GetName(conf.Label.Symbol2.Value)
				// Update register to be j = min{j | reg(j) \notin fn(P′,Q)}.
				newLabel := parConf.Registers.UpdateMin(name,
					append(freeNamesP, freeNamesQ...))
				// Update the label j.
				parConf.Label = conf.Label
				parConf.Label.Symbol2.Value = newLabel
			} else {
				parConf.Label = conf.Label
				parConf.Registers = conf.Registers
			}
			// Insert P' to P' | Q.
			parConf.Process.(*ElemParallel).ProcessL = conf.Process

			lconfs = append(lconfs, parConf)
		}

		// PAR1_R
		parConf = deepcopy.Copy(conf).(Configuration)
		parElem = parConf.Process.(*ElemParallel)
		parConf.Process = parElem.ProcessR
		tconfs = trans(parConf)

		// PAR2_R
		for _, conf := range tconfs {
			parConf = deepcopy.Copy(basePar).(Configuration)
			// When DBPINP/DBLOUT and the 2nd label is fresh input/fresh output.
			if conf.Label.Symbol2.Type == SymbolTypFreshInput ||
				conf.Label.Symbol2.Type == SymbolTypFreshOutput {
				// Find fn(P, Q').
				freeNamesQ := GetAllFreeNames(conf.Process)
				freeNamesP := GetAllFreeNames(parElem.ProcessL)
				// Get the name reg(i).
				name := conf.Registers.GetName(conf.Label.Symbol2.Value)
				// Update register to be j = min{j | reg(j) \notin fn(P,Q')}.
				newLabel := parConf.Registers.UpdateMin(name,
					append(freeNamesP, freeNamesQ...))
				// Update the label j.
				parConf.Label = conf.Label
				parConf.Label.Symbol2.Value = newLabel
			} else {
				parConf.Label = conf.Label
				parConf.Registers = conf.Registers
			}
			// Insert Q' to P | Q'.
			parConf.Process.(*ElemParallel).ProcessR = conf.Process

			rconfs = append(rconfs, parConf)
		}

		confs = append(confs, append(lconfs, rconfs...)...)

		// COMM_L
		for _, lconf := range lconfs {
			for _, rconf := range rconfs {
				if lconf.Label.Symbol.Type == SymbolTypOutput &&
					lconf.Label.Symbol2.Type == SymbolTypKnown &&
					rconf.Label.Symbol.Type == SymbolTypInput &&
					rconf.Label.Symbol2.Type == SymbolTypKnown &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value &&
					lconf.Label.Symbol2.Value == rconf.Label.Symbol2.Value {
					lproc := deepcopy.Copy(lconf.Process).(Element).(*ElemParallel).ProcessL
					rproc := deepcopy.Copy(rconf.Process).(Element).(*ElemParallel).ProcessR
					comm := deepcopy.Copy(basePar).(Configuration)
					comm.Process = &ElemParallel{
						ProcessL: lproc,
						ProcessR: rproc,
					}
					comm.Label = Label{
						Symbol: Symbol{
							Type: SymbolTypTau,
						},
					}
					confs = append(confs, comm)
				}
			}
		}

		// COMM_R
		for _, lconf := range lconfs {
			for _, rconf := range rconfs {
				if lconf.Label.Symbol.Type == SymbolTypInput &&
					lconf.Label.Symbol2.Type == SymbolTypKnown &&
					rconf.Label.Symbol.Type == SymbolTypOutput &&
					rconf.Label.Symbol2.Type == SymbolTypKnown &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value &&
					lconf.Label.Symbol2.Value == rconf.Label.Symbol2.Value {
					lproc := deepcopy.Copy(lconf.Process).(Element).(*ElemParallel).ProcessL
					rproc := deepcopy.Copy(rconf.Process).(Element).(*ElemParallel).ProcessR
					comm := deepcopy.Copy(basePar).(Configuration)
					comm.Process = &ElemParallel{
						ProcessL: lproc,
						ProcessR: rproc,
					}
					comm.Label = Label{
						Symbol: Symbol{
							Type: SymbolTypTau,
						},
					}
					confs = append(confs, comm)
				}
			}
		}

		// CLOSE
		clconf := deepcopy.Copy(conf).(Configuration)
		// (#+o)
		clconf.Registers.AddEmptyName()
		parElem = clconf.Process.(*ElemParallel)
		// (#+o) ¦- P
		clconf.Process = parElem.ProcessL
		// -t-> (b+o) ¦- P'
		clconfs := trans(clconf)

		crconf := deepcopy.Copy(conf).(Configuration)
		// (#+o)
		crconf.Registers.AddEmptyName()
		parElem = crconf.Process.(*ElemParallel)
		// (#+o) ¦- Q
		crconf.Process = parElem.ProcessR
		// -t-> (b+o) ¦- Q'
		crconfs := trans(crconf)

		for _, lconf := range clconfs {
			for _, rconf := range crconfs {
				// CLOSE_L
				if lconf.Label.Symbol.Type == SymbolTypOutput &&
					lconf.Label.Symbol2.Type == SymbolTypFreshOutput &&
					lconf.Label.Symbol2.Value == 1 &&
					rconf.Label.Symbol.Type == SymbolTypInput &&
					rconf.Label.Symbol2.Type == SymbolTypFreshInput &&
					rconf.Label.Symbol2.Value == 1 &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value {
					{
						close := deepcopy.Copy(basePar).(Configuration)
						lproc := deepcopy.Copy(lconf.Process).(Element)
						rproc := deepcopy.Copy(rconf.Process).(Element)

						// Q'{a/b}
						resName := lconf.Registers.GetName(1)
						oldName := Name{
							Name: rconf.Registers.GetName(1),
							Type: Free,
						}
						newName := Name{
							Name: resName,
							Type: Bound,
						}
						substituteName(rproc, oldName, newName)

						// Convert restriction free name in P' to bound name.
						oldName = Name{
							Name: resName,
							Type: Free,
						}
						newName = Name{
							Name: resName,
							Type: Bound,
						}
						substituteName(lproc, oldName, newName)

						close.Process = &ElemRestriction{
							Restrict: Name{
								Name: resName,
								Type: Bound,
							},
							Next: &ElemParallel{
								ProcessL: lproc,
								ProcessR: rproc,
							},
						}
						close.Label = Label{
							Symbol: Symbol{
								Type: SymbolTypTau,
							},
						}
						confs = append(confs, close)
					}
				}
				// CLOSE_R
				if lconf.Label.Symbol.Type == SymbolTypInput &&
					lconf.Label.Symbol2.Type == SymbolTypFreshInput &&
					lconf.Label.Symbol2.Value == 1 &&
					rconf.Label.Symbol.Type == SymbolTypOutput &&
					rconf.Label.Symbol2.Type == SymbolTypFreshOutput &&
					rconf.Label.Symbol2.Value == 1 &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value {
					{
						close := deepcopy.Copy(basePar).(Configuration)
						lproc := deepcopy.Copy(lconf.Process).(Element)
						rproc := deepcopy.Copy(rconf.Process).(Element)

						// P'{a/b}
						resName := rconf.Registers.GetName(1)
						oldName := Name{
							Name: lconf.Registers.GetName(1),
							Type: Free,
						}
						newName := Name{
							Name: resName,
							Type: Bound,
						}
						substituteName(lproc, oldName, newName)

						// Convert restriction free name in Q' to bound name.
						oldName = Name{
							Name: resName,
							Type: Free,
						}
						newName = Name{
							Name: resName,
							Type: Bound,
						}
						substituteName(rproc, oldName, newName)

						close.Process = &ElemRestriction{
							Restrict: Name{
								Name: resName,
								Type: Bound,
							},
							Next: &ElemParallel{
								ProcessL: lproc,
								ProcessR: rproc,
							},
						}
						close.Label = Label{
							Symbol: Symbol{
								Type: SymbolTypTau,
							},
						}
						confs = append(confs, close)
					}
				}
			}
		}

		return confs

	case ElemTypRoot:
		rootConf := deepcopy.Copy(conf).(Configuration)
		rootConf.Process = rootConf.Process.(*ElemRoot).Next
		tconfs := trans(rootConf)
		// Reattach the root element.
		for i, conf := range tconfs {
			tconfs[i].Process = &ElemRoot{
				Next: conf.Process,
			}
		}
		return tconfs
	}
	return nil
}
