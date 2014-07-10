package aranGO

import (
	"errors"
  "strconv"
)

// Aql query
type Query struct {
	// mandatory
	Aql string `json:"query,omitempty"`
	//Optional values
	Batch    int                    `json:"batchSize,omitempty"`
	Count    bool                   `json:"count,omitempty"`
	BindVars map[string]interface{} `json:"bindVars,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
  // opetions fullCount bool
  // Note that the fullCount sub-attribute will only be present in the result if the query has a LIMIT clause and the LIMIT clause is actually used in the query.

	// Control
	Validate bool   `json:"-"`
	ErrorMsg string `json:"errorMessage,omitempty"`
}

func NewQuery(query string) *Query {
  var q Query
  // alocate maps
  q.Options = make(map[string]interface{})
  q.BindVars= make(map[string]interface{})

	if query == "" {
		return &q
	} else {
		q.Aql = query
		return &q
	}
}

func (q *Query) Modify(query string) error {
	if query == "" {
		return errors.New("query must not be empty")
	} else {
		q.Aql = query
		return nil
	}
}

// Validate query before execution
func (q *Query) MustCheck() {
	q.Validate = true
	return
}

type AqlStructer interface{
  Generate()     string
}

// Basic structure
type AqlStruct struct {
  // main loop var
  main        string
  list        string
  lines       []AqlStructer
  vars        map[string]string
  //Collect
  Groups      Collect                   `json:"groups"`
  // Return
  // could be string or AqlStruct
  View        `json:"view"`
}

func (aq *AqlStruct) Generate() string{
  q:= "FOR "+aq.main+" IN "+aq.list

  for _,line :=range(aq.lines){
    q+= line.Generate()
  }

  // Add default view
  if aq.View == nil {
    q+= `
RETURN `+aq.main
  }else{
    // Generate view
    q+= aq.View.Generate()
  }

  return q
}

type View map[string]interface{}

func (v View) Generate() string{
  q:= ""
  return q
}

type Collects struct{
  // COLLECT key = Obj.Var,..., INTO Gro
  Collect map[string]Group  `json:"collect"`
  Gro   string              `json:"group"`
}

type Group struct{
  Obj   string  `json:"obj"`
  Var   string  `json:"var"`
}

type Limits struct{
  Skip   int64
  Limit  int64
}

func (l Limits) Generate() string {
  skip := strconv.FormatInt(l.Skip,10)
  limit:= strconv.FormatInt(l.Limit,10)
  li := `
LIMIT `+skip+`,`+limit
  return li
}

func (aq *AqlStruct) AddLimit(skip,limit int64) *AqlStruct{
  var l Limits
  l.Skip = skip
  l.Limit = limit
  aq.lines = append(aq.lines,l)
  return aq
}

type Lets struct {
  list    map[string]interface{}    `json:"lets"`
}

type Filters struct{
  Key    string  `json:"key"`
  Filter []Pair  `json:"filters"`
}

type Pair struct {
  Obj     string      `json:"obj"`
  Logic   string      `json:"log"`
  Value   interface{} `json:"val"`
}

func (f Filters) Generate() string{
  // check if filters available
  if len(f.Filter) == 0 {
    return ""
  }
  var oper      string

  lenmap := 0
  q := ""

  if f.Filter == nil{
    return ""
  }

  pairs := f.Filter
  key   := f.Key
  // iterate over filters
  // first
  q += `
FILTER (`
  oper = "||"

  for i,pair := range(pairs){
    if i == len(pairs) -1 {
      oper = ""
    }
    switch pair.Value.(type) {
      case bool:
        q += key+"."+pair.Obj+" "+pair.Logic+" "+strconv.FormatBool(pair.Value.(bool))+" "+oper+" "
      case int:
        q += key+"."+pair.Obj+" "+pair.Logic+" "+strconv.Itoa(pair.Value.(int))+" "+oper+" "
      case int64:
        q += key+"."+pair.Obj+" "+pair.Logic+" "+strconv.FormatInt(pair.Value.(int64),10)+" "+oper+" "
      case string:
        q += key+"."+pair.Obj+" "+pair.Logic+" '"+pair.Value.(string)+"' "+oper+" "
      case float32,float64:
        q += key+"."+pair.Obj+" "+pair.Logic+" "+strconv.FormatFloat(pair.Value.(float64),'f',6,64)+" "+oper+" "
      case Var:
        q += key+"."+pair.Obj+" "+pair.Logic+" "+pair.Value.(Var).Obj+"."+pair.Value.(Var).Name+" "+oper+" "
    }
    if i == len(pairs)-1{
      q += ")"
    }
  }
  // next key
  lenmap++
  return q
}

// Variable into document
type Var  struct {
  Obj     string      `json:"obj"`
  Name    string      `json:"name"`
}


//If collect set, vars are reset to context vars
type Collect struct {
  Groups   map[string]string   `json:"groups"`
  Into     string              `json:"into"`
}

/*
func (aq *AqlStruct) ToString() string{
  q := ""
  aq.vars = make(map[string]string)
  // Start with list
  if len(aq.List) == 1 {
    for key,list := range(aq.List){
      aq.vars[key] = ""
      q += "FOR "+key+" IN "+list+" "
    }
  }
  // add filters
  if len(aq.Filters) > 0{
    q += aq.filters()+`
    `
  }
  //
  // view
  q += aq.view()
  return q
}
*/

func (f AqlStruct) view() string{
  if len(f.View) == 0{
    return "RETURN "+f.main
  }
  return ""
}


func (aq *AqlStruct) SetList(obj string,list string) *AqlStruct{
  aq.main = obj
  aq.list = list
  return aq
}

func (aq *AqlStruct) AddFilter(key string,values []Pair) *AqlStruct{
  var fil Filters
  if key != "" && values != nil{
    fil.Key = key
    fil.Filter = values
    aq.lines = append(aq.lines,fil)
  }
  return aq
}


