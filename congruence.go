package pifra

import (
	"sort"
	"strconv"
)

var bnPrefix = "&"
var fnPrefix = "#"

var disableGarbageCollection bool

func applyStructrualCongruence(conf Configuration) {
	if !disableGarbageCollection {
		garbageCollection(conf)
	}

	normaliseNilProc(conf.Process)
	normaliseFreshNames(conf)
	normaliseBoundNames(conf)
	sortSumPar(conf.Process)
}

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + ppCongruentProc(conf.Process)
}

func garbageCollection(conf Configuration) {
	fns := GetAllFreeNames(conf.Process)
	freshNames := make(map[string]bool)
	for _, freshName := range fns {
		freshNames[freshName] = true
	}

	for label, name := range conf.Register.Register {
		if !freshNames[name] {
			delete(conf.Register.Register, label)
		}
	}
}

func normaliseFreshNames(conf Configuration) {
	fni := 1
	genFn := func(usedNames map[string]bool) string {
		fn := fnPrefix + strconv.Itoa(fni)
		for usedNames[fn] {
			fni = fni + 1
			fn = fnPrefix + strconv.Itoa(fni)
		}
		fni = fni + 1
		return fn
	}

	usedNames := make(map[string]bool)

	labels := conf.Register.Labels()
	for _, label := range labels {
		name := conf.Register.GetName(label)
		usedNames[name] = true
	}

	for _, label := range labels {
		name := conf.Register.GetName(label)
		if string(name[0]) == bnPrefix {
			fn := genFn(usedNames)
			subName(conf.Process, Name{
				Name: name,
			}, Name{
				Name: fn,
			})
			conf.Register.Register[label] = fn
		}
	}
}

func normaliseBoundNames(conf Configuration) {
	bni := 1
	oldNames := make(map[string]string)

	genBn := func(oldName string) string {
		if newName, ok := oldNames[oldName]; ok {
			return newName
		}
		newName := bnPrefix + strconv.Itoa(bni)
		bni = bni + 1
		oldNames[oldName] = newName
		return newName
	}

	var normaliseBn func(elem Element)
	normaliseBn = func(elem Element) {
		elemTyp := elem.Type()
		switch elemTyp {
		case ElemTypNil:
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			if outElem.Channel.Type == Bound {
				outElem.Channel.Name = genBn(outElem.Channel.Name)
			}
			if outElem.Output.Type == Bound {
				outElem.Output.Name = genBn(outElem.Output.Name)
			}
			normaliseBn(outElem.Next)
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			if inpElem.Channel.Type == Bound {
				inpElem.Channel.Name = genBn(inpElem.Channel.Name)
			}
			if inpElem.Input.Type == Bound {
				inpElem.Input.Name = genBn(inpElem.Input.Name)
			}
			normaliseBn(inpElem.Next)
		case ElemTypMatch:
			matchElem := elem.(*ElemEquality)
			if matchElem.NameL.Type == Bound {
				matchElem.NameL.Name = genBn(matchElem.NameL.Name)
			}
			if matchElem.NameR.Type == Bound {
				matchElem.NameR.Name = genBn(matchElem.NameR.Name)
			}
			normaliseBn(matchElem.Next)
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			normaliseBn(resElem.Next)
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			normaliseBn(sumElem.ProcessL)
			normaliseBn(sumElem.ProcessR)
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			normaliseBn(parElem.ProcessL)
			normaliseBn(parElem.ProcessR)
		case ElemTypProcess:
			procElem := elem.(*ElemProcess)
			for i, param := range procElem.Parameters {
				if param.Type == Bound {
					procElem.Parameters[i].Name = genBn(param.Name)
				}
			}
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			normaliseBn(rootElem.Next)
		}
	}

	var normaliseBnRes func(elem Element)
	normaliseBnRes = func(elem Element) {
		elemTyp := elem.Type()
		switch elemTyp {
		case ElemTypNil:
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			normaliseBnRes(outElem.Next)
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			normaliseBnRes(inpElem.Next)
		case ElemTypMatch:
			matchElem := elem.(*ElemEquality)
			normaliseBnRes(matchElem.Next)
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			if resElem.Restrict.Type == Bound {
				resElem.Restrict.Name = genBn(resElem.Restrict.Name)
			}
			normaliseBnRes(resElem.Next)
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			normaliseBnRes(sumElem.ProcessL)
			normaliseBnRes(sumElem.ProcessR)
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			normaliseBnRes(parElem.ProcessL)
			normaliseBnRes(parElem.ProcessR)
		case ElemTypProcess:
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			normaliseBnRes(rootElem.Next)
		}
	}

	// Rename bound names, skipping restrictions.
	normaliseBn(conf.Process)
	// Rename bound names in restrictions.
	normaliseBnRes(conf.Process)

	// Rename bound names in register.
	for label, name := range conf.Register.Register {
		if newName, ok := oldNames[name]; ok {
			conf.Register.Register[label] = newName
		}
	}
}

func normaliseNilProc(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = normaliseNilProc(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = normaliseNilProc(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemEquality)
		matchElem.Next = normaliseNilProc(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = normaliseNilProc(resElem.Next)
		if resElem.Next.Type() == ElemTypNil {
			return &ElemNil{}
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = normaliseNilProc(sumElem.ProcessL)
		sumElem.ProcessR = normaliseNilProc(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.ProcessL = normaliseNilProc(parElem.ProcessL)
		parElem.ProcessR = normaliseNilProc(parElem.ProcessR)
		if parElem.ProcessL.Type() == ElemTypNil {
			return parElem.ProcessR
		}
		if parElem.ProcessR.Type() == ElemTypNil {
			return parElem.ProcessL
		}
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = normaliseNilProc(rootElem.Next)
	}
	return elem
}

func sortSumPar(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = sortSumPar(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = sortSumPar(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemEquality)
		matchElem.Next = sortSumPar(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = sortSumPar(resElem.Next)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumChildren := getSum(sumElem)
		for i, child := range sumChildren {
			sumChildren[i] = sortSumPar(child)
		}
		procs := []struct {
			Rank    string
			Process Element
		}{}
		for _, child := range sumChildren {
			procs = append(procs, struct {
				Rank    string
				Process Element
			}{ppCongruentProc(child), child})
		}
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].Rank < procs[j].Rank
		})
		head := &ElemSum{
			ProcessL: procs[0].Process,
		}
		prev := head
		for i := 1; i < len(procs)-1; i++ {
			cur := &ElemSum{
				ProcessL: procs[i].Process,
			}
			prev.ProcessR = cur
			prev = cur
		}
		prev.ProcessR = procs[len(procs)-1].Process
		return head
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parChildren := getPar(parElem)
		for i, child := range parChildren {
			parChildren[i] = sortSumPar(child)
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
			}{ppCongruentProc(child), child})
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
		rootElem.Next = sortSumPar(rootElem.Next)
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

func getSum(elem Element) []Element {
	var sumChildren []Element
	if elem.Type() == ElemTypSum {
		sumElem := elem.(*ElemSum)
		if sumElem.ProcessL.Type() == ElemTypSum {
			sumChildren = append(sumChildren, getSum(sumElem.ProcessL)...)
		} else {
			sumChildren = append(sumChildren, sumElem.ProcessL)
		}
		if sumElem.ProcessR.Type() == ElemTypSum {
			sumChildren = append(sumChildren, getSum(sumElem.ProcessR)...)
		} else {
			sumChildren = append(sumChildren, sumElem.ProcessR)
		}
	}
	return sumChildren
}

func ppCongruentProc(elem Element) string {
	var ppcpAcc func(Element, string) string
	ppcpAcc = func(elem Element, str string) string {
		elemTyp := elem.Type()
		switch elemTyp {
		case ElemTypNil:
			str = str + "0"
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			str = str + outElem.Channel.Name + "'<" + outElem.Output.Name + ">."
			return ppcpAcc(outElem.Next, str)
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			str = str + inpElem.Channel.Name + "(" + inpElem.Input.Name + ")."
			return ppcpAcc(inpElem.Next, str)
		case ElemTypMatch:
			matchElem := elem.(*ElemEquality)
			if matchElem.Inequality {
				str = str + "[" + matchElem.NameL.Name + "!=" + matchElem.NameR.Name + "]"
			} else {
				str = str + "[" + matchElem.NameL.Name + "=" + matchElem.NameR.Name + "]"
			}
			return ppcpAcc(matchElem.Next, str)
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			return ppcpAcc(resElem.Next, str)
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			left := ppcpAcc(sumElem.ProcessL, "")
			right := ppcpAcc(sumElem.ProcessR, "")
			str = str + "(" + left + " + " + right + ")"
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			left := ppcpAcc(parElem.ProcessL, "")
			right := ppcpAcc(parElem.ProcessR, "")
			str = str + "(" + left + " | " + right + ")"
		case ElemTypProcess:
			pcsElem := elem.(*ElemProcess)
			if len(pcsElem.Parameters) == 0 {
				str = str + pcsElem.Name
			} else {
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
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			return ppcpAcc(rootElem.Next, str)
		}
		return str
	}
	return ppcpAcc(elem, "")
}
