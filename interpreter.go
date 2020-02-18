package pifra

import (
	"container/list"
	"fmt"
	"sort"
	"strconv"

	"github.com/mohae/deepcopy"
)

var maxStatesExplored = 1

type State struct {
	Configuration Configuration
	NextStates    []*State
}

type Configuration struct {
	Process  Element
	Register Register
	Label    Label
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
	Double  bool
	Symbol  Symbol
	Symbol2 Symbol
}

type Register struct {
	Index    int
	Register map[int]string
}

// UpdateAfter adds a free name to the register at the next label.
// reg+v = reg U {(|reg|+1, v)}.
func (reg *Register) UpdateAfter(freeName string) int {
	index := reg.Index
	reg.Register[index] = freeName
	reg.Index = reg.Index + 1
	return index
}

// RemoveLastLabel removes the last label from the register.
// Undos UpdateAfter, but retains the modified registers.
func (reg *Register) RemoveLastLabel() {
	reg.Index = reg.Index - 1
	delete(reg.Register, reg.Index)
}

// AddEmptyName increments all labels by one while retaining mapping
// to their name and leaves an empty name (#) at label 1.
// #+o = {(1, #)} U {(i+1, v′) | (i, v′) E o}.
func (reg *Register) AddEmptyName() {
	labels := reg.Labels()
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]
		reg.Register[label+1] = reg.GetName(label)
	}
	reg.Register[1] = "#"
	reg.Index = reg.Index + 1
}

// Labels returns register labels in sorted order.
func (reg *Register) Labels() []int {
	var labels []int
	for k := range reg.Register {
		labels = append(labels, k)
	}
	sort.Ints(labels)
	return labels
}

// UpdateMin updates the register with a name at the minimum label
// where it does not exist in the set of free names.
func (reg *Register) UpdateMin(name string, freshNames []string) int {
	freshNamesSet := make(map[string]bool)
	for _, freshName := range freshNames {
		freshNamesSet[freshName] = true
	}
	labels := reg.Labels()
	for _, label := range labels {
		if reg.Register[label] == "#" || !freshNamesSet[reg.GetName(label)] {
			reg.Register[label] = name
			return label
		}
	}
	label := reg.Index
	reg.Register[label] = name
	reg.Index = reg.Index + 1
	return label
}

// GetName returns register name corresponding to the label.
func (reg *Register) GetName(label int) string {
	if name, ok := reg.Register[label]; ok {
		return name
	}
	return "NAME_NOT_FOUND"
}

// GetLabel returns register label corresponding to the name.
func (reg *Register) GetLabel(name string) int {
	labels := reg.Labels()
	for _, label := range labels {
		n := reg.Register[label]
		if n == name {
			return label
		}
	}
	return -1
}

func newTransitionStateRoot(process Element) *State {
	fns := GetAllFreshNames(process)
	freshNamesSet := make(map[string]bool)

	for _, freshName := range fns {
		freshNamesSet[freshName] = true
	}

	var freshNames []string
	for name := range freshNamesSet {
		freshNames = append(freshNames, name)
	}
	sort.Strings(freshNames)

	register := make(map[int]string)
	index := 1
	for _, name := range freshNames {
		register[index] = name
		index = index + 1
	}
	return &State{
		Configuration: Configuration{
			Process: process,
			Register: Register{
				Index:    len(register) + 1,
				Register: register,
			},
		},
		NextStates: []*State{},
	}
}

var infProc bool

func exploreTransitions(root *State) (map[int]Configuration, []GraphEdge) {
	queue := list.New()
	queue.PushBack(root)
	dequeue := func() *State {
		s := queue.Front()
		queue.Remove(s)
		return s.Value.(*State)
	}

	visited := make(map[string]int)
	vertices := make(map[int]Configuration)
	var edges []GraphEdge
	var vertexId int

	// BFS traversal state exploration.
	var statesExplored int
	for queue.Len() > 0 && statesExplored < maxStatesExplored {
		state := dequeue()

		srcKey := prettyPrintRegister(state.Configuration.Register) + PrettyPrintAst(state.Configuration.Process)
		if _, ok := visited[srcKey]; !ok {
			visited[srcKey] = vertexId
			vertices[vertexId] = state.Configuration
			vertexId = vertexId + 1
		}

		confs := trans(state.Configuration)
		for _, conf := range confs {
			fmt.Println(PrettyPrintConfiguration(conf))

			dstKey := prettyPrintRegister(conf.Register) + PrettyPrintAst(conf.Process)
			if _, ok := visited[dstKey]; !ok {
				visited[dstKey] = vertexId
				vertices[vertexId] = conf
				vertexId = vertexId + 1
			}
			edges = append(edges, GraphEdge{
				Source:      visited[srcKey],
				Destination: visited[dstKey],
				Label:       conf.Label,
			})

			nextState := &State{
				Configuration: conf,
				NextStates:    []*State{},
			}
			state.NextStates = append(state.NextStates, nextState)
			queue.PushBack(nextState)
		}
		if len(confs) > 0 {
			fmt.Println(len(confs))
			statesExplored = statesExplored + 1
		}
	}
	return vertices, edges
}

func trans(conf Configuration) []Configuration {
	process := conf.Process
	switch process.Type() {
	// DBLINP = INP1 + INP2A/INP2B
	case ElemTypInput:
		inp1Conf := deepcopy.Copy(conf).(Configuration)
		inpElem := inp1Conf.Process.(*ElemInput)

		// Find the input channel label in the register.
		inpLabel := inp1Conf.Register.GetLabel(inpElem.Channel.Name)
		inp1Conf.Label = Label{
			Symbol: Symbol{
				Type:  SymbolTypInput,
				Value: inpLabel,
			},
		}

		// Replace the input element with the inp element.
		inp1Conf.Process = &ElemInpInput{
			Input:   inpElem.Input,
			Next:    inpElem.Next,
			SetType: ElemSetInp,
		}

		// INP2A
		var confs []Configuration
		for _, label := range inp1Conf.Register.Labels() {
			inp2aConf := deepcopy.Copy(inp1Conf).(Configuration)
			inpInputElem := inp2aConf.Process.(*ElemInpInput)
			substituteName(inpInputElem, inpInputElem.Input, Name{
				Name: inp2aConf.Register.GetName(label),
				Type: Fresh,
			})
			inp2aConf.Label.Double = true
			inp2aConf.Label.Symbol2 = Symbol{
				Type:  SymbolTypKnown,
				Value: label,
			}
			inp2aConf.Process = inpInputElem.Next
			confs = append(confs, inp2aConf)
		}

		// INP2B
		inp2bConf := deepcopy.Copy(inp1Conf).(Configuration)
		inpInpElem := inp2bConf.Process.(*ElemInpInput)

		name := inpInpElem.Input.Name
		freshNamesP := GetAllFreshNames(inpInpElem.Next)
		inp2bConf.Label.Double = true
		inp2bConf.Label.Symbol2 = Symbol{
			Type:  SymbolTypFreshInput,
			Value: inp2bConf.Register.UpdateMin(name, freshNamesP),
		}
		inp2bConf.Process = inpInpElem.Next

		return append(confs, inp2bConf)

	// DBLOUT = OUT1 + OUT2
	case ElemTypOutput:
		out1Conf := deepcopy.Copy(conf).(Configuration)
		outElem := out1Conf.Process.(*ElemOutput)

		outLabel := out1Conf.Register.GetLabel(outElem.Channel.Name)
		out1Conf.Label = Label{
			Symbol: Symbol{
				Type:  SymbolTypOutput,
				Value: outLabel,
			},
		}
		out1Conf.Process = &ElemOutOutput{
			Output:  outElem.Output,
			Next:    outElem.Next,
			SetType: ElemSetOut,
		}

		// OUT2
		var confs []Configuration
		for _, label := range out1Conf.Register.Labels() {
			out2Conf := deepcopy.Copy(out1Conf).(Configuration)
			outOutputElem := out2Conf.Process.(*ElemOutOutput)
			out2Conf.Label.Double = true
			out2Conf.Label.Symbol2 = Symbol{
				Type:  SymbolTypKnown,
				Value: label,
			}
			out2Conf.Process = outOutputElem.Next
			confs = append(confs, out2Conf)
		}
		return confs

	// MATCH
	case ElemTypMatch:
		var confs []Configuration

		matchElem := conf.Process.(*ElemMatch)
		// o ¦- [a=a]P
		if matchElem.NameL.Name == matchElem.NameR.Name {
			// o ¦- P
			matchConf := deepcopy.Copy(conf).(Configuration)
			matchElem = matchConf.Process.(*ElemMatch)
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
		baseResConf := deepcopy.Copy(conf).(Configuration)

		// RES
		// P^
		resConf := deepcopy.Copy(conf).(Configuration)
		resElem := resConf.Process.(*ElemRestriction)
		resName := resElem.Restrict.Name
		resConf.Process = resElem.Next
		// (o+a) ¦- P^
		resLabel := resConf.Register.UpdateAfter(resName)
		// (o+a) ¦- P^ -t-> (o'+a) ¦- P^' -t-> (o'+a) ¦- P^'
		tconfs := trans(resConf)
		// (o'+a) ¦- P^
		for _, conf := range tconfs {
			// t != (|o|+1)
			if conf.Label.Symbol.Value == resLabel {
				continue
			}

			// OPEN
			if conf.Label.Double && conf.Label.Symbol.Type == SymbolTypOutput &&
				conf.Label.Symbol2.Type == SymbolTypKnown &&
				conf.Label.Symbol2.Value == resLabel {
				// o
				conf.Register = deepcopy.Copy(baseResConf).(Configuration).Register
				// fn(P')
				freeNamesP := GetAllFreshNames(conf.Process)
				// o[j -> a], j = min{j | reg(j) !E fn(P')}
				label := conf.Register.UpdateMin(resName, freeNamesP)
				// ij
				conf.Label.Symbol2.Value = label
				// ij^
				conf.Label.Symbol2.Type = SymbolTypFreshOutput
				// o |- P'
				confs = append(confs, conf)
				continue
			}

			// $a.P^'
			conf.Process = &ElemRestriction{
				Restrict: resElem.Restrict,
				Next:     conf.Process,
			}
			// o' ¦- $a.P^'
			conf.Register.RemoveLastLabel()
			confs = append(confs, conf)
		}

		return confs

	// REC
	case ElemTypProcess:
		procConf := deepcopy.Copy(conf).(Configuration)
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

		// Detects infinitely recursive processes such as P(a) = P(a).
		if infProc {
			return []Configuration{}
		}
		infProc = true
		tconfs := trans(procConf)
		infProc = false

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
		proc := conf.Process.(*ElemParallel)
		basePar := deepcopy.Copy(conf).(Configuration)

		// PAR1_L
		if getElemSetType(proc.ProcessL) == ElemSetReg {
			parConf := deepcopy.Copy(conf).(Configuration)
			parElem := parConf.Process.(*ElemParallel)
			parConf.Process = parElem.ProcessL
			tconfs := trans(parConf)

			// PAR2_L
			for _, conf := range tconfs {
				parConf = deepcopy.Copy(basePar).(Configuration)

				// When DBPINP/DBLOUT and the 2nd label is fresh input/fresh output.
				if conf.Label.Double &&
					(conf.Label.Symbol2.Type == SymbolTypFreshInput ||
						conf.Label.Symbol2.Type == SymbolTypFreshOutput) {
					// Find fn(P', Q).
					freeNamesP := GetAllFreshNames(conf.Process)
					freeNamesQ := GetAllFreshNames(parElem.ProcessR)
					// Get the name reg(i).
					name := conf.Register.GetName(conf.Label.Symbol2.Value)
					// Update register to be j = min{j | reg(j) \notin fn(P′,Q)}.
					newLabel := parConf.Register.UpdateMin(name,
						append(freeNamesP, freeNamesQ...))
					// Update the label j.
					parConf.Label = conf.Label
					parConf.Label.Symbol2.Value = newLabel
				} else {
					parConf.Label = conf.Label
					parConf.Register = conf.Register
				}
				// Insert P' to P' | Q.
				parConf.Process.(*ElemParallel).ProcessL = conf.Process

				lconfs = append(lconfs, parConf)
			}
		}

		// PAR1_R
		if getElemSetType(proc.ProcessR) == ElemSetReg {
			parConf := deepcopy.Copy(conf).(Configuration)
			parElem := parConf.Process.(*ElemParallel)
			parConf.Process = parElem.ProcessR
			tconfs := trans(parConf)

			// PAR2_R
			for _, conf := range tconfs {
				parConf = deepcopy.Copy(basePar).(Configuration)
				// When DBPINP/DBLOUT and the 2nd label is fresh input/fresh output.
				if conf.Label.Double &&
					(conf.Label.Symbol2.Type == SymbolTypFreshInput ||
						conf.Label.Symbol2.Type == SymbolTypFreshOutput) {
					// Find fn(P, Q').
					freeNamesQ := GetAllFreshNames(conf.Process)
					freeNamesP := GetAllFreshNames(parElem.ProcessL)
					// Get the name reg(i).
					name := conf.Register.GetName(conf.Label.Symbol2.Value)
					// Update register to be j = min{j | reg(j) \notin fn(P,Q')}.
					newLabel := parConf.Register.UpdateMin(name,
						append(freeNamesP, freeNamesQ...))
					// Update the label j.
					parConf.Label = conf.Label
					parConf.Label.Symbol2.Value = newLabel
				} else {
					parConf.Label = conf.Label
					parConf.Register = conf.Register
				}
				// Insert Q' to P | Q'.
				parConf.Process.(*ElemParallel).ProcessR = conf.Process

				rconfs = append(rconfs, parConf)
			}
		}

		confs = append(confs, append(lconfs, rconfs...)...)

		// COMM_L
		for _, lconf := range lconfs {
			for _, rconf := range rconfs {
				if lconf.Label.Double &&
					lconf.Label.Symbol.Type == SymbolTypOutput &&
					lconf.Label.Symbol2.Type == SymbolTypKnown &&
					rconf.Label.Double &&
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
				if lconf.Label.Double &&
					lconf.Label.Symbol.Type == SymbolTypInput &&
					lconf.Label.Symbol2.Type == SymbolTypKnown &&
					rconf.Label.Double &&
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
		clconf.Register.AddEmptyName()
		parElem := clconf.Process.(*ElemParallel)
		// (#+o) ¦- P
		clconf.Process = parElem.ProcessL
		// -t-> (b+o) ¦- P'
		clconfs := trans(clconf)

		crconf := deepcopy.Copy(conf).(Configuration)
		// (#+o)
		crconf.Register.AddEmptyName()
		parElem = crconf.Process.(*ElemParallel)
		// (#+o) ¦- Q
		crconf.Process = parElem.ProcessR
		// -t-> (b+o) ¦- Q'
		crconfs := trans(crconf)

		for _, lconf := range clconfs {
			for _, rconf := range crconfs {
				// CLOSE_L
				if lconf.Label.Double &&
					lconf.Label.Symbol.Type == SymbolTypOutput &&
					lconf.Label.Symbol2.Type == SymbolTypFreshOutput &&
					lconf.Label.Symbol2.Value == 1 &&
					rconf.Label.Double &&
					rconf.Label.Symbol.Type == SymbolTypInput &&
					rconf.Label.Symbol2.Type == SymbolTypFreshInput &&
					rconf.Label.Symbol2.Value == 1 &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value {
					{
						close := deepcopy.Copy(basePar).(Configuration)
						lproc := deepcopy.Copy(lconf.Process).(Element)
						rproc := deepcopy.Copy(rconf.Process).(Element)

						// Q'{a/b}
						resName := lconf.Register.GetName(1)
						oldName := Name{
							Name: rconf.Register.GetName(1),
							Type: Bound,
						}
						newName := Name{
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
				// CLOSE_R
				if lconf.Label.Double &&
					lconf.Label.Symbol.Type == SymbolTypInput &&
					lconf.Label.Symbol2.Type == SymbolTypFreshInput &&
					lconf.Label.Symbol2.Value == 1 &&
					rconf.Label.Double &&
					rconf.Label.Symbol.Type == SymbolTypOutput &&
					rconf.Label.Symbol2.Type == SymbolTypFreshOutput &&
					rconf.Label.Symbol2.Value == 1 &&
					lconf.Label.Symbol.Value == rconf.Label.Symbol.Value {
					{
						close := deepcopy.Copy(basePar).(Configuration)
						lproc := deepcopy.Copy(lconf.Process).(Element)
						rproc := deepcopy.Copy(rconf.Process).(Element)

						// P'{a/b}
						resName := rconf.Register.GetName(1)
						oldName := Name{
							Name: lconf.Register.GetName(1),
							Type: Bound,
						}
						newName := Name{
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

// PrettyPrintConfiguration returns a pretty printed string of the configuration.
func PrettyPrintConfiguration(conf Configuration) string {
	return prettyPrintLabel(conf.Label) + " -> " + prettyPrintRegister(conf.Register) + " ¦- " +
		PrettyPrintAst(conf.Process)

}

func prettyPrintRegister(register Register) string {
	str := "{"
	labels := register.Labels()
	reg := register.Register

	for i, label := range labels {
		if i == len(labels)-1 {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + ")"
		} else {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + "),"
		}
	}
	return str + "}"
}

func prettyPrintLabel(label Label) string {
	if !label.Double {
		return prettyPrintSymbol(label.Symbol)
	}
	return prettyPrintSymbol(label.Symbol) + prettyPrintSymbol(label.Symbol2)
}

func prettyPrintSymbol(symbol Symbol) string {
	s := symbol.Value
	switch symbol.Type {
	case SymbolTypInput:
		return strconv.Itoa(s) + " "
	case SymbolTypOutput:
		return strconv.Itoa(s) + "'"
	case SymbolTypFreshInput:
		return strconv.Itoa(s) + "*"
	case SymbolTypFreshOutput:
		return strconv.Itoa(s) + "^"
	case SymbolTypTau:
		return "t   "
	case SymbolTypKnown:
		return strconv.Itoa(s) + " "
	}
	return ""
}
