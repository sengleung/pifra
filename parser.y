%{
package pifra

var undeclaredProcs []Element

var curProcParams []string

// All elements
var curElem Element          // Tracks the current element chain

// Process Constants element
var curPconstNames []Name  // Tracks the process constant names

// Sum element
var curSum Element            // Current sum process.
var sumStack []Element        // Sum processes encountered.
var curSumLevel int           // Current sum element level.
var curSumLevelStack []int    // Saves curSumLevel at different bracket levels.
var numSumStack []int         // Saves the maximum curSumLevel at different bracket levels.
                              // Used for knowing how many elements to pop from sumStack.

// Parallel element
var curPar Element            // Current parallel process.
var parStack []Element        // Parallel processes encountered.
var curParLevel int           // Current parallel element level.
var curParLevelStack []int    // Saves curParLevel at different bracket levels.
var numParStack []int         // Saves the maximum curParLevel at different bracket levels.
                              // Used for knowing how many elements to pop from parStack.
%}

%union {
   name string
}

%token <name> NAME
%token NAME
    LBRACKET RBRACKET 
    LANGLE RANGLE
    LSQBRACKET RSQBRACKET
    COMMA
    EQUAL
    VERTBAR
    DOT
    COMMENT
    ZERO
    APOSTROPHE
    DOLLARSIGN
    PLUS
    EXCLAMATION

%nonassoc LOWPREC
%nonassoc LOWER_THAN_LBRACKET
%nonassoc LBRACKET
%nonassoc RSQBRACKET
%right VERTBAR
%right PLUS
%nonassoc DOT
%%

stmts: /* empty */
    | stmts stmt

stmt:
    pconstants_decl
    |
    process_decl
    |
    undecl

pconstants_decl:
    NAME LBRACKET pconst_decl_names EQUAL elem
    {
        // Reverse order of curProcParams
        for i := len(curProcParams)/2-1; i >= 0; i-- {
            j := len(curProcParams)-1-i
            curProcParams[i], curProcParams[j] = curProcParams[j], curProcParams[i]
        }
        name := $1
        DeclaredProcs[name] = DeclaredProcess{
            Process: curElem,
            Parameters: curProcParams,
        }
        curElem = nil
        curProcParams = []string{}

        Log("pconst decl")
    }

pconst_decl_names:
    NAME COMMA pconst_decl_names
    {
        curProcParams = append(curProcParams, $1)
    }
    |
    NAME RBRACKET
    {
        curProcParams = append(curProcParams, $1)
    }

process_decl:
    NAME EQUAL elem
    {
        name := $1
        DeclaredProcs[name] = DeclaredProcess{
            Process: curElem,
            Parameters: []string{},
        }
        curElem = nil

        Log("process")
    }

undecl:
    elem
    {
        undeclaredProcs = append(undeclaredProcs, curElem)
        curElem = nil
    }

elem:
    parentheses
    |
    parallel
    |
    sum
    |
    output
    |
    input
    |
    equality
    |
    inequality
    |
    restriction
    |
    nil
    |
    process
    |
    pconstants

nil:
    ZERO
    {
        Log("nil")
        curElem = &ElemNil{}
    }

output:
    NAME APOSTROPHE LANGLE NAME RANGLE DOT elem
    {
        channel := $1
        output := $4
        outputElem := &ElemOutput{
            Channel: Name{
                Name: channel,
            },
            Output: Name{
                Name: output,
            },
            Next: curElem,
        }
        curElem = outputElem

        Log("out:", channel, output)
    }
    |
    NAME LANGLE NAME RANGLE DOT elem
    {
        channel := $1
        output := $3
        outputElem := &ElemOutput{
            Channel: Name{
                Name: channel,
            },
            Output: Name{
                Name: output,
            },
            Next: curElem,
        }
        curElem = outputElem

        Log("out:", channel, output)
    }

input:
    NAME LBRACKET NAME RBRACKET DOT elem
    {
        channel := $1
        input := $3
        inputElem := &ElemInput{
            Channel: Name{
                Name: channel,
            },
            Input: Name{
                Name: input,
            },
            Next: curElem,
        }
        curElem = inputElem

        Log("inp:", channel, input)
    }

equality:
    LSQBRACKET NAME EQUAL NAME RSQBRACKET elem
    {
        equalityElem := &ElemEquality{
            NameL: Name{
                Name: $2,
            },
            NameR: Name{
                Name: $4,
            },
            Next: curElem,
        }
        curElem = equalityElem
        Log("equality:", $2, $4)
    }

inequality:
    LSQBRACKET NAME EXCLAMATION EQUAL NAME RSQBRACKET elem
    {
        equalityElem := &ElemEquality{
            Inequality: true,
            NameL: Name{
                Name: $2,
            },
            NameR: Name{
                Name: $5,
            },
            Next: curElem,
        }
        curElem = equalityElem
        Log("inequality:", $2, $5)
    }

restriction:
    DOLLARSIGN NAME DOT elem
    {
        resElem := &ElemRestriction{
            Restrict: Name{
                Name: $2,
            },
            Next: curElem,
        }
        curElem = resElem
        Log("new:", $2)
    }

sum: 
    elem PLUS
    {
        // Track the maximum curSumLevel, i.e. no. of sums at this 
        // bracket level.
        if curSumLevel == 0 {
            numSumStack = append(numSumStack, curSumLevel)
        }
        _, numSumStack = pop(numSumStack)
        numSumStack = append(numSumStack, curSumLevel)

        sumStack = append(sumStack, curElem)
        curElem = nil
        curSumLevel = curSumLevel + 1

        Log("+")
    }
    elem
    {
        curSumLevel = curSumLevel - 1
        if curSumLevel == 0 {
            // Create sum element using penultimate element and 
            // terminal element.
            elem := popSumStack()
            sumTerminal := &ElemSum{
                ProcessL: elem,
                ProcessR: curElem,
            }
            curSum = sumTerminal

            // Append sum processes (up to no. of sums at this level) 
            // to form right-leaning sum element tree.
            var numSum int
            numSum, numSumStack = pop(numSumStack)
            for i := 0; i < numSum; i++ {
                elem = popSumStack()
                sumNonTerminal := &ElemSum{
                    ProcessL: elem,
                    ProcessR: curSum,
                }
                curSum = sumNonTerminal
            }
            curElem = curSum
        }
    }

parallel:
    elem VERTBAR
    {
        // Track the maximum curParLevel, i.e. no. of parallels at this 
        // bracket level.
        if curParLevel == 0 {
            numParStack = append(numParStack, curParLevel)
        }
        _, numParStack = pop(numParStack)
        numParStack = append(numParStack, curParLevel)

        parStack = append(parStack, curElem)
        curElem = nil
        curParLevel = curParLevel + 1

        Log("|")
    }
    elem  /* %prec LOWPREC */
    {
        curParLevel = curParLevel - 1
        if curParLevel == 0 {
            // Create parallel element using penultimate element and 
            // terminal element.
            elem := popParStack()
            parTerminal := &ElemParallel{
                ProcessL: elem,
                ProcessR: curElem,
            }
            curPar = parTerminal

            // Append parallel processes (up to no. of parallels at this level) 
            // to form right-leaning parallel element tree.
            var numPar int
            numPar, numParStack = pop(numParStack)
            for i := 0; i < numPar; i++ {
                elem = popParStack()
                parNonTerminal := &ElemParallel{
                    ProcessL: elem,
                    ProcessR: curPar,
                }
                curPar = parNonTerminal
            }
            curElem = curPar
        }
    }

pconstants:
    NAME LBRACKET names
    {
        // Reverse order of curPconstNames
        for i := len(curPconstNames)/2-1; i >= 0; i-- {
            j := len(curPconstNames)-1-i
            curPconstNames[i], curPconstNames[j] = curPconstNames[j], curPconstNames[i]
        }
        name := $1
        pconstElem := &ElemProcess{
            Name: name,
            Parameters: curPconstNames,
        }
        curElem = pconstElem
        curPconstNames = []Name{}
        Log("pconsts:", name)
    }

names:
    NAME COMMA names
    {
        curPconstNames = append(curPconstNames, Name{
            Name: $1,
        })
    }
    |
    NAME RBRACKET
    {
        curPconstNames = append(curPconstNames, Name{
            Name: $1,
        })
    }

process:
    NAME        %prec LOWER_THAN_LBRACKET
    {
        name := $1
        processElem := &ElemProcess{
            Name: name,
        }
        curElem = processElem
        Log("process:", name)
    }

parentheses:
    LBRACKET
    {
        // Sum elements:
        // Save no. of sum on stack.
        curSumLevelStack = append(curSumLevelStack, curSumLevel)
        // Reset no. of sums.
        curSumLevel = 0

        // Parallel elements:
        // Save no. of parallels on stack.
        curParLevelStack = append(curParLevelStack, curParLevel)
        // Reset no. of parallels.
        curParLevel = 0
        Log("(")
    }
    elem RBRACKET
    {
        // Sum elements:
        // Restore upper level no. of sums. 
        curSumLevel, curSumLevelStack = pop(curSumLevelStack)

        // Parallel elements:
        // Restore upper level no. of parallels. 
        curParLevel, curParLevelStack = pop(curParLevelStack)
        Log(")")
    }
