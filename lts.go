package pifra

import (
	"bytes"
	"sort"
	"strconv"
	"text/template"
)

type GraphEdge struct {
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

func generateGraphVizFile(vertices map[int]Configuration, edges []GraphEdge) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(`digraph {
	size="8.3,11.7!";
    ratio="fill";
    margin=0;
    rankdir = TB;

`)

	var ids []int
	for id := range vertices {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		conf := vertices[id]
		vertex := VertexTemplate{
			State:  "s" + strconv.Itoa(id),
			Config: prettyPrintRegister(conf.Register) + " ⊢\n" + PrettyPrintAst(conf.Process),
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
		return prettyPrintGraphSymbol(label.Symbol) + " "
	}
	return prettyPrintGraphSymbol(label.Symbol) + prettyPrintGraphSymbol(label.Symbol2) + " "
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
