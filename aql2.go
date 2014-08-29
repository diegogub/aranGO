package aranGO

import (
    "errors"
    "strconv"
)

type AqlStructer interface{
  Generate() string
}

// Basic Aql struct to build Aql Query
type AqlStruct struct {
    lines []AqlStructer
    // number of loops and vars 
    nlopp uint
    vars  []string
}

func NewAqlStruct() *AqlStruct{
    var aq AqlStruct
    return &aq
}

// Returns sub struct with same var context
func (a *AqlStruct) subStruct() (*AqlStruct){
    var substruct AqlStruct
    if len(a.vars) > 0 {
        for v in a.vars{
            substruct.vars = append(substruct.vars,v)
        }
        return substruct
    }else{
        // fatal error
        panic("getting substruct from empty struct")
    }
}
