package pifra

import (
	"bytes"
	"container/list"
	"fmt"
	"sort"
	"strconv"
	"text/template"
)

type Lts struct {
	States      map[int]Configuration
	Transitions []Transition

	RegSizeReached map[int]bool

	StatesExplored  int
	StatesGenerated int
}

type Transition struct {
	Source      int
	Destination int
	Label       Label
}

type VertexTemplate struct {
	State  string
	Config string
	Layout string
}

type EdgeTemplate struct {
	Source      string
	Destination string
	Label       string
}

var a4GVLayout = []byte(`
    size="8.3,11.7!";
    ratio="fill";
    margin=0;
    rankdir = TB;
`)

var gvLayout string

func explore(root Configuration) Lts {
	// Visited states.
	visited := make(map[string]int)
	// Encountered transitions.
	trnsSeen := make(map[Transition]bool)
	// Track which states have reached the register size.
	regSizeReached := make(map[int]bool)
	// LTS states.
	states := make(map[int]Configuration)
	// LTS transitions.
	var trns []Transition
	// State ID.
	var stateId int

	applyStructrualCongruence(root)
	rootKey := getConfigurationKey(root)
	visited[rootKey] = stateId
	states[stateId] = root
	stateId++

	queue := list.New()
	queue.PushBack(root)
	dequeue := func() Configuration {
		c := queue.Front()
		queue.Remove(c)
		return c.Value.(Configuration)
	}

	var statesExplored int
	var statesGenerated int

	// BFS traversal state exploration.
	for queue.Len() > 0 && statesExplored < maxStatesExplored {
		state := dequeue()

		srcId := visited[getConfigurationKey(state)]

		if len(state.Registers.Registers) > registerSize {
			regSizeReached[srcId] = true
		} else {
			confs := trans(state)
			for _, conf := range confs {
				statesGenerated++
				applyStructrualCongruence(conf)
				dstKey := getConfigurationKey(conf)
				if _, ok := visited[dstKey]; !ok {
					visited[dstKey] = stateId
					states[stateId] = conf
					stateId++
					queue.PushBack(conf)
				}
				trn := Transition{
					Source:      srcId,
					Destination: visited[dstKey],
					Label:       conf.Label,
				}
				if !trnsSeen[trn] {
					trnsSeen[trn] = true
					trns = append(trns, trn)
				}
			}
		}

		statesExplored++
	}

	return Lts{
		States:          states,
		Transitions:     trns,
		RegSizeReached:  regSizeReached,
		StatesExplored:  statesExplored,
		StatesGenerated: statesGenerated,
	}
}

func generateGraphVizFile(lts Lts, outputStateNo bool) []byte {
	vertices := lts.States
	edges := lts.Transitions

	var buffer bytes.Buffer

	gvl := ""
	if gvLayout != "" {
		gvl = "\n    " + gvLayout + "\n"
	}
	buffer.WriteString("digraph {" + gvl + "\n")

	var ids []int
	for id := range vertices {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		conf := vertices[id]

		var config string
		if outputStateNo {
			config = "s" + strconv.Itoa(id)
		} else {
			config = prettyPrintRegister(conf.Registers) + " ⊢\n" + PrettyPrintAst(conf.Process)
		}

		var layout string
		if id == 0 {
			layout = layout + "peripheries=2,"
		}
		if lts.RegSizeReached[id] {
			layout = layout + "peripheries=3,"
		}

		vertex := VertexTemplate{
			State:  "s" + strconv.Itoa(id),
			Config: config,
			Layout: layout,
		}
		var tmpl *template.Template
		tmpl, _ = template.New("todos").Parse("    {{.State}} [{{.Layout}}label=\"{{.Config}}\"]\n")
		tmpl.Execute(&buffer, vertex)
	}

	buffer.WriteString("\n")

	for _, edge := range edges {
		edg := EdgeTemplate{
			Source:      "s" + strconv.Itoa(edge.Source),
			Destination: "s" + strconv.Itoa(edge.Destination),
			Label:       prettyPrintGraphLabel(edge.Label),
		}
		tmpl, _ := template.New("todos").Parse("    {{.Source}} -> {{.Destination}} [label=\"{{ .Label}}\"]\n")
		tmpl.Execute(&buffer, edg)
	}

	buffer.WriteString("}\n")

	var output bytes.Buffer
	buffer.WriteTo(&output)
	return output.Bytes()
}

func prettyPrintGraphLabel(label Label) string {
	if label.Symbol.Type == SymbolTypTau {
		return "τ"
	}
	return prettyPrintGraphSymbol(label.Symbol) + prettyPrintGraphSymbol(label.Symbol2)
}

func prettyPrintGraphSymbol(symbol Symbol) string {
	s := symbol.Value
	switch symbol.Type {
	case SymbolTypInput:
		return strconv.Itoa(s) + " "
	case SymbolTypOutput:
		return strconv.Itoa(s) + "' "
	case SymbolTypFreshInput:
		return strconv.Itoa(s) + "●"
	case SymbolTypFreshOutput:
		return strconv.Itoa(s) + "⊛"
	case SymbolTypTau:
		return "τ"
	case SymbolTypKnown:
		return strconv.Itoa(s)
	}
	return ""
}

func generateGraphVizTexFile(lts Lts, outputStateNo bool) []byte {
	vertices := lts.States
	edges := lts.Transitions

	var buffer bytes.Buffer

	gvl := ""
	if gvLayout != "" {
		gvl = "\n    " + gvLayout + "\n"
	}
	buffer.WriteString("digraph {" + gvl + "\n")

	buffer.WriteString(`    d2toptions="--format tikz --crop --autosize --nominsize";`)
	buffer.WriteString("\n")
	buffer.WriteString(`    d2tdocpreamble="\usepackage{amssymb}";`)
	buffer.WriteString("\n\n")

	var ids []int
	for id := range vertices {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		conf := vertices[id]

		var config string
		if outputStateNo {
			config = "s_{" + strconv.Itoa(id) + "}"
		} else {
			config = `\begin{matrix} ` +
				prettyPrintTexRegister(conf.Registers) +
				` \vdash \\ ` +
				prettyPrintTexAst(conf.Process) +
				` \end{matrix}`
		}

		var layout string
		if id == 0 {
			layout = layout + `style="double",`
		}
		if lts.RegSizeReached[id] {
			layout = layout + `style="thick",`
		}

		vertex := VertexTemplate{
			State:  "s" + strconv.Itoa(id),
			Config: config,
			Layout: layout,
		}
		var tmpl *template.Template
		tmpl, _ = template.New("todos").Parse("    {{.State}} [{{.Layout}}texlbl=\"${{.Config}}$\"]\n")
		tmpl.Execute(&buffer, vertex)
	}

	buffer.WriteString("\n")

	for _, edge := range edges {
		edg := EdgeTemplate{
			Source:      "s" + strconv.Itoa(edge.Source),
			Destination: "s" + strconv.Itoa(edge.Destination),
			Label:       prettyPrintTexGraphLabel(edge.Label),
		}
		tmpl, _ := template.New("todos").Parse(
			"    {{.Source}} -> {{.Destination}} [label=\"\",texlbl=\"${{.Label}}$\"]\n")
		tmpl.Execute(&buffer, edg)
	}

	buffer.WriteString("}\n")

	var output bytes.Buffer
	buffer.WriteTo(&output)
	return output.Bytes()
}

func prettyPrintTexRegister(register Registers) string {
	str := `\{`
	labels := register.Labels()
	reg := register.Registers

	for i, label := range labels {
		if i == len(labels)-1 {
			str = str + "(" + strconv.Itoa(label) + "," + getTexName(reg[label]) + ")"
		} else {
			str = str + "(" + strconv.Itoa(label) + "," + getTexName(reg[label]) + "),"
		}
	}
	return str + `\}`
}

func getTexName(name string) string {
	if string(name[0]) == "#" {
		return "a" + "_{" + name[1:] + "}"
	}
	if string(name[0]) == "&" {
		return "x" + "_{" + name[1:] + "}"
	}
	if string(name[0]) == "_" {
		return name[1:]
	}
	return name
}

// PrettyPrintAst returns a string containing the pi-calculus syntax of the AST.
func prettyPrintTexAst(elem Element) string {
	return prettyPrintTexAstAcc(elem, "")
}

func prettyPrintTexAstAcc(elem Element, str string) string {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
		str += "0"
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		str += fmt.Sprintf(`\bar{%s} \langle %s \rangle . `,
			getTexName(outElem.Channel.Name), getTexName(outElem.Output.Name))
		return prettyPrintTexAstAcc(outElem.Next, str)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		str += fmt.Sprintf(`%s ( %s ) . `,
			getTexName(inpElem.Channel.Name), getTexName(inpElem.Input.Name))
		return prettyPrintTexAstAcc(inpElem.Next, str)
	case ElemTypMatch:
		matchElem := elem.(*ElemEquality)
		if matchElem.Inequality {
			str += fmt.Sprintf(`\lbrack %s \neq %s \rbrack . `,
				getTexName(matchElem.NameL.Name), getTexName(matchElem.NameR.Name))
		} else {
			str += fmt.Sprintf(`\lbrack %s = %s \rbrack . `,
				getTexName(matchElem.NameL.Name), getTexName(matchElem.NameR.Name))
		}
		return prettyPrintTexAstAcc(matchElem.Next, str)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		str += fmt.Sprintf(`\nu %s . `,
			getTexName(resElem.Restrict.Name))
		return prettyPrintTexAstAcc(resElem.Next, str)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		left := prettyPrintTexAstAcc(sumElem.ProcessL, "")
		right := prettyPrintTexAstAcc(sumElem.ProcessR, "")
		str += fmt.Sprintf(`( %s + %s )`, left, right)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		left := prettyPrintTexAstAcc(parElem.ProcessL, "")
		right := prettyPrintTexAstAcc(parElem.ProcessR, "")
		str += fmt.Sprintf(`( %s \mid %s )`, left, right)
	case ElemTypProcess:
		pcsElem := elem.(*ElemProcess)
		if len(pcsElem.Parameters) == 0 {
			str = str + pcsElem.Name
		} else {
			params := "("
			for i, param := range pcsElem.Parameters {
				if i == len(pcsElem.Parameters)-1 {
					params = params + getTexName(param.Name) + ")"
				} else {
					params = params + getTexName(param.Name) + ", "
				}
			}
			str = str + pcsElem.Name + params
		}
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return prettyPrintTexAstAcc(rootElem.Next, str)
	}
	return str
}

func prettyPrintTexGraphLabel(label Label) string {
	if label.Symbol.Type == SymbolTypTau {
		return `\tau`
	}
	return prettyPrintTexGraphSymbol(label.Symbol) + ` \, ` + prettyPrintTexGraphSymbol(label.Symbol2)
}

func prettyPrintTexGraphSymbol(symbol Symbol) string {
	s := symbol.Value
	switch symbol.Type {
	case SymbolTypInput:
		return strconv.Itoa(s)
	case SymbolTypOutput:
		return `\bar{` + strconv.Itoa(s) + `}`
	case SymbolTypFreshInput:
		return strconv.Itoa(s) + `^{\bullet}`
	case SymbolTypFreshOutput:
		return strconv.Itoa(s) + `^{\circledast}`
	case SymbolTypTau:
		return `\tau`
	case SymbolTypKnown:
		return strconv.Itoa(s)
	}
	return ""
}

func generatePrettyLts(lts Lts) []byte {
	vertices := lts.States
	edges := lts.Transitions

	// When there is no root state.
	if _, ok := vertices[0]; !ok {
		return []byte{}
	}
	var buffer bytes.Buffer

	root := vertices[0]

	rootR := ""
	if lts.RegSizeReached[0] {
		rootR = "+"
	}

	rootString := "s0" + rootR + " = " +
		prettyPrintRegister(root.Registers) + " |- " + PrettyPrintAst(root.Process)
	buffer.WriteString(rootString)

	// Prevent extraneous new line if there are no edges.
	if len(edges) != 0 {
		buffer.WriteString("\n")
	}

	for i, edge := range edges {
		vertex := vertices[edge.Destination]
		srcR := ""
		if lts.RegSizeReached[edge.Source] {
			srcR = "+"
		}
		dstR := ""
		if lts.RegSizeReached[edge.Destination] {
			dstR = "+"
		}
		transString := "s" + strconv.Itoa(edge.Source) + srcR + "  " +
			prettyPrintLabel(edge.Label) + "  s" + strconv.Itoa(edge.Destination) + dstR + " = " +
			prettyPrintRegister(vertex.Registers) + " |- " + PrettyPrintAst(vertex.Process)
		buffer.WriteString(transString)

		// Prevent extraneous new line at last edge.
		if i != len(edges)-1 {
			buffer.WriteString("\n")
		}
	}

	var output bytes.Buffer
	buffer.WriteTo(&output)
	return output.Bytes()
}

// PrettyPrintConfiguration returns a pretty printed string of the configuration.
func PrettyPrintConfiguration(conf Configuration) string {
	return prettyPrintLabel(conf.Label) + " -> " + prettyPrintRegister(conf.Registers) + " ¦- " +
		PrettyPrintAst(conf.Process)

}

func prettyPrintRegister(register Registers) string {
	str := "{"
	labels := register.Labels()
	reg := register.Registers

	for i, label := range labels {
		if i == len(labels)-1 {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + ")"
		} else {
			str = str + "(" + strconv.Itoa(label) + "," + reg[label] + "),"
		}
	}
	return str + "}"
}

func prettyPrintLabel(label Label) string {
	if label.Symbol.Type == SymbolTypTau {
		return "t   "
	}
	return prettyPrintSymbol(label.Symbol) + prettyPrintSymbol(label.Symbol2)
}

func prettyPrintSymbol(symbol Symbol) string {
	s := symbol.Value
	switch symbol.Type {
	case SymbolTypInput:
		return strconv.Itoa(s) + " "
	case SymbolTypOutput:
		return strconv.Itoa(s) + "'"
	case SymbolTypFreshInput:
		return strconv.Itoa(s) + "*"
	case SymbolTypFreshOutput:
		return strconv.Itoa(s) + "^"
	case SymbolTypTau:
		return "t   "
	case SymbolTypKnown:
		return strconv.Itoa(s) + " "
	}
	return ""
}
