package pifra

import "sort"

func applyStructrualCongruence(conf Configuration) {
	removeNilRestrictions(conf.Process)
	sortRes(conf.Process)
}

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + PrettyPrintAst(conf.Process)
}

func removeNilRestrictions(elem Element) {
	elem = rmNilRes(elem)
}

func rmNilRes(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = rmNilRes(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = rmNilRes(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = rmNilRes(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = rmNilRes(resElem.Next)
		if resElem.Next.Type() == ElemTypNil {
			return &ElemNil{}
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = rmNilRes(sumElem.ProcessL)
		sumElem.ProcessR = rmNilRes(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.ProcessL = rmNilRes(parElem.ProcessL)
		parElem.ProcessR = rmNilRes(parElem.ProcessR)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = rmNilRes(rootElem.Next)
	}
	return elem
}

func sortRes(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = sortRes(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = sortRes(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = sortRes(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resNames, lastElem := getRes(resElem, []Name{})
		sort.Slice(resNames, func(i, j int) bool {
			return resNames[i].Name < resNames[j].Name
		})
		head := &ElemRestriction{
			Restrict: resNames[0],
		}
		prev := head
		for i := 1; i < len(resNames); i++ {
			cur := &ElemRestriction{
				Restrict: resNames[i],
			}
			prev.Next = cur
			prev = cur
		}
		prev.Next = lastElem
		return head
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = sortRes(sumElem.ProcessL)
		sumElem.ProcessR = sortRes(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.ProcessL = sortRes(parElem.ProcessL)
		parElem.ProcessR = sortRes(parElem.ProcessR)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = sortRes(rootElem.Next)
	}
	return elem
}

func getRes(elem Element, names []Name) ([]Name, Element) {
	if elem.Type() == ElemTypRestriction {
		resElem := elem.(*ElemRestriction)
		return getRes(resElem.Next, append(names, resElem.Restrict))
	}
	return names, elem
}
