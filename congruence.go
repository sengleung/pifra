package pifra

func applyStructrualCongruence(conf Configuration) {
	removeNilRestrictions(conf.Process)
}

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + PrettyPrintAst(conf.Process)
}

func removeNilRestrictions(elem Element) {
	elem = rmNilRes(elem)
}

func rmNilRes(elem Element) Element {
	elemTyp := elem.Type()
	switch elemTyp {
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
		if resElem.Next.Type() == ElemTypNil {
			return &ElemNil{}
		}
		resElem.Next = rmNilRes(resElem.Next)
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
