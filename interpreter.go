package main

import (
	"sort"
	"strconv"

	"github.com/mohae/deepcopy"
)

var registerSize = 10000

type TransitionType int

const (
	Inp1 TransitionType = iota
	Inp2A
	Inp2B
	DblInp
	Out1
	Out2
	DblOut
	Match
	Res
	Rec
	Sum
	Par1
	Par2
	Comm
	Close
)

type TransitionLabelType int

const (
	Known TransitionLabelType = iota
	FreshInput
	FreshOutput
	Tau
	OutKnown
	OutFreshOutput
	InpKnown
	InpFreshInput
)

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

type Labell struct {
	Type   TransitionLabelType
	Label1 int
	Label2 int
}

type Register struct {
	Size      int
	Index     int
	Register  map[int]string
	NameRange map[string]int
}

// Update adds a fresh name to the register at the minimum index.
func (reg *Register) Update() int {
	freeName := generateFreshName("fn")
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

type TransitionLabel struct {
	Rule  TransitionType
	Label Labell
}

type TransitionState struct {
	Configuration Configuration
	Label         TransitionLabel
	Transitions   []*TransitionState
}

type Direction int

const (
	Next Direction = iota
	Left
	Right
)

type Path struct {
	Directions        []Direction
	BeforeElementType ElementType
	ElementType       ElementType
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

func popDirs(dirs []Direction) (Direction, []Direction) {
	var direction Direction
	direction, dirs = dirs[len(dirs)-1], dirs[:len(dirs)-1]
	return direction, dirs
}

func produceTransitionStates(ts *TransitionState) {
	var acc func(Element, Configuration, []Direction) []*TransitionState
	acc = func(elem Element, state Configuration, dirs []Direction) []*TransitionState {
		transitionStates := []*TransitionState{}
		switch elem.Type() {
		case ElemTypNil:
		case ElemTypOutput:

		// DBLINP
		case ElemTypInput:
			dblInputs := doDblInp(state, dirs)
			transitionStates = append(transitionStates, dblInputs...)
		case ElemTypMatch:
			matchElem := elem.(*ElemMatch)
			if matchElem.NameL.Name == matchElem.NameR.Name {
				dirs = append(dirs, Next)
				acc(matchElem.Next, state, dirs)
			}
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			dirs = append(dirs, Next)
			acc(resElem.Next, state, dirs)

		// SUM
		case ElemTypSum:
			sumElem := elem.(*ElemSum)

			// Copy the process.
			// Remove the sum element and take the left-hand side process.
			// No need to append to the directions because element is removed.
			sc := deepcopy.Copy(state)
			stateCopy := sc.(Configuration)
			penultimateDirs := dirs[:len(dirs)-1]
			lastDir := dirs[len(dirs)-1]
			process, _ := findElement(stateCopy.Process, penultimateDirs)
			removeElementAfter(process, lastDir, Left)
			trns := acc(sumElem.ProcessL, stateCopy, dirs)
			transitionStates = append(transitionStates, trns...)

			// Remove the sum element and take the right-hand side process.
			sc = deepcopy.Copy(state)
			stateCopy = sc.(Configuration)
			process, _ = findElement(stateCopy.Process, penultimateDirs)
			removeElementAfter(process, lastDir, Right)
			trns = acc(sumElem.ProcessR, stateCopy, dirs)
			transitionStates = append(transitionStates, trns...)

		// PAR1, PAR2
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)

			dirs = append(dirs, Left)
			trns := acc(parElem.ProcessL, state, dirs)
			transitionStates = append(transitionStates, trns...)
			_, dirs = popDirs(dirs)

			dirs = append(dirs, Right)
			trns = acc(parElem.ProcessR, state, dirs)
			transitionStates = append(transitionStates, trns...)
			_, dirs = popDirs(dirs)

		case ElemTypProcess:
		case ElemTypProcessConstants:
		case ElemTypOutOutput:
			outOutput := elem.(*ElemOutOutput)
			dirs = append(dirs, Next)
			trns := acc(outOutput.Next, state, dirs)
			transitionStates = append(transitionStates, trns...)
		case ElemTypInpInput:
			inpInput := elem.(*ElemInpInput)
			dirs = append(dirs, Next)
			trns := acc(inpInput.Next, state, dirs)
			transitionStates = append(transitionStates, trns...)
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			dirs = append(dirs, Next)
			trns := acc(rootElem.Next, state, dirs)
			transitionStates = append(transitionStates, trns...)
		}
		return transitionStates
	}
	trns := acc(ts.Configuration.Process, ts.Configuration, []Direction{})
	ts.Transitions = append(ts.Transitions, trns...)
}

func doDblInp(tsState Configuration, prefixPath []Direction) []*TransitionState {
	inputs := getFirstInputs(tsState.Process, prefixPath)
	dblInputs := []*TransitionState{}

	for _, path := range inputs {
		// INP1 transition relation
		sc := deepcopy.Copy(tsState)
		inp1 := sc.(Configuration)

		// Get the input element.
		inpElem, _ := findElement(inp1.Process, path.Directions)
		// Find the input channel label in the register.
		inpLabel := inp1.Register.NameRange[inpElem.(*ElemInput).Channel.Name]

		penultimateDirs := path.Directions[:len(path.Directions)-1]
		lastDir := path.Directions[len(path.Directions)-1]

		// Find the penultimate element before the input element.
		elemBefore, _ := findElement(inp1.Process, penultimateDirs)
		// Replace the input element with the inp element.
		replaceElement(elemBefore, &ElemInpInput{
			Input: inpElem.(*ElemInput).Input,
			Next:  inpElem.(*ElemInput).Next,
		}, lastDir)

		// INP2A transition relation
		for _, label := range inp1.Register.Labels() {
			sc2 := deepcopy.Copy(inp1)
			inp2a := sc2.(Configuration)

			// Substitute the input bound name with the labelled fresh name.
			inpInputElem, _ := findElement(inp2a.Process, path.Directions)
			substituteName(inpInputElem, inpInputElem.(*ElemInpInput).Input, Name{
				Name: inp2a.Register.GetName(label),
				Type: Fresh,
			})

			// Find the penultimate element before the inp element in the copied AST.
			elemBefore, _ := findElement(inp2a.Process, penultimateDirs)
			// Remove the input element.
			removeElementAfter(elemBefore, lastDir, Next)

			dblInp2a := &TransitionState{
				Configuration: inp2a,
				Label: TransitionLabel{
					Rule: DblInp,
					Label: Labell{
						Type:   InpKnown,
						Label1: inpLabel,
						Label2: label,
					},
				},
			}
			dblInputs = append(dblInputs, dblInp2a)
		}

		// INP2B transition relation
		sc2 := deepcopy.Copy(inp1)
		inp2b := sc2.(Configuration)

		// Find the penultimate element before the inp element in the copied AST.
		elemBefore, _ = findElement(inp2b.Process, penultimateDirs)
		// Remove the input element.
		removeElementAfter(elemBefore, lastDir, Next)

		dblInp2b := &TransitionState{
			Configuration: inp2b,
			Label: TransitionLabel{
				Rule: DblInp,
				Label: Labell{
					Type:   InpFreshInput,
					Label1: inpLabel,
					Label2: inp2b.Register.Update(),
				},
			},
		}
		dblInputs = append(dblInputs, dblInp2b)
	}

	return dblInputs
}

func getFirstInputs(elem Element, prefixPath []Direction) []Path {
	return getFirstInpOuts(elem, prefixPath, true)
}

func getFirstOutputs(elem Element, prefixPath []Direction) []Path {
	return getFirstInpOuts(elem, prefixPath, false)
}

func getFirstInpOuts(elem Element, prefixPath []Direction, wantInputs bool) []Path {
	paths := []Path{}
	curPath := []Direction{}

	elem, _ = findElement(elem, prefixPath)

	popPath := func() Direction {
		var direction Direction
		direction, curPath = curPath[len(curPath)-1], curPath[:len(curPath)-1]
		return direction
	}

	var acc func(Element)
	acc = func(elem Element) {
		switch elem.Type() {
		case ElemTypNil:
		case ElemTypOutput:
			if !wantInputs {
				foundDirection := append(deepcopy.Copy(prefixPath).([]Direction),
					deepcopy.Copy(curPath).([]Direction)...)
				paths = append(paths, Path{
					Directions:  foundDirection,
					ElementType: ElemTypOutput,
				})
			}
		case ElemTypInput:
			if wantInputs {
				foundDirection := append(deepcopy.Copy(prefixPath).([]Direction),
					deepcopy.Copy(curPath).([]Direction)...)
				paths = append(paths, Path{
					Directions:  foundDirection,
					ElementType: ElemTypInput,
				})
			}
		case ElemTypMatch:
			matchElem := elem.(*ElemMatch)
			if matchElem.NameL.Name == matchElem.NameR.Name {
				curPath = append(curPath, Next)
				acc(matchElem.Next)
				popPath()
			}
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			curPath = append(curPath, Next)
			acc(resElem.Next)
			popPath()
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			curPath = append(curPath, Left)
			acc(sumElem.ProcessL)
			popPath()
			curPath = append(curPath, Right)
			acc(sumElem.ProcessR)
			popPath()
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			curPath = append(curPath, Left)
			acc(parElem.ProcessL)
			popPath()
			curPath = append(curPath, Right)
			acc(parElem.ProcessR)
			popPath()
		case ElemTypProcess:
		case ElemTypProcessConstants:
		case ElemTypOutOutput:
			outOutput := elem.(*ElemOutOutput)
			curPath = append(curPath, Next)
			acc(outOutput.Next)
			popPath()
		case ElemTypInpInput:
			inpInput := elem.(*ElemInpInput)
			curPath = append(curPath, Next)
			acc(inpInput.Next)
			popPath()
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			curPath = append(curPath, Next)
			acc(rootElem.Next)
			popPath()
		}
	}
	acc(elem)

	return paths
}

func findElement(elem Element, directions []Direction) (Element, bool) {
	for _, direction := range directions {
		switch elem.Type() {
		case ElemTypNil:
			return nil, false
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			if direction == Next {
				elem = outElem.Next
				continue
			}
			return nil, false
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			if direction == Next {
				elem = inpElem.Next
				continue
			}
			return nil, false
		case ElemTypMatch:
			matchElem := elem.(*ElemMatch)
			if direction == Next {
				elem = matchElem.Next
				continue
			}
			return nil, false
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			if direction == Next {
				elem = resElem.Next
				continue
			}
			return nil, false
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			switch direction {
			case Next:
				return nil, false
			case Left:
				elem = sumElem.ProcessL
				continue
			case Right:
				elem = sumElem.ProcessR
				continue
			}
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			switch direction {
			case Next:
				return nil, false
			case Left:
				elem = parElem.ProcessL
				continue
			case Right:
				elem = parElem.ProcessR
				continue
			}
		case ElemTypProcess:
			return nil, false
		case ElemTypProcessConstants:
			return nil, false
		case ElemTypOutOutput:
			outOutput := elem.(*ElemOutOutput)
			if direction == Next {
				elem = outOutput.Next
				continue
			}
			return nil, false
		case ElemTypInpInput:
			inpInput := elem.(*ElemInpInput)
			if direction == Next {
				elem = inpInput.Next
				continue
			}
			return nil, false
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			if direction == Next {
				elem = rootElem.Next
				continue
			}
			return nil, false
		}
	}
	if elem != nil {
		return elem, true
	}
	return nil, false
}

func replaceElement(elemBefore Element, elemNext Element, direction Direction) {
	switch elemBefore.Type() {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elemBefore.(*ElemOutput)
		outElem.Next = elemNext
	case ElemTypInput:
		inpElem := elemBefore.(*ElemInput)
		inpElem.Next = elemNext
	case ElemTypMatch:
		matchElem := elemBefore.(*ElemMatch)
		matchElem.Next = elemNext
	case ElemTypRestriction:
		resElem := elemBefore.(*ElemRestriction)
		resElem.Next = elemNext
	case ElemTypSum:
		sumElem := elemBefore.(*ElemSum)
		switch direction {
		case Next:
		case Left:
			sumElem.ProcessL = elemNext
		case Right:
			sumElem.ProcessR = elemNext
		}
	case ElemTypParallel:
		parElem := elemBefore.(*ElemParallel)
		switch direction {
		case Next:
		case Left:
			parElem.ProcessL = elemNext
		case Right:
			parElem.ProcessR = elemNext
		}
	case ElemTypProcess:
	case ElemTypProcessConstants:
	case ElemTypOutOutput:
		outOutput := elemBefore.(*ElemOutOutput)
		outOutput.Next = elemNext
	case ElemTypInpInput:
		inpInput := elemBefore.(*ElemInpInput)
		inpInput.Next = elemNext
	case ElemTypRoot:
		rootElem := elemBefore.(*ElemRoot)
		rootElem.Next = elemNext
	}
}

func removeElementAfter(elem Element, direction Direction, nextDirection Direction) {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = nextElement(outElem.Next, nextDirection)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = nextElement(inpElem.Next, nextDirection)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = nextElement(matchElem.Next, nextDirection)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = nextElement(resElem.Next, nextDirection)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		switch direction {
		case Next:
		case Left:
			sumElem.ProcessL = nextElement(sumElem.ProcessL, nextDirection)
		case Right:
			sumElem.ProcessR = nextElement(sumElem.ProcessR, nextDirection)
		}
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		switch direction {
		case Next:
		case Left:
			parElem.ProcessL = nextElement(parElem.ProcessL, nextDirection)
		case Right:
			parElem.ProcessR = nextElement(parElem.ProcessR, nextDirection)
		}
	case ElemTypProcess:
	case ElemTypProcessConstants:
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		outOutput.Next = nextElement(outOutput.Next, nextDirection)
	case ElemTypInpInput:
		inpInput := elem.(*ElemInpInput)
		inpInput.Next = nextElement(inpInput.Next, nextDirection)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = nextElement(rootElem.Next, nextDirection)
	}
}

func nextElement(elem Element, direction Direction) Element {
	switch elem.Type() {
	case ElemTypNil:
		return nil
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		return outElem.Next
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		return inpElem.Next
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		return matchElem.Next
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		return resElem.Next
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		switch direction {
		case Next:
			return sumElem.ProcessL
		case Left:
			return sumElem.ProcessL
		case Right:
			return sumElem.ProcessR
		}
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		switch direction {
		case Next:
			return parElem.ProcessL
		case Left:
			return parElem.ProcessL
		case Right:
			return parElem.ProcessR
		}
	case ElemTypProcess:
	case ElemTypProcessConstants:
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		return outOutput.Next
	case ElemTypInpInput:
		elemInpInput := elem.(*ElemInpInput)
		return elemInpInput.Next
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return rootElem.Next
	}
	return nil
}

func prettyPrintTransitionState(ts *TransitionState) string {
	return prettyPrintConfiguration(ts.Configuration) + " ¦- " + PrettyPrintAst(ts.Configuration.Process) + " : " +
		prettyPrintTransitionRule(ts.Label) + " : " +
		prettyPrintTransitionLabel(ts.Label)
}

func prettyPrintState(conf Configuration) string {
	return prettyPrintConfiguration(conf) + " ¦- " + PrettyPrintAst(conf.Process) + " : " + prettyPrintLabel(conf.Label)
}

func prettyPrintConfiguration(state Configuration) string {
	str := "{"
	labels := state.Register.Labels()
	reg := state.Register.Register

	for i, label := range labels {
		if i == len(labels)-1 {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + ")"
		} else {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + "),"
		}
	}
	return str + "}"
}

func prettyPrintTransitionRule(label TransitionLabel) string {
	switch label.Rule {
	case Inp1:
		return "INP1"
	case Inp2A:
		return "INP2A"
	case Inp2B:
		return "INP2B"
	case DblInp:
		return "DBLINP"
	case Out1:
		return "OUT1"
	case Out2:
		return "OUT2"
	case DblOut:
		return "DBLOUT"
	case Match:
		return "MATCH"
	case Res:
		return "RES"
	case Rec:
		return "REC"
	case Sum:
		return "SUM"
	case Par1:
		return "PAR1"
	case Par2:
		return "PAR2"
	case Comm:
		return "COMM"
	case Close:
		return "CLOSE"
	}
	return ""
}

func prettyPrintTransitionLabel(label TransitionLabel) string {
	l1 := label.Label.Label1
	l2 := label.Label.Label2
	switch label.Label.Type {
	case Known:
		return strconv.Itoa(l1)
	case FreshInput:
		return strconv.Itoa(l1) + "*"
	case FreshOutput:
		return strconv.Itoa(l1) + "^"
	case Tau:
		return "t "
	case OutKnown:
		return strconv.Itoa(l1) + "^" + strconv.Itoa(l2) + " "
	case OutFreshOutput:
		return strconv.Itoa(l1) + "'" + strconv.Itoa(l2) + "^"
	case InpKnown:
		return strconv.Itoa(l1) + " " + strconv.Itoa(l2) + " "
	case InpFreshInput:
		return strconv.Itoa(l1) + " " + strconv.Itoa(l2) + "*"
	}
	return ""
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

func prettyPrintLabel(label Label) string {
	if !label.Double {
		return prettyPrintSymbol(label.Symbol) + " "
	}
	return prettyPrintSymbol(label.Symbol) + prettyPrintSymbol(label.Symbol2) + " "
}
