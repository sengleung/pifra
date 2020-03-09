package pifra

import (
	"bytes"
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
			config = prettyPrintRegister(conf.Register) + " ⊢\n" + PrettyPrintAst(conf.Process)
		}
		vertex := VertexTemplate{
			State:  "s" + strconv.Itoa(id),
			Config: config,
		}
		var tmpl *template.Template
		if id == 0 {
			tmpl, _ = template.New("todos").Parse("    {{.State}} [peripheries=2,label=\"{{.Config}}\"]\n")
		} else {
			tmpl, _ = template.New("todos").Parse("    {{.State}} [label=\"{{.Config}}\"]\n")
		}
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
	if !label.Double {
		return prettyPrintGraphSymbol(label.Symbol)
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
		return strconv.Itoa(s) + "● "
	case SymbolTypFreshOutput:
		return strconv.Itoa(s) + "⊛ "
	case SymbolTypTau:
		return "τ "
	case SymbolTypKnown:
		return strconv.Itoa(s) + " "
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
		prettyPrintRegister(root.Register) + " ¦- " + PrettyPrintAst(root.Process)
	buffer.WriteString(rootString)

	// Prevent extraneous new line if there are no edges.
	if len(edges) != 0 {
		buffer.WriteString("\n\n")
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
			prettyPrintRegister(vertex.Register) + " ¦- " + PrettyPrintAst(vertex.Process)
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
