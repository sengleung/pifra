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
	ElemTypInpInput

	ElemTypRoot
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

type Elem struct {
	Parent Element
}

type Element interface {
	Type() ElementType
}

type ElemNil struct {
	Elem
}

func (e *ElemNil) Type() ElementType {
	return ElemTypNil
}

type ElemOutput struct {
	Elem
	Channel Name
	Output  Name
	Next    Element
}

func (e *ElemOutput) Type() ElementType {
	return ElemTypOutput
}

type ElemInput struct {
	Elem
	Channel Name
	Input   Name
	Next    Element
}

func (e *ElemInput) Type() ElementType {
	return ElemTypInput
}

type ElemMatch struct {
	Elem
	NameL Name
	NameR Name
	Next  Element
}

func (e *ElemMatch) Type() ElementType {
	return ElemTypMatch
}

type ElemRestriction struct {
	Elem
	Restrict Name
	Next     Element
}

func (e *ElemRestriction) Type() ElementType {
	return ElemTypRestriction
}

type ElemSum struct {
	Elem
	ProcessL Element
	ProcessR Element
}

func (e *ElemSum) Type() ElementType {
	return ElemTypSum
}

type ElemParallel struct {
	Elem
	ProcessL Element
	ProcessR Element
}

func (e *ElemParallel) Type() ElementType {
	return ElemTypParallel
}

type ElemProcess struct {
	Elem
	Name string
}

func (e *ElemProcess) Type() ElementType {
	return ElemTypProcess
}

type ElemProcessConstants struct {
	Elem
	Name       string
	Parameters []Name
}

func (e *ElemProcessConstants) Type() ElementType {
	return ElemTypProcessConstants
}

type ElemOutOutput struct {
	Elem
	Output Name
	Next   Element
}

func (e *ElemOutOutput) Type() ElementType {
	return ElemTypOutOutput
}

type ElemInpInput struct {
	Elem
	Input Name
	Next  Element
}

func (e *ElemInpInput) Type() ElementType {
	return ElemTypInpInput
}

type Root struct {
	Elem
	Next Element
}

func (e *Root) Type() ElementType {
	return ElemTypRoot
}
