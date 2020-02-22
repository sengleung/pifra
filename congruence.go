package pifra

import "sort"

func applyStructrualCongruence(conf Configuration) {
	rmNilRes(conf.Process)
	rmNilPar(conf.Process)
	sortRes(conf.Process)
	sortPar(conf.Process)
}

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + PrettyPrintAst(conf.Process)
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

func rmNilPar(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = rmNilPar(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = rmNilPar(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = rmNilPar(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = rmNilPar(resElem.Next)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = rmNilPar(sumElem.ProcessL)
		sumElem.ProcessR = rmNilPar(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.ProcessL = rmNilPar(parElem.ProcessL)
		parElem.ProcessR = rmNilPar(parElem.ProcessR)
		if parElem.ProcessL.Type() == ElemTypNil {
			return parElem.ProcessR
		}
		if parElem.ProcessR.Type() == ElemTypNil {
			return parElem.ProcessL
		}
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = rmNilPar(rootElem.Next)
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

func sortPar(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = sortPar(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = sortPar(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = sortPar(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = sortPar(resElem.Next)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = sortPar(sumElem.ProcessL)
		sumElem.ProcessR = sortPar(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parChildren := getPar(parElem)
		for _, child := range parChildren {
			sortPar(child)
		}
		// Size of procs is minimum of 2.
		procs := []struct {
			Rank    string
			Process Element
		}{}
		for _, child := range parChildren {
			procs = append(procs, struct {
				Rank    string
				Process Element
			}{PrettyPrintAst(child), child})
		}
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].Rank < procs[j].Rank
		})
		head := &ElemParallel{
			ProcessL: procs[0].Process,
		}
		prev := head
		for i := 1; i < len(procs)-1; i++ {
			cur := &ElemParallel{
				ProcessL: procs[i].Process,
			}
			prev.ProcessR = cur
			prev = cur
		}
		prev.ProcessR = procs[len(procs)-1].Process
		return head
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = sortPar(rootElem.Next)
	}
	return elem
}

func getPar(elem Element) []Element {
	var parChildren []Element
	if elem.Type() == ElemTypParallel {
		parElem := elem.(*ElemParallel)
		if parElem.ProcessL.Type() == ElemTypParallel {
			parChildren = append(parChildren, getPar(parElem.ProcessL)...)
		} else {
			parChildren = append(parChildren, parElem.ProcessL)
		}
		if parElem.ProcessR.Type() == ElemTypParallel {
			parChildren = append(parChildren, getPar(parElem.ProcessR)...)
		} else {
			parChildren = append(parChildren, parElem.ProcessR)
		}
	}
	return parChildren
}
