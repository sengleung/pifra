# pifra - Pi-Calculus Fresh-Register Automata

pifra is a tool for generating labelled transition systems (LTS) of pi-calculus models represented by fresh-register automata (FRA).

The LTS is generated according to the xπ-calculus FRA transition relation described in the paper [Fresh-Register Automata](http://www.cs.ox.ac.uk/people/nikos.tzevelekos/FRA_11.pdf) by [Nikos Tzevelekos](http://www.tzevelekos.org/). The LTS is minimised by normalising each explored state based on equivalance rules.

This work was carried out as part of a master's dissertation with [Vasileios Koutavas](https://www.scss.tcd.ie/Vasileios.Koutavas/) at Trinity College Dublin.

## Installation

```
go get -u github.com/sengleung/pifra/pifra
```

`go get` downloads the package and its dependencies to `$GOPATH/src` and creates an executable in `$GOPATH/bin`.

## Command-line interface

```
pifra --help
```

```
pifra generates labelled transition systems (LTS) of
pi-calculus models represented by fresh-register automata.

Usage:
pifra [OPTION...] FILE

Options:
  -n, --max-states int         maximum number of states explored (default 20)
  -r, --max-registers int      maximum number of registers (default is unlimited)
  -d, --disable-gc             disable garbage collection
  -i, --interactive            inspect interactively the LTS in a prompt
  -o, --output string          output the LTS to a file (default format is the Graphviz DOT language)
  -t, --output-tex             output the LTS file with LaTeX labels for use with dot2tex
  -p, --output-pretty          output the LTS file in a pretty-printed format
  -s, --output-states          output state numbers instead of configurations for the Graphviz DOT file
  -l, --output-layout string   layout of the GraphViz DOT file, e.g., "rankdir=TB; margin=0;"
  -q, --quiet                  do not print or output the LTS
  -v, --stats                  print LTS generation statistics
  -h, --help                   show this help message and exit
```

## Pi-calculus models

### Syntax

```
P,Q ::=
      | a(b).P     input
      | <a>'b.P    output
      | [a=b]P     equality
      | [a!=b]P    inequality
      | $a.P       restriction
      | P + Q      summation
      | P | Q      composition
      | p(a)       process
      | 0          inaction

Pdef ::= p(a) = P
```

```
Pdef...
Pundecl
```

### Example models

The below and additional pi-calculus models can be found in `test/`.

`fresh.pi`
```
$x.a'<x>.b'<x>.0 | b(y).0
```

`ping1.pi`
```
P = a(x).x'<x>.0 | P
P
```

`tzevelekos.pi`
```
P(a,b) = a'<b>.$c.P(b,c)
$b.P(a,b)
```

`vk-inf-st3.pi`
```
P(a) =  a(x).$y.( x'<y>.0  |  b(z).[z=y] P(a) )
P(a)
```

`password.pi`
```
GenPass(requestNewPass) = requestNewPass(x). $pass. x'<pass>.0

KeepSecret(requestNewPass) = $p. requestNewPass'<p>. p(pass). ( StoreSecret(pass) | TestSecret(pass) )

StoreSecret(pass) = $secret. pass'<secret>. StoreSecret(pass)

TestSecret(pass) = pub(x). pass(secret). ( TestSecret(pass) + [x=secret] _BAD'<_BAD>.0 )

$requestNewPass. (GenPass(requestNewPass)  |  KeepSecret(requestNewPass))
```

## LTS output

The below and additional LTS outputs can be found in `test/`.

#### Input model

`fresh.pi`

```
$x.a'<x>.b'<x>.0 | b(y).0
```

### Pretty-printed LTS

```
pifra fresh.pi
```

```
s0 = {(1,#1),(2,#2)} |- (#2(&2).0 | $&1.#1'<&1>.#2'<&1>.0)
s0  2 1   s1 = {(1,#1),(2,#2)} |- $&1.#1'<&1>.#2'<&1>.0
s0  2 2   s1 = {(1,#1),(2,#2)} |- $&1.#1'<&1>.#2'<&1>.0
s0  2 3*  s1 = {(1,#1),(2,#2)} |- $&1.#1'<&1>.#2'<&1>.0
s0  1'1^  s2 = {(1,#1),(2,#2)} |- (#2'<#1>.0 | #2(&1).0)
s1  1'1^  s3 = {(1,#1),(2,#2)} |- #2'<#1>.0
s2  2'1   s4 = {(2,#2)} |- #2(&1).0
s2  2 1   s3 = {(1,#1),(2,#2)} |- #2'<#1>.0
s2  2 2   s3 = {(1,#1),(2,#2)} |- #2'<#1>.0
s2  2 3*  s3 = {(1,#1),(2,#2)} |- #2'<#1>.0
s2  t     s5 = {} |- 0
s3  2'1   s5 = {} |- 0
s4  2 2   s5 = {} |- 0
s4  2 1*  s5 = {} |- 0
```

| output    | meaning       |
|-----------|---------------|
| `{} ⊢ P`  | configuration |
| `#1`      | free name     |
| `&1`      | bound name    |
| `1 1`     | known input   |
| `1'1`     | known output  |
| `1 1*`    | fresh input   |
| `1 1^`    | fresh output  |
| `t`       | tau step      |

### GraphViz DOT LTS

The LTS can be outputted as a [GraphViz](https://www.graphviz.org/) DOT graph description language.

```
pifra -o lts.dot fresh.pi
cat lts.dot
```

```
digraph {
    s0 [peripheries=2,label="{(1,#1),(2,#2)} ⊢
(#2(&2).0 | $&1.#1'<&1>.#2'<&1>.0)"]
    s1 [label="{(1,#1),(2,#2)} ⊢
$&1.#1'<&1>.#2'<&1>.0"]
    s2 [label="{(1,#1),(2,#2)} ⊢
(#2'<#1>.0 | #2(&1).0)"]
    s3 [label="{(1,#1),(2,#2)} ⊢
#2'<#1>.0"]
    s4 [label="{(2,#2)} ⊢
#2(&1).0"]
    s5 [label="{} ⊢
0"]

    s0 -> s1 [label="2 1"]
    s0 -> s1 [label="2 2"]
    s0 -> s1 [label="2 3●"]
    s0 -> s2 [label="1' 1⊛"]
    s1 -> s3 [label="1' 1⊛"]
    s2 -> s4 [label="2' 1"]
    s2 -> s3 [label="2 1"]
    s2 -> s3 [label="2 2"]
    s2 -> s3 [label="2 3●"]
    s2 -> s5 [label="τ"]
    s3 -> s5 [label="2' 1"]
    s4 -> s5 [label="2 2"]
    s4 -> s5 [label="2 1●"]
}
```

Using the GraphViz toolset, a visualisation of the LTS can be generated.

```
pifra -o lts.dot fresh.pi
dot -Tpdf -o lts.pdf lts.dot
```

<img src="https://gist.github.com/sengleung/2cb39973c38e28b0fc1d39848cba13d2/raw/34fd15faa0fda23038c9ab2f454d034a3d583fd9/lts.png" width="500">


### GraphViz DOT with LaTeX labels LTS

The LTS can be outputted as a [dot2tex](https://github.com/kjellmf/dot2tex)-compatible DOT file, which uses LaTeX labels.

```
pifra -t -o lts.dot fresh.pi
dot2tex -o lts.tex lts.dot
pdflatex lts.tex
```

<img src="https://gist.github.com/sengleung/2cb39973c38e28b0fc1d39848cba13d2/raw/34fd15faa0fda23038c9ab2f454d034a3d583fd9/lts-tex.png" width="500">

State numbers can be outputted instead of configurations.

```
pifra -st -o lts.dot fresh.pi
dot2tex -o lts.tex lts.dot
pdflatex lts.tex
```

<img src="https://gist.github.com/sengleung/2cb39973c38e28b0fc1d39848cba13d2/raw/34fd15faa0fda23038c9ab2f454d034a3d583fd9/lts-tex-states.png" width="500">
