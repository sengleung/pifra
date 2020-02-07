package main

import (
	"sort"
	"strconv"

	"github.com/mohae/deepcopy"
)

var registerSize = 10000

type SymbolType int

const (
	SymbolTypTau SymbolType = iota
	SymbolTypInput
	SymbolTypOutput
	SymbolTypFreshInput
	SymbolTypFreshOutput
	SymbolTypTransition
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
	Size      int
	Index     int
	Register  map[int]string
	NameRange map[string]int
}

// UpdateAfter adds a free name to the register at the next label.
// reg+v = reg U {(|reg|+1, v)}.
func (reg *Register) UpdateAfter(freeName string) int {
	index := reg.Index
	reg.Register[index] = freeName
	reg.NameRange[freeName] = index
	reg.Index = reg.Index + 1
	return index
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
		if !freshNamesSet[reg.GetName(label)] {
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
	return reg.Register[label]
}

// GetLabel returns register label corresponding to the name.
func (reg *Register) GetLabel(name string) int {
	return reg.NameRange[name]
}

func (reg *Register) find(i int) string { return "" }         // TODO
func (reg *Register) findAll() []string { return []string{} } // TODO

type Configuration struct {
	Process  Element
	Register Register
	Label    Label
}

type TransitionState struct {
	Configuration Configuration
	Transitions   []*TransitionState
}

func newTransitionStateRoot(process Element) *TransitionState {
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

	register := make(map[int]string, registerSize)
	index := 1
	for _, name := range freshNames {
		register[index] = name
		index = index + 1
	}
	nameRange := make(map[string]int, registerSize)
	for label, name := range register {
		nameRange[name] = label
	}
	return &TransitionState{
		Configuration: Configuration{
			Process: process,
			Register: Register{
				Size:      registerSize,
				Index:     len(register) + 1,
				Register:  register,
				NameRange: nameRange,
			},
		},
		Transitions: []*TransitionState{},
	}
}

func trans(conf Configuration) []Configuration {
	process := conf.Process
	switch process.Type() {
	// INP1
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
		return []Configuration{inp1Conf}

	// INP2A / INP2B
	case ElemTypInpInput:
		// INP2A
		var confs []Configuration
		for _, label := range conf.Register.Labels() {
			inp2aConf := deepcopy.Copy(conf).(Configuration)
			inpInputElem := inp2aConf.Process.(*ElemInpInput)
			substituteName(inpInputElem, inpInputElem.Input, Name{
				Name: inp2aConf.Register.GetName(label),
				Type: Fresh,
			})
			inp2aConf.Label = Label{
				Symbol: Symbol{
					Type:  SymbolTypInput,
					Value: label,
				},
			}
			inp2aConf.Process = inpInputElem.Next
			confs = append(confs, inp2aConf)
		}

		// INP2B
		inp2bConf := deepcopy.Copy(conf).(Configuration)
		inpInpElem := inp2bConf.Process.(*ElemInpInput)

		name := inpInpElem.Input.Name
		freshNamesP := GetAllFreshNames(inpInpElem.Next)
		inp2bConf.Label = Label{
			Symbol: Symbol{
				Type:  SymbolTypFreshInput,
				Value: inp2bConf.Register.UpdateMin(name, freshNamesP),
			},
		}
		inp2bConf.Process = inpInpElem.Next

		return append(confs, inp2bConf)

	// OUT1
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
		return []Configuration{out1Conf}

	// OUT2
	case ElemTypOutOutput:
		var confs []Configuration
		for _, label := range conf.Register.Labels() {
			out2Conf := deepcopy.Copy(conf).(Configuration)
			outOutputElem := out2Conf.Process.(*ElemOutOutput)
			out2Conf.Label = Label{
				Symbol: Symbol{
					Type:  SymbolTypTransition,
					Value: label,
				},
			}
			out2Conf.Process = outOutputElem.Next
			confs = append(confs, out2Conf)
		}
		return confs

	// MATCH
	case ElemTypMatch:
	// RES, OPEN
	case ElemTypRestriction:
		var confs []Configuration
		resConf := deepcopy.Copy(conf).(Configuration)
		resElem := resConf.Process.(*ElemRestriction)
		resConf.Process = resElem.Next
		tconfs := trans(resConf)
		dconfs := dblTrans(tconfs)
		for _, conf := range dconfs {
			if conf.Label.Double && conf.Label.Symbol.Type == SymbolTypOutput {

			} else {
				confs = append(confs, conf)
			}
		}

	// REC
	case ElemTypProcess, ElemTypProcessConstants:
	// SUM
	case ElemTypSum:
		var confs []Configuration

		// SUM_L
		sumConf := deepcopy.Copy(conf).(Configuration)
		sumElem := sumConf.Process.(*ElemSum)
		sumConf.Process = sumElem.ProcessL
		lconfs := trans(sumConf)
		dconfs := dblTrans(lconfs)
		confs = append(confs, dconfs...)

		// SUM_R
		sumConf = deepcopy.Copy(conf).(Configuration)
		sumElem = sumConf.Process.(*ElemSum)
		sumConf.Process = sumElem.ProcessR
		rconfs := trans(sumConf)
		dconfs = dblTrans(rconfs)
		confs = append(confs, dconfs...)

		return confs

	// PAR1, PAR2, COMM, CLOSE
	case ElemTypParallel:
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
			dconfs := dblTrans(tconfs)

			// PAR2_L
			for _, conf := range dconfs {
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
			dconfs := dblTrans(tconfs)

			// PAR2_R
			for _, conf := range dconfs {
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
				// Insert P' to P | Q'.
				parConf.Process.(*ElemParallel).ProcessR = conf.Process

				rconfs = append(rconfs, parConf)
			}
		}

		return append(lconfs, rconfs...)

	case ElemTypRoot:
		rootConf := deepcopy.Copy(conf).(Configuration)
		rootConf.Process = rootConf.Process.(*ElemRoot).Next
		tconfs := trans(rootConf)
		dconfs := dblTrans(tconfs)
		return dconfs
	}
	return nil
}

func dblTrans(confs []Configuration) []Configuration {
	var dblInpOuts []Configuration

	// Keep existing double inputs/double outputs.
	for _, conf := range confs {
		if conf.Label.Double {
			dblInpOuts = append(dblInpOuts, conf)
		}
	}

	// trans() intermediate input processes.
	var interConfs []Configuration

	for _, conf := range confs {
		elemSetType := getElemSetType(conf.Process)
		if !conf.Label.Double && (elemSetType == ElemSetInp || elemSetType == ElemSetOut) {
			interConfs = append(interConfs, conf)
		}
	}

	for _, conf := range interConfs {
		tconfs := trans(conf)

		var dconfs []Configuration
		for _, dblConf := range tconfs {
			if getElemSetType(dblConf.Process) == ElemSetReg && !conf.Label.Double {
				dconfs = append(dconfs, dblConf)
			}
		}

		for _, dconf := range dconfs {
			dconf.Label = Label{
				Double:  true,
				Symbol:  conf.Label.Symbol,
				Symbol2: dconf.Label.Symbol,
			}
			dblInpOuts = append(dblInpOuts, dconf)
		}
	}

	return dblInpOuts
}

func getElemSetType(elem Element) ElemSetType {
	switch elem.Type() {
	case ElemTypNil:
		nilElem := elem.(*ElemNil)
		return nilElem.SetType
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		return outElem.SetType
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		return inpElem.SetType
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		return matchElem.SetType
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		return resElem.SetType
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		return sumElem.SetType
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		return parElem.SetType
	case ElemTypProcess:
		procElem := elem.(*ElemProcess)
		return procElem.SetType
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		return pcsElem.SetType
	case ElemTypOutOutput:
		elemOutOut := elem.(*ElemOutOutput)
		return elemOutOut.SetType
	case ElemTypInpInput:
		elemInpInp := elem.(*ElemInpInput)
		return elemInpInp.SetType
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return rootElem.SetType
	}
	return ElemSetReg
}

// PrettyPrintConfiguration returns a pretty printed string of the configuration.
func PrettyPrintConfiguration(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + " ¦- " +
		PrettyPrintAst(conf.Process) + " : " +
		prettyPrintLabel(conf.Label)
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
		return prettyPrintSymbol(label.Symbol) + " "
	}
	return prettyPrintSymbol(label.Symbol) + prettyPrintSymbol(label.Symbol2) + " "
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
		return "t "
	case SymbolTypTransition:
		return strconv.Itoa(s) + " "
	}
	return ""
}
