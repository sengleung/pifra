package main

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
	Transitions []TransitionState
}

func newTransitionStateRoot(process Element) TransitionState {
	freshNames := getAllFreshNames(process)
	var register map[int]string
	for i, name := range freshNames {
		register[i+1] = name
	}
	return TransitionState{
		State: State{
			Process: process,
			Register: Register{
				Size:     registerSize,
				Register: register,
			},
		},
	}
}
