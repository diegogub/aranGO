package aranGO

import(
  "errors"
)

// Graph structure
type Graph struct {
  Id        string              `json:"_id,omitempty"`
  Key       string              `json:"_key"`
  Name      string              `json:"name"`
  EdgesDef  []EdgeDefinition    `json:"edgeDefinitions"`
  Orphan    []string            `json:"orphanCollections"`
  db        *Database
}

// Remove vertex collections
func(g *Graph) RemoveVertex(col string) error{
  if g.db == nil{
    return errors.New("Invalid db")
  }
  if col == "" {
    return errors.New("Invalid collection")
  }

  res, err :=g.db.get("gharial",g.Key+"/vertex/"+col,"DELETE",nil,nil,nil)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 400:
      return errors.New("Cannot remove collection")
    default:
      return errors.New("Invalid graph")
  }
}

// Remove edge 
func(g *Graph) RemoveEdge(col string) error{
  if g.db == nil{
    return errors.New("Invalid db")
  }
  if col == "" {
    return errors.New("Invalid edge")
  }

  res, err :=g.db.get("gharial",g.Key+"/edge/"+col,"DELETE",nil,nil,nil)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 400:
      return errors.New("Cannot remove edge")
    default:
      return errors.New("Invalid graph")
  }
}

// Add vertex collections
func (g *Graph) AddVertex(col string) error{
  if g.db == nil{
    return errors.New("Invalid db")
  }

  pay := map[string]string { "collection" : col}
  res, err := g.db.send("gharial",g.Name+"/edge","POST",pay,g,g)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 400:
      return errors.New("Unable to add vertex collection")
    default:
      return errors.New("Invalid graph")
  }
}


func (g *Graph) AddEdge(ed *EdgeDefinition) error{
  if g.db == nil{
    return errors.New("Invalid db")
  }

  res, err := g.db.send("gharial",g.Name+"/edge","POST",ed,g,g)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 400:
      return errors.New("Unable to add edge definition")
    default:
      return errors.New("Invalid graph")
  }
}

func (g *Graph) ReplaceEdge(name string,ed *EdgeDefinition) error{
  if g.db == nil{
    return errors.New("Invalid db")
  }

  res, err := g.db.send("gharial",g.Name+"/edge/"+name,"POST",ed,g,g)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 400:
      return errors.New("Unable to replace edge definition")
    default:
      return errors.New("Invalid graph")
  }
}

func (g *Graph) ListEdges() ([]string,error){
  if g.db == nil{
    return []string{},errors.New("Invalid db")
  }
  var gr graphResponse
  res, err := g.db.get("gharial",g.Name+"/edge","GET",nil,&gr,&gr)
  if err != nil {
    return []string{},err
  }

  switch res.Status(){
    case 200:
      return gr.Col,nil
    default:
      return []string{},errors.New("Invalid graph")
  }
}

func (g *Graph) ListVertex() ([]string,error){
  if g.db == nil{
    return []string{},errors.New("Invalid db")
  }
  var gr graphResponse
  res, err := g.db.get("gharial",g.Name+"/vertex","GET",nil,&gr,&gr)
  if err != nil {
    return []string{},err
  }

  switch res.Status(){
    case 200:
      return gr.Col,nil
    default:
      return []string{},errors.New("Invalid graph")
  }
}

func (g *Graph) AddEdgeDefinition(ed EdgeDefinition) error {
  if ed.Collection == "" {
    return errors.New("Invalid collection")
  }
  g.EdgesDef = append(g.EdgesDef,ed)
  return nil
}

type graphResponse struct {
  Error   bool    `json:"error"`
  Code    int     `json:"code"`
  Graph   Graph   `json:"graph"`
  Graphs  []Graph `json:"graphs"`
  Col     []string `json:"collections"`
}

type EdgeDefinition struct {
  Collection  string    `json:"collection"`
  From        []string  `json:"from"`
  To          []string  `json:"to"`
}

func NewEdgeDefinition(col string,from []string,to []string ) *EdgeDefinition{
  var e EdgeDefinition
  if col == "" {
    return nil
  }
  e.Collection = col
  e.From = from
  e.To = to
  return &e
}

// Creates graphs
func (db *Database) CreateGraph(name string,eds []EdgeDefinition) (*Graph,error) {
  var g Graph
  var gr graphResponse
  if name != "" {
    g.Name = name
    g.EdgesDef = eds
  }else{
    return nil,errors.New("Invalid graph name")
  }
  if eds == nil {
    return nil,errors.New("Invalid edges")
  }
  res,err := db.send("gharial","","POST",g,&gr,&gr)
  if err != nil {
    return nil,err
  }

  switch res.Status() {
    case 201:
      return &g,nil
    case 409:
      return nil,errors.New("Conflic creating graph")
    default:
      return nil,errors.New("Conflic creating graph")
  }
}

func (db *Database) DropGraph(name string) error{

  if name == ""{
    return errors.New("Invalid graph name")
  }
  res, err := db.get("gharial",name,"DELETE",nil,nil,nil)
  if err != nil {
    return err
  }

  switch res.Status() {
    case 200:
      return nil
    case 404:
      return errors.New("Graph not found")
    default:
      return nil
  }

}

func (db *Database) Graph(name string) (*Graph){
  var g graphResponse
  if name == ""{
    return nil
  }
  _, err := db.get("gharial",name,"GET",nil,&g,&g)
  if err != nil {
    return nil
  }
  // set DB
  g.Graph.db = db
  return &g.Graph
}

func (db *Database) ListGraphs() ([]Graph,error){
  var gr graphResponse

  res, err := db.get("gharial","","GET",nil,&gr,&gr)
  if err != nil {
    return nil,err
  }

  switch res.Status(){
    case 200:
      return gr.Graphs,nil
    default:
      return nil,errors.New("Unable to list graphs")
  }
}
