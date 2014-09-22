package aranGO

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// AqlObject
type Obj map[string]interface{}

// AqlList
type List []interface{}

func (l List) String() string {
	code := "["
	var aux []string
	for _, i := range l {
		var pair string
		switch i.(type) {
		case string:
			pair += "'" + i.(string) + "'"
		case int:
			pair += strconv.Itoa(i.(int))
		case int32:
			pair += strconv.FormatInt(i.(int64), 10)
		case int64:
			pair += strconv.FormatInt(i.(int64), 10)
		case float64:
			pair += strconv.FormatFloat(i.(float64), 'f', 6, 64)
		default:
			pair += " "
		}
		aux = append(aux, pair)
	}
	code += strings.Join(aux, ",") + "]"
	return code
}

func (ob Obj) String() string {
	code := "{ "
	var aux []string
	for key, val := range ob {
		var pair string
		pair += key + " : "
		switch val.(type) {
		case Var:
			pair += val.(Var).String()
		case string:
			pair += "'" + val.(string) + "'"
		case int32:
			pair += strconv.FormatInt(val.(int64), 10)
		case int64:
			pair += strconv.FormatInt(val.(int64), 10)
		case AqlStruct, *AqlStruct:
			pair += "( " + val.(*AqlStruct).Generate() + " )"
		case AqlFunction:
			pair += val.(*AqlFunction).Generate()
		default:
			pair += genValue(val)
		}
		aux = append(aux, pair)
	}
	code += strings.Join(aux, ",") + " }"
	// just code it to json?
	return code
}

// AQL Variable
type Var struct {
	Obj  string
	Name string
}

// Represent AQL collection
func Col(name string) Var {
	var v Var
	v.Obj = name
	return v
}

func Atr(obj, name string) Var {
	var v Var
	v.Obj = obj
	v.Name = name
	return v
}

func (v Var) String() string {
	if v.Obj != "" && v.Name == "" {
		return v.Obj
	}

	if v.Obj != "" && v.Name != "" {
		q := v.Obj + "." + v.Name
		return q
	}

	return ""
}

// Aql query
type Query struct {
	// mandatory
	Aql string `json:"query,omitempty"`
	//Optional values Batch    int                    `json:"batchSize,omitempty"`
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
	q.BindVars = make(map[string]interface{})

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

type AqlStructer interface {
	Generate() string
}

// Basic Aql struct to build Aql Query
type AqlStruct struct {
	lines []AqlStructer
	// number of loops and vars
	nlopp uint
	vars  []string
	err   bool
}

//Generate Aql query string
func (aq *AqlStruct) Generate() string {
	if len(aq.lines) == 0 {
		return ""
	}

	var code string
	for _, aql := range aq.lines {
		code += `
            ` + aql.Generate()
	}

	return code
}

func NewAqlStruct() *AqlStruct {
	var aq AqlStruct
	return &aq
}

//Execute AqlStuct into database
func (aq *AqlStruct) Execute(db *Database) (*Cursor, error) {
	q := NewQuery(aq.Generate())
	c, err := db.Execute(q)
	return c, err
}

// Returns sub struct with same var context
func (aq *AqlStruct) subStruct() *AqlStruct {
	var substruct AqlStruct
	if len(aq.vars) > 0 {
		for _, v := range aq.vars {
			substruct.vars = append(substruct.vars, v)
		}
		return &substruct
	} else {
		// fatal error
		panic("getting substruct from empty struct")
	}
}

// FOR var IN [] //
type aqlFor struct {
	in interface{}
	v  string
}

func (aqf aqlFor) Generate() string {
	code := ""
	if aqf.v == "" {
		return code
	}

	code += "FOR " + aqf.v + " IN "

	switch aqf.in.(type) {
	case string:
		code += aqf.in.(string)
	case *AqlFunction:
		code += aqf.in.(AqlFunction).Generate()
	case []string:
		code += "[" + strings.Join(aqf.in.([]string), ", ") + "]"
	case List:
		code += aqf.in.(List).String()
	default:
		return code
	}

	return code
}

func (aq *AqlStruct) For(v string, in interface{}) *AqlStruct {
	var afor aqlFor
	afor.v = v
	afor.in = in
	aq.lines = append(aq.lines, afor)
	return aq
}

// Aql Return
type aqlReturn struct {
	Atr Var
	Var string
	Ret Obj
}

func (ar aqlReturn) Generate() string {
	code := "RETURN "
	if ar.Var != "" {
		code += ar.Var
	} else {
		if len(ar.Ret) > 0 {
			code += ar.Ret.String()
		} else {
			code += ar.Atr.String()
		}
	}
	return code
}

func (aq *AqlStruct) Return(view interface{}) *AqlStruct {
	var ret aqlReturn
	switch view.(type) {
	case string:
		if view.(string) == "" {
			return aq
		} else {
			ret.Var = view.(string)
		}
	case Obj:
		ret.Ret = view.(Obj)
	case Var:
		ret.Atr = view.(Var)
	default:
		return aq
	}
	aq.lines = append(aq.lines, ret)
	return aq
}

//Aql filter add Filter()  to AqlQuery
// Could be use like:
//        - Filter(custom ... string)
//            example:
//                Filter("u.name == 'Diego' && u.age > 20")
//                out:
//                    FILTER u.name == 'Diego' && u.age > 21
//        - Filter(key string,fil ... Filter || AqlFunction, any bool)
//            example:
//                Filter("u",Fil("sum","eq",213),Fil("age","gt",21),true)
//                out:
//                    FILTER u.sum == 213 || u.age > 21
//                Filter("u",Fil("sum","eq",213),FilField("id","==","adm.id"),false)
//                out:
//                    FILTER u.sum == 213
//                    FILTER u.id == adm.id
//                Filter("u",Fil("sum","eq",213),FilField("id","==","adm.id"),false)
//                out:
//                    FILTER u.sum == 213
//                    FILTER u.id == adm.id
//                Filter("u",Fun("LIKE",Atr("u","name"),"di%",true)
//                out:
//                    FILTER u.age > 21  ||  LIKE(u.name,'di%',true)
//        -FILTER(jsonFilter string)
//          example:
//            Filter(`{ "key" : "u" , "filters": [{ "name": "name", "like": "gt", "val": "die%" },{ "name": "status", "op": "eq", "val":"A"}] }`)
//            out:
//                FILTER (IS_NULL(u.name) == false && LIKE(u.name,'die%',true)) && u.status == 'A'
//            Filter(`{ "key" : "u" , "filters": [{ "name": "status", "op": "==", "val": "P" },{ "name": "status", "op": "eq", "val":"A"}], "any" : true }`)
//            out:
//                FILTER u.status == 'P' || u.status == 'A'
//            Filter(`{ "key" : "u" , "filters": [{ "name": "id", "op": "==", "field": "adm.id" },{ "name": "status", "op": "eq", "val":"A"}], "any" : true }`)
//            out:
//                Filter(`{ "key" : "u" , "filters": [{ "name": "id", "op": "==", "field": "adm.id" },{ "name": "status", "op": "eq", "val":"A"}], "any" : true }`)
func (aq *AqlStruct) Filter(f ...interface{}) *AqlStruct {
	// no parameters
	if len(f) == 0 {
		return aq
	}
	// check last parameter
	var any bool
	var fil2 AqlFilter
	// filter key
	var key string
	// number of arguments to skip
	var jump int
	switch f[len(f)-1].(type) {
	case bool:
		any = f[len(f)-1].(bool)
		// check first place for default key
		switch f[0].(type) {
		case string:
			fil2.DefaultKey = f[0].(string)
			key = fil2.DefaultKey
			jump++
		}
	default:
		any = false
	}
	// concurrent filter
	fil2.Any = any

	for _, i := range f {
		// jump few arguments
		if jump > 0 {
			jump--
			continue
		}
		var fil AqlFilter

		switch i.(type) {
		// check if []byte or string if it's valid JSON to create Filter
		case []byte:
			s := string(i.([]byte))
			if isJSON(s) {
				fil = FilterJSON(s)
			} else {
				fil.Custom = s
			}
		case string:
			if isJSON(i.(string)) {
				fil = FilterJSON(i.(string))
			} else {
				// how can i validate it's normal text?
				fil.Custom = i.(string)
			}
		case AqlFilter:
			fil = i.(AqlFilter)
		case AqlFunction:
			if any {
				fil2.Functions = append(fil2.Functions, i.(AqlFunction))
				continue
			} else {
				fil.DefaultKey = key
				fil.Functions = append(fil.Functions, i.(AqlFunction))
			}
		case Filter:
			if any {
				fil2.Filters = append(fil2.Filters, i.(Filter))
				continue
			} else {
				fil.DefaultKey = key
				fil.Filters = append(fil.Filters, i.(Filter))
			}
		default:
			continue
		}
		aq.lines = append(aq.lines, fil)
	}
	if len(fil2.Filters) > 0 || len(fil2.Functions) > 0 {
		aq.lines = append(aq.lines, fil2)
	}
	return aq
}

type AqlFilter struct {
	DefaultKey string `json:"key"`
	// never include in json parsing
	Custom string `json:"-"`
	// Function filters
	Functions []AqlFunction `json:"functions"`
	// Filters
	Filters []Filter `json:"filters"`
	// Match all the filters or any of them
	Any bool `json:"any"`
}

func (aqf AqlFilter) Generate() string {
	code := "FILTER "
	if aqf.Custom != "" {
		code += aqf.Custom
		return code
	}
	var fun, fil []string
	var sfun, sfil string
	var aux string

	logic := " && "
	if aqf.Any {
		logic = " || "
	}

	for _, f := range aqf.Functions {
		aux = f.Generate()
		if aux != "" {
			fun = append(fun, aux)
		}
	}

	for _, f := range aqf.Filters {
		aux = f.String(aqf.DefaultKey)
		if aux != "" {
			fil = append(fil, aux)
		}
	}

	if len(fil) > 1 {
		sfil = strings.Join(fil, logic)
	} else {
		if len(fil) == 1 {
			sfil = fil[0]
		}
	}

	if len(fun) > 1 {
		sfun = strings.Join(fun, logic)
	} else {
		if len(fun) == 1 {
			sfun = fun[0]
		}
	}

	if sfil != "" {
		code += sfil
	}

	if sfun != "" {
		code += " " + logic + " " + sfun
	}

	return code
}

// Returns AqlFilter parsing valid json string
func FilterJSON(s string) AqlFilter {
	var aqf AqlFilter
	json.Unmarshal([]byte(s), &aqf)
	return aqf
}

type Filter struct {
	AtrR string `json:"name"`
	// compare to value or function depending on operand
	Value interface{} `json:"val,omitempty"`
	// compare to field, need to check if it's valid variable!
	Field string `json:"field,omitempty"`
	// could be AqlFunction too
	Function *AqlFunction `json:"-"`
	// Operator
	Oper string `json:"op"`
	// All valid operations
	/*
	   ==, eq, equals, equals_to
	   !=, neq, does_not_equal, not_equal_to
	   >, gt
	   <, lt
	   >=, ge, gte, geq
	   <=, le, lte, leq
	   like

	   To implement
	   is_null
	   is_not_null
	   in, not_in
	   has
	*/
}

// Returns filter , comparing with value
func Fil(atr string, oper string, i interface{}) Filter {
	var f Filter
	f.AtrR = atr
	f.Oper = oper
	f.Value = i
	return f
}

// Returns filter , comparing 2 fields
func FilField(atr string, oper string, i string) Filter {
	var f Filter
	f.AtrR = atr
	f.Oper = oper
	f.Field = i
	return f
}

func (f Filter) String(key string) string {
	var code string
	if key == "" || f.Oper == "" {
		return ""
	}

	switch f.Oper {
	case "==", "eq", "equals", "equals_to":
		f.Oper = "=="
	case "!=", "neq", "does_not_equal", "not_equal_to":
		f.Oper = "!="
	case ">", "gt":
		f.Oper = ">"
	case "<", "lt":
		f.Oper = "<"
	case ">=", "ge", "gte":
		f.Oper = ">="
	case "<=", "le", "lte":
		f.Oper = "<="
	case "like":
		// auxiliar function,
		NotNull := "(IS_NULL(" + Atr(key, f.AtrR).String() + ") == false && "
		fun := Fun("LIKE", Atr(key, f.AtrR), f.Value, true)
		return NotNull + fun.Generate() + ")"
		// operator as "like" should create corresponding AqlFunction
	}

	if f.Function != nil {
		return f.Function.Generate()
	}

	code += key + "." + f.AtrR + " " + f.Oper + " "
	if f.Field != "" {
		aux := strings.Split(f.Field, ".")
		if len(aux) == 2 {
			code += f.Field
		} else {
			return ""
		}
	} else {
		code += genValue(f.Value)
	}

	return code
}

//Aql Sort
// Usage:
//  Sort("u.name","u.age","ASC","u.created","DESC")
// or
//  Sort(Atr("u","name),Atr("u","age"),"ASC",Var("u","created"),"DESC")
func (aq *AqlStruct) Sort(i ...interface{}) *AqlStruct {
	var sort AqlSort
	for _, p := range i {
		var so Sort
		switch p.(type) {
		case string:
			s := p.(string)
			if s == "DESC" || s == "ASC" {
				so.Direction = s
			} else {
				so.Variable = s
			}
		case Var:
			so.Variable = p.(Var)
		}
		sort.List = append(sort.List, so)
	}
	if len(sort.List) > 0 {
		aq.lines = append(aq.lines, sort)
	}

	return aq
}

type AqlSort struct {
	List []Sort `json:"sort"`
}

func (aqs AqlSort) Generate() string {
	var list []string
	var code, aux string
	i := 0
	for _, s := range aqs.List {
		aux = s.String()
		if aux != "" {
			if aux != "DESC" && aux != "ASC" {
				list = append(list, aux)
				i++
			} else {
				list[i-1] += " " + aux
			}
		}
	}
	code = "SORT " + strings.Join(list, ",")

	return code
}

type Sort struct {
	Variable  interface{} `json:"field"`
	Direction string      `json:"direction,omitempty"`
}

func (s Sort) String() string {
	var code string
	switch s.Variable.(type) {
	case string:
		aux := strings.Split(s.Variable.(string), ".")
		if len(aux) == 2 {
			code += s.Variable.(string)
		}
	case Var:
		code += s.Variable.(Var).String()
	case nil:
		return s.Direction
	}

	return code
}

type AqlLimit struct {
	Skip  int64 `json:"skip"`
	Limit int64 `json:"limit"`
	limit bool
}

func (l AqlLimit) Generate() string {
	var li string
	if !l.limit {
		skip := strconv.FormatInt(l.Skip, 10)
		limit := strconv.FormatInt(l.Limit, 10)
		li = `LIMIT ` + skip + `,` + limit
	} else {
		limit := strconv.FormatInt(l.Limit, 10)
		li = `LIMIT ` + limit
	}

	return li
}

// Aql Limit
// Usage:
//  Limit(10)
//  out: LIMIT 10
//  Limit(5,15)
//  out: LIMIT 5,15
func (aq *AqlStruct) Limit(s ...int64) *AqlStruct {
	var l AqlLimit
	if len(s) > 1 {
		l.Skip = s[0]
		l.Limit = s[1]
	}

	if len(s) == 1 {
		l.Limit = s[0]
		l.limit = true
	}

	if len(s) == 0 {
		return aq
	}

	aq.lines = append(aq.lines, l)
	return aq
}

// Aql Collect
// Usage:
// Collect("first = u.firstName, age = u.age INTO g")
func (aq *AqlStruct) Collect(sentence string) *AqlStruct {
	var col AqlCollect
	if sentence == "" {
		return aq
	}

	col.Sentence = sentence
	aq.lines = append(aq.lines, col)
	return aq
}

type AqlCollect struct {
	Sentence string `json:"collect"`
}

func (aqc AqlCollect) Generate() string {
	return "COLLECT " + aqc.Sentence
}

//Aql Let
func (aq *AqlStruct) Let(v string, i interface{}) *AqlStruct {
	var let AqlLet

	if v == "" {
		return aq
	}
	let.Var = v
	let.Exp = i
	aq.lines = append(aq.lines, let)
	return aq

}

type AqlLet struct {
	Var string
	Exp interface{}
}

func (aql AqlLet) Generate() string {
	if aql.Var == "" {
		return ""
	}

	code := "LET " + aql.Var + " = "
	switch aql.Exp.(type) {
	case AqlFunction:
		code += aql.Exp.(AqlFunction).Generate()
	case *AqlStruct:
		code += "( " + aql.Exp.(*AqlStruct).Generate() + " )"
	default:
		code += genValue(aql.Exp)
	}

	return code
}

//Aql INSERT
// Usage:
//   Insert(Obj{ "name" : Atr("u","name") , "test" : 2},"backup")
func (aq *AqlStruct) Insert(obj Obj, col string) *AqlStruct {
	if col == "" {
		return aq
	}
	var aqi AqlInsert
	aqi.Data = obj
	aqi.Col = col
	aq.lines = append(aq.lines, aqi)
	return aq
}

type AqlInsert struct {
	Data Obj
	Col  string
}

func (aqi AqlInsert) Generate() string {
	if aqi.Col == "" {
		return ""
	}

	code := "INSERT " + aqi.Data.String() + " IN " + aqi.Col

	return code
}

//Aql UPDATE
// Usage:
//
func (aq *AqlStruct) Update(doc interface{}, with Obj, col string, options interface{}) *AqlStruct {
	var aqu AqlUpdate
	if col == "" {
		return aq
	}

	aqu.Doc = doc
	aqu.Col = col
	aqu.With = with
	aqu.Options = options
	aq.lines = append(aq.lines, aqu)
	return aq
}

type AqlUpdate struct {
	Doc     interface{}
	Col     string
	With    Obj
	Options interface{}
}

func (aqu AqlUpdate) Generate() string {
	if aqu.Col == "" {
		return ""
	}
	var code string

	code = "UPDATE "

	switch aqu.Doc.(type) {
	case Obj:
		code += aqu.Doc.(Obj).String()
	case string:
		code += aqu.Doc.(string)
	case Var:
		code += aqu.Doc.(Var).String()
	default:
		return ""
	}

	code += " WITH " + aqu.With.String() + " IN " + aqu.Col

	switch aqu.Options.(type) {
	case Obj:
		code += " OPTIONS " + aqu.Options.(Obj).String()
	default:
	}

	return code
}

//Aql Replace
//Usage:
//  Replace("u._id",Obj{ "name" : "Diego" },"users",nil)
func (aq *AqlStruct) Replace(id interface{}, replace Obj, col string, options interface{}) *AqlStruct {
	if col == "" {
		return aq
	}
	var aqr AqlReplace
	aqr.Id = id
	aqr.Repl = replace
	aqr.Col = col
	aqr.Options = options
	aq.lines = append(aq.lines, aqr)
	return aq
}

type AqlReplace struct {
	Id      interface{}
	Repl    Obj
	Col     string
	Options interface{}
}

func (aqr AqlReplace) Generate() string {
	var code string

	code = "REPLACE "
	switch aqr.Id.(type) {
	case string:
		code += aqr.Id.(string)
	case Var:
		code += aqr.Id.(Var).String()
	}

	code += " WITH " + aqr.Repl.String() + " IN " + aqr.Col

	switch aqr.Options.(type) {
	case Obj:
		code += " OPTIONS " + aqr.Options.(Obj).String()
	default:
	}

	return code
}

//Aql Remove
//Usage:
// Remove("u._id","users",nil)
func (aq *AqlStruct) Remove(id interface{}, col string, options interface{}) *AqlStruct {
	if col == "" {
		return aq
	}
	var aqr AqlRemove

	aqr.Id = id
	aqr.Col = col
	aqr.Options = options

	aq.lines = append(aq.lines, aqr)

	return aq
}

type AqlRemove struct {
	Id      interface{}
	Col     string
	Options interface{}
}

func (aqr AqlRemove) Generate() string {
	if aqr.Col == "" {
		return ""
	}
	var code string
	code = "REMOVE "
	switch aqr.Id.(type) {
	case string:
		code += aqr.Id.(string)
	case Var:
		code += aqr.Id.(Var).String()
	default:
		return ""
	}

	code += " IN " + aqr.Col
	switch aqr.Options.(type) {
	case Obj:
		code += " OPTIONS " + aqr.Options.(Obj).String()
	default:
	}

	return code
}

// Aql functions
type AqlFunction struct {
	Name   string
	Params []interface{}
}

//Creates a AqlFunction to use into AqlStruct
func Fun(name string, i ...interface{}) AqlFunction {
	var f AqlFunction
	f.Name = name
	f.Params = i
	return f
}

func (f AqlFunction) Generate() string {
	code := f.Name
	if len(f.Params) == 0 {
		code += "()"
		return code
	}

	var ParamList []string
	var aux string

	for _, param := range f.Params {
		switch param.(type) {
		case AqlFunction:
			aux = param.(AqlFunction).Generate()
		case bool:
			aux = strconv.FormatBool(param.(bool))
		case Var:
			aux = param.(Var).String()
		case int:
			aux = strconv.Itoa(param.(int))
		case int32:
			aux = strconv.FormatInt(param.(int64), 10)
		case int64:
			aux = strconv.FormatInt(param.(int64), 10)
		case float64:
			aux = strconv.FormatFloat(param.(float64), 'f', 6, 64)
		case string:
			aux = "'" + param.(string) + "'"
		case *AqlFunction:
			aux = param.(AqlFunction).Generate()
		case []string:
			aux = "[" + strings.Join(param.([]string), ", ") + "]"
		case List:
			aux = param.(List).String()
		case nil:
			aux = ""
		default:
			aux = ""
		}
		if aux != "" {
			ParamList = append(ParamList, aux)
			aux = ""
		}
	}

	if len(ParamList) > 0 {
		code += "(" + strings.Join(ParamList, ",") + ")"
	}

	return code
}

func genValue(v interface{}) string {
	var q string
	switch v.(type) {
	case bool:
		q = strconv.FormatBool(v.(bool))
	case int:
		q = strconv.Itoa(v.(int))
	case int64:
		q = strconv.FormatInt(v.(int64), 10)
	case string:
		q = "'" + v.(string) + "'"
	case float32, float64:
		q = strconv.FormatFloat(v.(float64), 'f', 6, 64)
	case Var:
		q = v.(Var).Obj + "." + v.(Var).Name
	case []string:
		q = "[" + strings.Join(v.([]string), ", ") + "]"
	case List:
		q = v.(List).String()
	case nil:
		q = "null"
	}
	return q
}
