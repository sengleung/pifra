package main

import (
	"strconv"

	"github.com/mohae/deepcopy"
)

var boundNameIndex int
var freshNameIndex int

func substituteName(elem Element, oldName Name, newName Name) Element {
	elemCopy := deepcopy.Copy(elem)
	elem = elemCopy.(Element)
	subName(elem, oldName, newName)
	return elem
}

func subName(elem Element, oldName Name, newName Name) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel == oldName {
			outElem.Channel = newName
		}
		if outElem.Output == oldName {
			outElem.Output = newName
		}
		subName(outElem.Next, oldName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel == oldName {
			inpElem.Channel = newName
		}
		if inpElem.Input == oldName {
			inpElem.Input = newName
		}
		subName(inpElem.Next, oldName, newName)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == oldName {
			matchElem.NameL = newName
		}
		if matchElem.NameR == oldName {
			matchElem.NameR = newName
		}
		subName(matchElem.Next, oldName, newName)
	case ElemTypRestriction:
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		subName(sumElem.ProcessL, oldName, newName)
		subName(sumElem.ProcessR, oldName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		subName(parElem.ProcessL, oldName, newName)
		subName(parElem.ProcessR, oldName, newName)
	case ElemTypProcess:
		// TODO
	case ElemTypProcessConstants:
		// TODO
	}
}

func open(elem Element, boundName Name) Element {
	if boundName.Type != Bound {
		return nil
	}
	elemCopy := deepcopy.Copy(elem)
	elem = elemCopy.(Element)
	freshName := Name{
		Name: generateFreshName("fn"),
		Type: Fresh,
	}
	subName(elem, boundName, freshName)
	return elem
}

func close(elem Element, freshName Name) Element {
	if freshName.Type != Fresh {
		return nil
	}
	elemCopy := deepcopy.Copy(elem)
	elem = elemCopy.(Element)
	boundName := Name{
		Name: generateBoundName(freshName.Name),
		Type: Bound,
	}
	subName(elem, freshName, boundName)
	return elem
}

func generateFreshName(namePrefix string) string {
	name := namePrefix + "_" + strconv.Itoa(freshNameIndex)
	freshNameIndex = freshNameIndex + 1
	return name
}

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
		if matchElem.NameL.Name == boundName {
			matchElem.NameL = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if matchElem.NameR.Name == boundName {
			matchElem.NameR = Name{
				Name: newName,
				Type: Bound,
			}
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
			if param.Name == boundName {
				pcsElem.Parameters[i].Name = newName
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
		if matchElem.NameL.Name == boundName {
			matchElem.NameL = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if matchElem.NameR.Name == boundName {
			matchElem.NameR = Name{
				Name: newName,
				Type: Bound,
			}
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
			if param.Name == boundName {
				pcsElem.Parameters[i].Name = newName
			}
		}
	}
}

func prettyPrint(elem Element) string {
	return prettyPrintAcc(elem, "")
}

func prettyPrintAcc(elem Element, str string) string {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
		str = str + "0"
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		str = str + outElem.Channel.Name + "'<" + outElem.Output.Name + ">."
		return prettyPrintAcc(outElem.Next, str)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		str = str + inpElem.Channel.Name + "(" + inpElem.Input.Name + ")."
		return prettyPrintAcc(inpElem.Next, str)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		str = str + "[" + matchElem.NameL.Name + "=" + matchElem.NameL.Name + "]"
		return prettyPrintAcc(matchElem.Next, str)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		str = str + "$" + resElem.Restrict.Name + "."
		return prettyPrintAcc(resElem.Next, str)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		left := prettyPrintAcc(sumElem.ProcessL, "")
		right := prettyPrintAcc(sumElem.ProcessR, "")
		str = str + "(" + left + " + " + right + ")"
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		left := prettyPrintAcc(parElem.ProcessL, "")
		right := prettyPrintAcc(parElem.ProcessR, "")
		str = str + "(" + left + " | " + right + ")"
	case ElemTypProcess:
		procElem := elem.(*ElemProcess)
		str = str + procElem.Name
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		params := "("
		for i, param := range pcsElem.Parameters {
			if i == len(pcsElem.Parameters)-1 {
				params = params + param.Name + ")"
			} else {
				params = params + param.Name + ", "
			}
		}
		str = str + pcsElem.Name + params
	}
	return str
}

func getAllFreshNames(elem Element) []string {
	return getAllFreshNamesAcc(elem, []string{})
}

func getAllFreshNamesAcc(elem Element, freshNames []string) []string {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel.Type == Fresh {
			freshNames = append(freshNames, outElem.Channel.Name)
		}
		if outElem.Output.Type == Fresh {
			freshNames = append(freshNames, outElem.Output.Name)
		}
		return getAllFreshNamesAcc(outElem.Next, freshNames)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel.Type == Fresh {
			freshNames = append(freshNames, inpElem.Channel.Name)
		}
		if inpElem.Input.Type == Fresh {
			freshNames = append(freshNames, inpElem.Input.Name)
		}
		return getAllFreshNamesAcc(inpElem.Next, freshNames)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL.Type == Fresh {
			freshNames = append(freshNames, matchElem.NameL.Name)
		}
		if matchElem.NameR.Type == Fresh {
			freshNames = append(freshNames, matchElem.NameR.Name)
		}
		return getAllFreshNamesAcc(matchElem.Next, freshNames)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Restrict.Type == Fresh {
			freshNames = append(freshNames, resElem.Restrict.Name)
		}
		return getAllFreshNamesAcc(resElem.Next, freshNames)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		freshNames = getAllFreshNamesAcc(sumElem.ProcessL, freshNames)
		freshNames = getAllFreshNamesAcc(sumElem.ProcessR, freshNames)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		freshNames = getAllFreshNamesAcc(parElem.ProcessL, freshNames)
		freshNames = getAllFreshNamesAcc(parElem.ProcessR, freshNames)
	case ElemTypProcess:
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		for _, param := range pcsElem.Parameters {
			if param.Type == Fresh {
				freshNames = append(freshNames, param.Name)
			}
		}
	}
	return freshNames
}

// ConvertToDllAst converts a singly-linked AST to a doubly-linked AST.
func ConvertToDllAst(elem Element) {
	convertToDllAstAcc(elem, nil)
}

func convertToDllAstAcc(elem Element, prev Element) {
	switch elem.Type() {
	case ElemTypNil:
		nilElem := elem.(*ElemNil)
		nilElem.Parent = prev
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Parent = prev
		prev = outElem
		convertToDllAstAcc(outElem.Next, prev)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Parent = prev
		prev = inpElem
		convertToDllAstAcc(inpElem.Next, prev)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Parent = prev
		prev = matchElem
		convertToDllAstAcc(matchElem.Next, prev)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Parent = prev
		prev = resElem
		convertToDllAstAcc(resElem.Next, prev)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.Parent = prev
		prev = sumElem
		convertToDllAstAcc(sumElem.ProcessL, prev)
		convertToDllAstAcc(sumElem.ProcessR, prev)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.Parent = prev
		prev = parElem
		convertToDllAstAcc(parElem.ProcessL, prev)
		convertToDllAstAcc(parElem.ProcessR, prev)
	case ElemTypProcess:
		procElem := elem.(*ElemProcess)
		procElem.Parent = prev
	case ElemTypProcessConstants:
		pcsElem := elem.(*ElemProcessConstants)
		pcsElem.Parent = prev
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		outOutput.Parent = prev
		convertToDllAstAcc(outOutput.Next, prev)
	case ElemTypInpInput:
		inpInput := elem.(*ElemInpInput)
		inpInput.Parent = prev
		convertToDllAstAcc(inpInput.Next, prev)
	}
}
