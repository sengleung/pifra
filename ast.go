package main

import (
	"strconv"
)

var boundNameIndex int

func doAlphaConversion(elem Element) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		doAlphaConversion(elem.(*ElemOutput).Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Input.Type != Bound {
			boundName := inpElem.Input.Name
			newName := generateBoundName(boundName)
			inpElem.Input = Name{
				Name: newName,
				Type: Bound,
			}
			substituteNamesInput(inpElem.Next, boundName, newName)
		}
		doAlphaConversion(inpElem.Next)
	case ElemTypMatch:
		doAlphaConversion(elem.(*ElemMatch).Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		boundName := resElem.Restrict.Name
		newName := generateBoundName(boundName)
		resElem.Restrict = Name{
			Name: newName,
			Type: Bound,
		}
		substituteNamesRestriction(resElem.Next, boundName, newName)
		doAlphaConversion(resElem.Next)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		doAlphaConversion(sumElem.ProcessL)
		doAlphaConversion(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		doAlphaConversion(parElem.ProcessL)
		doAlphaConversion(parElem.ProcessR)
	case ElemTypProcess:
	case ElemTypProcessConstants:
	}
}

func generateBoundName(namePrefix string) string {
	name := namePrefix + "_" + strconv.Itoa(boundNameIndex)
	boundNameIndex = boundNameIndex + 1
	return name
}

func substituteNamesInput(elem Element, boundName string, newName string) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel.Name == boundName {
			outElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if outElem.Output.Name == boundName {
			outElem.Output = Name{
				Name: newName,
				Type: Bound,
			}
		}
		substituteNamesInput(outElem.Next, boundName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel.Name == boundName {
			inpElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if inpElem.Input.Name != boundName {
			substituteNamesInput(inpElem.Next, boundName, newName)
		}
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		substituteNamesInput(matchElem.Next, boundName, newName)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Restrict.Name != boundName {
			substituteNamesInput(resElem.Next, boundName, newName)
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		substituteNamesInput(sumElem.ProcessL, boundName, newName)
		substituteNamesInput(sumElem.ProcessR, boundName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		substituteNamesInput(parElem.ProcessL, boundName, newName)
		substituteNamesInput(parElem.ProcessR, boundName, newName)
	case ElemTypProcess:
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		for i, param := range pcsElem.Parameters {
			if param == boundName {
				pcsElem.Parameters[i] = newName
			}
		}
	}
}

func substituteNamesRestriction(elem Element, boundName string, newName string) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel.Name == boundName {
			outElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if outElem.Output.Name == boundName {
			outElem.Output = Name{
				Name: newName,
				Type: Bound,
			}
		}
		substituteNamesRestriction(outElem.Next, boundName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel.Name == boundName {
			inpElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if inpElem.Input.Name == boundName {
			inpElem.Input = Name{
				Name: newName,
				Type: Bound,
			}
		}
		substituteNamesRestriction(inpElem.Next, boundName, newName)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		substituteNamesRestriction(matchElem.Next, boundName, newName)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Restrict.Name != boundName {
			substituteNamesRestriction(resElem.Next, boundName, newName)
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		substituteNamesRestriction(sumElem.ProcessL, boundName, newName)
		substituteNamesRestriction(sumElem.ProcessR, boundName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		substituteNamesRestriction(parElem.ProcessL, boundName, newName)
		substituteNamesRestriction(parElem.ProcessR, boundName, newName)
	case ElemTypProcess:
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		for i, param := range pcsElem.Parameters {
			if param == boundName {
				pcsElem.Parameters[i] = newName
			}
		}
	}
}
