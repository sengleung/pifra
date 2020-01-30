package main

import "strconv"

var boundNameIndex int

func doAlphaConversion(elem Element) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		doAlphaConversion(elem.(*ElemOutput).Next)
	case ElemTypInput:
		doAlphaConversion(elem.(*ElemInput).Next)
	case ElemTypMatch:
		doAlphaConversion(elem.(*ElemMatch).Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		boundName := resElem.Name
		newName := generateBoundName()
		resElem.Name = newName
		substituteNames(resElem.Next, boundName, newName)
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

func generateBoundName() string {
	name := "bn_" + strconv.Itoa(boundNameIndex)
	boundNameIndex = boundNameIndex + 1
	return name
}

func substituteNames(elem Element, boundName string, newName string) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel == boundName {
			outElem.Channel = newName
		}
		if outElem.Output == boundName {
			outElem.Output = newName
		}
		substituteNames(outElem.Next, boundName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel == boundName {
			inpElem.Channel = newName
		}
		if inpElem.Input == boundName {
			inpElem.Input = newName
		}
		substituteNames(inpElem.Next, boundName, newName)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		if matchElem.NameL == boundName {
			matchElem.NameR = newName
		}
		substituteNames(matchElem.Next, boundName, newName)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Name != boundName {
			substituteNames(resElem.Next, boundName, newName)
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		substituteNames(sumElem.ProcessL, boundName, newName)
		substituteNames(sumElem.ProcessR, boundName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		substituteNames(parElem.ProcessL, boundName, newName)
		substituteNames(parElem.ProcessR, boundName, newName)
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
