package main

import (
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
	Type  TransitionLabelType
	Label []int
}

type Register struct {
	Size      int
	Register  map[int]string
	NameRange map[string]int
}

func (reg *Register) update()           {}                    // TODO
func (reg *Register) find(i int) string { return "" }         // TODO
func (reg *Register) findAll() []string { return []string{} } // TODO

type State struct {
	Process  Element
	Register Register
}

type TransitionLabel struct {
	Type  TransitionType
	Label Label
}

type TransitionState struct {
	State       State
	Label       TransitionLabel
	Transitions *[]*TransitionState
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
	register := make(map[int]string, registerSize)
	for i, name := range freshNames {
		register[i+1] = name
	}
	nameRange := make(map[string]int, registerSize)
	for i, name := range freshNames {
		nameRange[name] = i + 1
	}
	return &TransitionState{
		State: State{
			Process: process,
			Register: Register{
				Size:      registerSize,
				Register:  register,
				NameRange: nameRange,
			},
		},
		Transitions: &[]*TransitionState{},
	}
}

func produceTransitionStates(ts *TransitionState) error {
	inputs := getFirstInputs(ts.State.Process)
	dblInputs := []TransitionState{}

	for _, path := range inputs {
		// INP1 transition relation
		sc := deepcopy.Copy(ts.State)
		inp1 := sc.(State)

		// Get the input element.
		inpElem, _ := findElement(inp1.Process, path.Directions)
		// Find the input channel label in the register.
		label := inp1.Register.NameRange[inpElem.(*ElemInput).Channel.Name]

		penultimateDirs := path.Directions[:len(path.Directions)-1]
		lastDir := path.Directions[len(path.Directions)-1]

		// Find the penultimate element before the input element.
		elemBefore, _ := findElement(inp1.Process, penultimateDirs)
		// Replace the input element with the inp element.
		transplantInpInput(elemBefore, lastDir)

		for i := range inp1.Register.Register {
			// INP2A transition relation
			sc2 := deepcopy.Copy(inp1)
			inp2a := sc2.(State)

			// Find the penultimate element before the input element in the copied AST.
			elemBefore, _ := findElement(inp2a.Process, penultimateDirs)
			// Remove the input element.
			removeElementAfter(elemBefore, lastDir)

			dblInp2a := TransitionState{
				State: State{
					Process:  inp2a.Process,
					Register: deepcopy.Copy(inp1.Register).(Register),
				},
				Label: TransitionLabel{
					Type: Inp1,
					Label: Label{
						Type:  InpKnown,
						Label: []int{label, i},
					},
				},
			}

			dblInputs = append(dblInputs, dblInp2a)
		}
	}

	return nil
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

func findElementBefore(elem Element, pathDirections []Direction) (Element, Direction, bool) {
	if len(pathDirections) == 0 {
		return nil, 0, false
	}
	directions := pathDirections
	if len(pathDirections) > 0 {
		directions = pathDirections[:len(pathDirections)-1]
	}
	lastDirection := pathDirections[len(pathDirections)-1]
	for _, direction := range directions {
		switch elem.Type() {
		case ElemTypNil:
			return nil, 0, false
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			if direction == Next {
				elem = outElem.Next
				continue
			}
			return nil, 0, false
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			if direction == Next {
				elem = inpElem.Next
				continue
			}
			return nil, 0, false
		case ElemTypMatch:
			matchElem := elem.(*ElemMatch)
			if direction == Next {
				elem = matchElem.Next
				continue
			}
			return nil, 0, false
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			if direction == Next {
				elem = resElem.Next
				continue
			}
			return nil, 0, false
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			switch direction {
			case Next:
				return nil, 0, false
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
				return nil, 0, false
			case Left:
				elem = parElem.ProcessL
				continue
			case Right:
				elem = parElem.ProcessR
				continue
			}
		case ElemTypProcess:
			return nil, 0, false
		case ElemTypProcessConstants:
			return nil, 0, false
		}
	}
	if elem != nil {
		return elem, lastDirection, true
	}
	return nil, 0, false
}

func transplantInpInput(elem Element, direction Direction) {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypOutput:
		// outElem := elem.(*ElemOutput)
		// TODO
	case ElemTypInput:
		// inpElem := elem.(*ElemInput)
		// TODO
	case ElemTypMatch:
		// matchElem := elem.(*ElemMatch)
		// TODO
	case ElemTypRestriction:
		// resElem := elem.(*ElemRestriction)
		// TODO
	case ElemTypSum:
		// sumElem := elem.(*ElemSum)
		// TODO
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		switch direction {
		case Next:
		case Left:
			nextElem := parElem.ProcessL.(*ElemInput)
			parElem.ProcessL = &ElemInpInput{
				Input: nextElem.Input,
				Next:  nextElem.Next,
			}
		case Right:
			nextElem := parElem.ProcessR.(*ElemInput)
			parElem.ProcessR = &ElemInpInput{
				Input: nextElem.Input,
				Next:  nextElem.Next,
			}
		}
	case ElemTypProcess:
	case ElemTypProcessConstants:
	case ElemTypOutOutput:
		// outOutput := elem.(*ElemOutOutput)
		// TODO
	case ElemTypInpInput:
		// inpInput := elem.(*ElemInpInput)
		// TODO
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		nextElem := rootElem.Next.(*ElemInput)
		rootElem.Next = &ElemInpInput{
			Input: nextElem.Input,
			Next:  nextElem.Next,
		}
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
