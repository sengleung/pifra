package main

import "github.com/mohae/deepcopy"

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
	Label []string
}

type Register struct {
	Size     int
	Register map[int]string
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
	Directions  []Direction
	ElementType ElementType
}

func newTransitionStateRoot(process Element) *TransitionState {
	freshNames := getAllFreshNames(process)
	register := make(map[int]string, registerSize)
	for i, name := range freshNames {
		register[i+1] = name
	}
	return &TransitionState{
		State: State{
			Process: process,
			Register: Register{
				Size:     registerSize,
				Register: register,
			},
		},
		Transitions: &[]*TransitionState{},
	}
}

func produceTransitionStates(ts *TransitionState) {
	// TODO
}

func getFirstInpOuts(elem Element) []Path {
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
			paths = append(paths, Path{
				Directions:  deepcopy.Copy(curPath).([]Direction),
				ElementType: ElemTypOutput,
			})
		case ElemTypInput:
			paths = append(paths, Path{
				Directions:  deepcopy.Copy(curPath).([]Direction),
				ElementType: ElemTypInput,
			})
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
		}
	}

	acc(elem)

	return paths
}
