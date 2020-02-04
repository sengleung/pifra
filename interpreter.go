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

type Label struct {
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

// GetName returns register name corresponding to the label.
func (reg *Register) GetName(label int) string {
	return reg.Register[label]
}

func (reg *Register) find(i int) string { return "" }         // TODO
func (reg *Register) findAll() []string { return []string{} } // TODO

type State struct {
	Process  Element
	Register Register
}

type TransitionLabel struct {
	Rule  TransitionType
	Label Label
}

type TransitionState struct {
	State       State
	Label       TransitionLabel
	Transitions []*TransitionState
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
	freshNames := GetAllFreshNames(process)
	freshNamesSet := make(map[string]bool)
	for _, freshName := range freshNames {
		freshNamesSet[freshName] = true
	}

	register := make(map[int]string, registerSize)
	index := 1
	for name := range freshNamesSet {
		register[index] = name
		index = index + 1
	}
	nameRange := make(map[string]int, registerSize)
	for label, name := range register {
		nameRange[name] = label
	}
	return &TransitionState{
		State: State{
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

func produceTransitionStates(ts *TransitionState) {
	dblInputs := doDblInp(ts)
	ts.Transitions = append(ts.Transitions, dblInputs...)
}

func doDblInp(ts *TransitionState) []*TransitionState {
	inputs := getFirstInputs(ts.State.Process)
	dblInputs := []*TransitionState{}

	for _, path := range inputs {
		// INP1 transition relation
		sc := deepcopy.Copy(ts.State)
		inp1 := sc.(State)

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
			inp2a := sc2.(State)

			// Substitute the input bound name with the labelled fresh name.
			inpInputElem, _ := findElement(inp2a.Process, path.Directions)
			substituteName(inpInputElem, inpInputElem.(*ElemInpInput).Input, Name{
				Name: inp2a.Register.GetName(label),
				Type: Fresh,
			})

			// Find the penultimate element before the inp element in the copied AST.
			elemBefore, _ := findElement(inp2a.Process, penultimateDirs)
			// Remove the input element.
			removeElementAfter(elemBefore, lastDir)

			dblInp2a := &TransitionState{
				State: inp2a,
				Label: TransitionLabel{
					Rule: Inp2A,
					Label: Label{
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
		inp2b := sc2.(State)

		// Find the penultimate element before the inp element in the copied AST.
		elemBefore, _ = findElement(inp2b.Process, penultimateDirs)
		// Remove the input element.
		removeElementAfter(elemBefore, lastDir)

		dblInp2b := &TransitionState{
			State: inp2b,
			Label: TransitionLabel{
				Rule: Inp2B,
				Label: Label{
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

func getFirstInputs(elem Element) []Path {
	return getFirstInpOuts(elem, true)
}

func getFirstOutputs(elem Element) []Path {
	return getFirstInpOuts(elem, false)
}

func getFirstInpOuts(elem Element, wantInputs bool) []Path {
	paths := []Path{}
	curPath := []Direction{}

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
				paths = append(paths, Path{
					Directions:  deepcopy.Copy(curPath).([]Direction),
					ElementType: ElemTypOutput,
				})
			}
		case ElemTypInput:
			if wantInputs {
				paths = append(paths, Path{
					Directions:  deepcopy.Copy(curPath).([]Direction),
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

func removeElementAfter(elem Element, direction Direction) {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = nextElement(outElem, direction)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = nextElement(inpElem, direction)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = nextElement(matchElem, direction)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = nextElement(resElem, direction)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		switch direction {
		case Next:
		case Left:
			sumElem.ProcessL = nextElement(sumElem.ProcessL, direction)
		case Right:
			sumElem.ProcessR = nextElement(sumElem.ProcessR, direction)
		}
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		switch direction {
		case Next:
		case Left:
			parElem.ProcessL = nextElement(parElem.ProcessL, direction)
		case Right:
			parElem.ProcessR = nextElement(parElem.ProcessR, direction)
		}
	case ElemTypProcess:
	case ElemTypProcessConstants:
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		outOutput.Next = nextElement(outOutput, direction)
	case ElemTypInpInput:
		inpInput := elem.(*ElemInpInput)
		inpInput.Next = nextElement(inpInput, direction)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = nextElement(rootElem, direction)
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
	return prettyPrintState(ts.State) + " ¦- " + PrettyPrintAst(ts.State.Process) + " : " +
		prettyPrintTransitionRule(ts.Label) + " : " +
		prettyPrintTransitionLabel(ts.Label)
}

func prettyPrintState(state State) string {
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
