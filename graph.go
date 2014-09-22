package aranGO

import (
	"errors"
)

// Graph structure
type Graph struct {
	Id       string           `json:"_id,omitempty"`
	Key      string           `json:"_key"`
	Name     string           `json:"name"`
	EdgesDef []EdgeDefinition `json:"edgeDefinitions"`
	Orphan   []string         `json:"orphanCollections"`
	db       *Database
}

type graphObj struct {
	V map[string]interface{} `json:"vertex"`
	E map[string]interface{} `json:"edge"`
}

// Tranvers graph
func (g *Graph) Traverse(t *Traversal, r interface{}) error {
	if g.Key == "" {
		return errors.New("Invalid graph to travers")
	}

	res, err := g.db.send("traversal", "", "POST", t, r, r)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 404:
		return errors.New("Invalid collections or graph")
	case 400:
		return errors.New("Traversal specification is either missing or malformed")
	case 500:
		return errors.New("Error inside traversal or traversal performed more than maxIterations iterations.")
	default:
		return nil
	}
}

// Remove Vertex
func (g *Graph) RemoveE(col string, key string) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	res, err := g.db.get("gharial", g.Name+"/edge/"+col+"/"+key, "DELETE", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

//Replace edge
func (g *Graph) ReplaceE(col string, key string, doc interface{}, patch interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}
	var gr graphObj

	res, err := g.db.send("gharial", g.Name+"/edge/"+col+"/"+key, "PUT", patch, &gr, &gr)
	subParse(gr.V, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or edge id")
	default:
		return nil
	}
}

//Patch Edge
func (g *Graph) PatchE(col string, key string, doc interface{}, patch interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	var gr graphObj
	res, err := g.db.send("gharial", g.Name+"/edge/"+col+"/"+key, "PATCH", patch, &gr, &gr)
	subParse(gr.E, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

// Get Edge
func (g *Graph) GetE(col string, key string, edge interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid collection or key value")
	}
	var gr graphObj
	res, err := g.db.get("gharial", g.Name+"/edge/"+col+"/"+key, "GET", nil, &gr, &gr)
	subParse(gr.E, edge)

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save edge")
	default:
		return nil
	}
}

//Creates a Egde
func (g *Graph) E(col string, edge interface{}) error {
	if col == "" {
		return errors.New("Invalid collection name")
	}
	var gr graphObj
	res, err := g.db.send("gharial", g.Name+"/edge/"+col, "POST", edge, &gr, &gr)
	subParse(gr.E, edge)

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save edge")
	default:
		return nil
	}
}

// Remove Vertex
func (g *Graph) RemoveV(col string, key string) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	res, err := g.db.get("gharial", g.Name+"/vertex/"+col+"/"+key, "DELETE", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

//Replace vertex
func (g *Graph) ReplaceV(col string, key string, doc interface{}, patch interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}
	var gr graphObj

	res, err := g.db.send("gharial", g.Name+"/vertex/"+col+"/"+key, "PUT", patch, &gr, &gr)
	subParse(gr.V, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

//Patch vertex
func (g *Graph) PatchV(col string, key string, doc interface{}, patch interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	var gr graphObj
	res, err := g.db.send("gharial", g.Name+"/vertex/"+col+"/"+key, "PATCH", patch, &gr, &gr)
	subParse(gr.V, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		return nil
		// need to add conditional update
	case 404, 412:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

//Creates a vertex in collection
func (g *Graph) V(col string, doc interface{}) error {
	if col == "" {
		return errors.New("Invalid collection name")
	}
	var gr graphObj
	res, err := g.db.send("gharial", g.Name+"/vertex/"+col, "POST", doc, &gr, &gr)
	subParse(gr.V, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save vertex")
	default:
		return nil
	}
}

//Gets Vertex from collection
func (g *Graph) GetV(col string, key string, doc interface{}) error {
	if key == "" || col == "" {
		return errors.New("Invalid key or collection to get")
	}
	// workaround nested vertex
	var gr graphObj
	res, err := g.db.get("gharial", g.Name+"/vertex/"+col+"/"+key, "GET", nil, &gr, &gr)
	subParse(gr.V, doc)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200:
		return nil
	case 404:
		return errors.New("Invalid collection or graph")
	case 412:
		return errors.New("Updated")
	default:
		return nil
	}

}

// Remove vertex collections
func (g *Graph) RemoveVertexCol(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}
	if col == "" {
		return errors.New("Invalid collection")
	}

	res, err := g.db.get("gharial", g.Name+"/vertex/"+col, "DELETE", nil, nil, nil)
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
func (g *Graph) RemoveEdgeDef(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}
	if col == "" {
		return errors.New("Invalid edge")
	}

	res, err := g.db.get("gharial", g.Name+"/edge/"+col, "DELETE", nil, nil, nil)
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
func (g *Graph) AddVertexCol(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	pay := map[string]string{"collection": col}
	res, err := g.db.send("gharial", g.Name+"/edge", "POST", pay, g, g)
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

func (g *Graph) AddEdgeDef(ed *EdgeDefinition) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	res, err := g.db.send("gharial", g.Name+"/edge", "POST", ed, g, g)
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

func (g *Graph) ReplaceEdgeDef(name string, ed *EdgeDefinition) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	res, err := g.db.send("gharial", g.Name+"/edge/"+name, "POST", ed, g, g)
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

func (g *Graph) ListEdgesDef() ([]string, error) {
	if g.db == nil {
		return []string{}, errors.New("Invalid db")
	}
	var gr graphResponse
	res, err := g.db.get("gharial", g.Name+"/edge", "GET", nil, &gr, &gr)
	if err != nil {
		return []string{}, err
	}

	switch res.Status() {
	case 200:
		return gr.Col, nil
	default:
		return []string{}, errors.New("Invalid graph")
	}
}

func (g *Graph) ListVertexCol() ([]string, error) {
	if g.db == nil {
		return []string{}, errors.New("Invalid db")
	}
	var gr graphResponse
	res, err := g.db.get("gharial", g.Name+"/vertex", "GET", nil, &gr, &gr)
	if err != nil {
		return []string{}, err
	}

	switch res.Status() {
	case 200:
		return gr.Col, nil
	default:
		return []string{}, errors.New("Invalid graph")
	}
}

func (g *Graph) AddEdgeDefinition(ed EdgeDefinition) error {
	if ed.Collection == "" {
		return errors.New("Invalid collection")
	}
	g.EdgesDef = append(g.EdgesDef, ed)
	return nil
}

type graphResponse struct {
	Error  bool     `json:"error"`
	Code   int      `json:"code"`
	Graph  Graph    `json:"graph"`
	Graphs []Graph  `json:"graphs"`
	Col    []string `json:"collections"`
}

type EdgeDefinition struct {
	Collection string   `json:"collection"`
	From       []string `json:"from"`
	To         []string `json:"to"`
}

func NewEdgeDefinition(col string, from []string, to []string) *EdgeDefinition {
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
func (db *Database) CreateGraph(name string, eds []EdgeDefinition) (*Graph, error) {
	var g Graph
	var gr graphResponse
	if name != "" {
		g.Name = name
		g.EdgesDef = eds
	} else {
		return nil, errors.New("Invalid graph name")
	}
	if eds == nil {
		return nil, errors.New("Invalid edges")
	}
	res, err := db.send("gharial", "", "POST", g, &gr, &gr)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 201:
		return &g, nil
	case 409:
		return nil, errors.New("Conflic creating graph")
	default:
		return nil, errors.New("Conflic creating graph")
	}
}

func (db *Database) DropGraph(name string) error {

	if name == "" {
		return errors.New("Invalid graph name")
	}
	res, err := db.get("gharial", name, "DELETE", nil, nil, nil)
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

func (db *Database) Graph(name string) *Graph {
	var g graphResponse
	if name == "" {
		return nil
	}
	_, err := db.get("gharial", name, "GET", nil, &g, &g)
	if err != nil {
		return nil
	}
	// set DB
	g.Graph.db = db
	return &g.Graph
}

func (db *Database) ListGraphs() ([]Graph, error) {
	var gr graphResponse

	res, err := db.get("gharial", "", "GET", nil, &gr, &gr)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 200:
		return gr.Graphs, nil
	default:
		return nil, errors.New("Unable to list graphs")
	}
}

type Traversal struct {
	graphName   string     `json:"graphName"`
	StartVertex string     `json:"startVertex"`
	Filter      string     `json:"filter"`
	MinDepth    int        `json:"minDepth"`
	MaxDepth    int        `json:"maxDepth"`
	Visitor     string     `json:"visitor"`
	Direction   string     `json:"direction"`
	Init        string     `json:"init"`
	Expander    string     `json:"expander"`
	Sort        string     `json:"sort"`
	Strategy    string     `json:"strategy"`
	Order       string     `json:"order"`
	ItemOrder   string     `json:"itemOrder"`
	Unique      Uniqueness `json:"uniqueness"`
	MaxIter     int        `json:"maxIterations"`
}

type Uniqueness struct {
	Edges    string `json:"edges"`
	Vertices string `json:"vertices"`
}
