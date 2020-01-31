package main

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
	KnownKnown
	KnownFreshInput
)

type Label struct {
	Type  TransitionLabelType
	Label string
}

type Register struct {
	Size     int
	Register map[int]string
}

func (reg *Register) update() {}

type State struct {
	Process  []Element
	Register Register
}

type TransitionLabel struct {
	Type  TransitionType
	Label []Label
}

type TransitionState struct {
	State       State
	Label       TransitionLabel
	Transitions []TransitionState
}