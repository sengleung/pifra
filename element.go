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

type ElemSetType int

const (
	ElemSetReg ElemSetType = iota
	ElemSetOut
	ElemSetInp
)

type Element interface {
	Type() ElementType
}

type ElemNil struct {
	SetType ElemSetType
}

func (e *ElemNil) Type() ElementType {
	return ElemTypNil
}

type ElemOutput struct {
	SetType ElemSetType
	Channel Name
	Output  Name
	Next    Element
}

func (e *ElemOutput) Type() ElementType {
	return ElemTypOutput
}

type ElemInput struct {
	SetType ElemSetType
	Channel Name
	Input   Name
	Next    Element
}

func (e *ElemInput) Type() ElementType {
	return ElemTypInput
}

type ElemMatch struct {
	SetType ElemSetType
	NameL   Name
	NameR   Name
	Next    Element
}

func (e *ElemMatch) Type() ElementType {
	return ElemTypMatch
}

type ElemRestriction struct {
	SetType  ElemSetType
	Restrict Name
	Next     Element
}

func (e *ElemRestriction) Type() ElementType {
	return ElemTypRestriction
}

type ElemSum struct {
	SetType  ElemSetType
	ProcessL Element
	ProcessR Element
}

func (e *ElemSum) Type() ElementType {
	return ElemTypSum
}

type ElemParallel struct {
	SetType  ElemSetType
	ProcessL Element
	ProcessR Element
}

func (e *ElemParallel) Type() ElementType {
	return ElemTypParallel
}

type ElemProcess struct {
	SetType    ElemSetType
	Name       string
	Parameters []Name
}

func (e *ElemProcess) Type() ElementType {
	return ElemTypProcess
}

type ElemOutOutput struct {
	SetType ElemSetType
	Output  Name
	Next    Element
}

func (e *ElemOutOutput) Type() ElementType {
	return ElemTypOutOutput
}

type ElemInpInput struct {
	SetType ElemSetType
	Input   Name
	Next    Element
}

func (e *ElemInpInput) Type() ElementType {
	return ElemTypInpInput
}

type ElemRoot struct {
	SetType ElemSetType
	Next    Element
}

func (e *ElemRoot) Type() ElementType {
	return ElemTypRoot
}
