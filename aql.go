package aranGO

import(
   "errors"
)

// Aql query
type Query struct{
  // mandatory
  Aql       string                  `json:"query,omitempty"`
  //Optional values
  Batch     int                     `json:"batchSize,omitempty"`
  Count     bool                    `json:"count,omitempty"`
  BindVars  map[string]interface{}  `json:"bindVars,omitempty"`
  Options   map[string]interface{}  `json:"bindVars,omitempty"`

  // Control
  Validate  bool                    `json:"-"`
  ErrorMsg  string                  `json:"errorMessage,omitempty"`
}

func NewQuery(query string) *Query{
  if query == ""{
    return nil
  }else{
    var q Query
    q.Aql = query
    return &q
  }
}

func (q *Query) Modify(query string) error{
  if query == ""{
    return errors.New("query must not be empty")
  }else{
    q.Aql = query
    return nil
  }
}

// Validate query before execution
func (q *Query) Check(opt bool) {
  q.Validate = opt
  return
}
