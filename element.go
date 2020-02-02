package main

type ElementType int

const (
	ElemTypNil ElementType = iota
	ElemTypOutput
	ElemTypInput
	ElemTypMatch
	ElemTypRestriction
	ElemTypSum
	ElemTypParallel
	ElemTypProcess
	ElemTypProcessConstants

	ElemTypOutOutput
	ElemTypOutRestriction
	ElemTypOutParallelR
	ElemTypOutParallelL

	ElemTypInpInput
	ElemTypInpRestriction
	ElemTypInpParallelR
	ElemTypInpParallelL

	Root
)

type NameType int

const (
	Fresh NameType = iota
	Bound
)

type Name struct {
	Name string
	Type NameType
}

type Element interface {
	Type() ElementType
}

type ElemNil struct{}

func (e *ElemNil) Type() ElementType {
	return ElemTypNil
}

type ElemOutput struct {
	Channel Name
	Output  Name
	Next    Element
}

func (e *ElemOutput) Type() ElementType {
	return ElemTypOutput
}

type ElemInput struct {
	Channel Name
	Input   Name
	Next    Element
}

func (e *ElemInput) Type() ElementType {
	return ElemTypInput
}

type ElemMatch struct {
	NameL Name
	NameR Name
	Next  Element
}

func (e *ElemMatch) Type() ElementType {
	return ElemTypMatch
}

type ElemRestriction struct {
	Restrict Name
	Next     Element
}

func (e *ElemRestriction) Type() ElementType {
	return ElemTypRestriction
}

type ElemSum struct {
	ProcessL Element
	ProcessR Element
}

func (e *ElemSum) Type() ElementType {
	return ElemTypSum
}

type ElemParallel struct {
	ProcessL Element
	ProcessR Element
}

func (e *ElemParallel) Type() ElementType {
	return ElemTypParallel
}

type ElemProcess struct {
	Name string
}

func (e *ElemProcess) Type() ElementType {
	return ElemTypProcess
}

type ElemProcessConstants struct {
	Name       string
	Parameters []Name
}

func (e *ElemProcessConstants) Type() ElementType {
	return ElemTypProcessConstants
}

type ElemOutOutput struct {
	Output Name
	Next   Element
}

func (e *ElemOutOutput) Type() ElementType {
	return ElemTypOutOutput
}

type ElemInpInput struct {
	Input Name
	Next  Element
}

func (e *ElemInpInput) Type() ElementType {
	return ElemTypInpInput
}
