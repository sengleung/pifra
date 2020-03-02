package pifra

import (
	"sort"
	"strconv"
)

var bnPrefix = "&"
var fnPrefix = "#"

func applyStructrualCongruence(conf Configuration) {
	rmNilRes(conf.Process)
	rmNilPar(conf.Process)
	normaliseFreeNames(conf)
	//scopeRes(conf.Process)
	//sortRes(conf.Process)
	normaliseBoundNames(conf)
	sortSumPar(conf.Process)
}

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + ppCongruentProc(conf.Process)
}

func normaliseFreeNames(conf Configuration) {
	var fni int
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
				Type: Bound,
			}, Name{
				Name: fn,
				Type: Fresh,
			})
			conf.Register.Register[label] = fn
		}
	}
}

func normaliseBoundNames(conf Configuration) {
	var bni int
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
			matchElem := elem.(*ElemMatch)
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
			matchElem := elem.(*ElemMatch)
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

func scopeRes(elem Element) Element {
	switch elem.Type() {
	case ElemTypNil:
	case ElemTypProcess:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		outElem.Next = scopeRes(outElem.Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		inpElem.Next = scopeRes(inpElem.Next)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		matchElem.Next = scopeRes(matchElem.Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		resElem.Next = scopeRes(resElem.Next)
		resName := resElem.Restrict
		if resElem.Next.Type() == ElemTypParallel {
			parElem := resElem.Next.(*ElemParallel)
			if !appearsIn(parElem.ProcessL, resName) {
				parElem.ProcessR = &ElemRestriction{
					Restrict: resName,
					Next:     parElem.ProcessR,
				}
				return parElem
			}
			if !appearsIn(parElem.ProcessR, resName) {
				parElem.ProcessL = &ElemRestriction{
					Restrict: resName,
					Next:     parElem.ProcessL,
				}
				return parElem
			}
		}
		return resElem
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		sumElem.ProcessL = scopeRes(sumElem.ProcessL)
		sumElem.ProcessR = scopeRes(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		parElem.ProcessL = scopeRes(parElem.ProcessL)
		parElem.ProcessR = scopeRes(parElem.ProcessR)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		rootElem.Next = scopeRes(rootElem.Next)
	}
	return elem
}

func appearsIn(elem Element, name Name) bool {
	switch elem.Type() {
	case ElemTypNil:
		return false
	case ElemTypProcess:
		procElem := elem.(*ElemProcess)
		for _, param := range procElem.Parameters {
			if param == name {
				return true
			}
		}
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel == name {
			return true
		}
		if outElem.Output == name {
			return true
		}
		return appearsIn(outElem.Next, name)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel == name {
			return true
		}
		if inpElem.Input == name {
			return true
		}
		return appearsIn(inpElem.Next, name)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == name {
			return true
		}
		if matchElem.NameR == name {
			return true
		}
		return appearsIn(matchElem.Next, name)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Restrict == name {
			return true
		}
		return appearsIn(resElem.Next, name)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		appears := appearsIn(sumElem.ProcessL, name)
		return appears || appearsIn(sumElem.ProcessR, name)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		appears := appearsIn(parElem.ProcessL, name)
		return appears || appearsIn(parElem.ProcessR, name)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return appearsIn(rootElem.Next, name)
	}
	return false
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
		matchElem := elem.(*ElemMatch)
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
			matchElem := elem.(*ElemMatch)
			str = str + "[" + matchElem.NameL.Name + "=" + matchElem.NameR.Name + "]"
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
