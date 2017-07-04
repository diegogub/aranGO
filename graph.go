package aranGO

import (
	"errors"
	"fmt"
)

	type EdgeDefinition struct {
	Collection string   		`json:"collection"`
	From       []string 		`json:"from"`
	To         []string 		`json:"to"`
}

func NewEdgeDefinition(collection string, from []string, to []string) *EdgeDefinition {
	var edge EdgeDefinition

	edge.Collection = collection
	edge.From = from
	edge.To = to
	return &edge
}

type VertexDefinition struct {
	Collection string   		`json:"collection"`
}

type GraphOptions struct {
	Name     string           	`json:"name"`
	Smart    bool           	`json:"isSmart,omitempty"`
	Options  *GraphExtOptions	`json:"options,omitempty"`
	EdgesDef []EdgeDefinition 	`json:"edgeDefinitions"`
	Orphan   []string         	`json:"orphanCollections,omitempty"`
}

type GraphExtOptions struct {
	Shards				int 	`json:"numberOfShards,omitempty"`
	SmartGraphAttr 		string 	`json:"smartGraphAttribute,omitempty"`
}

func NewGraphOptions(name string, edges []EdgeDefinition) *GraphOptions {
	var gopt GraphOptions

	gopt.Name = name
	gopt.Smart = false
	gopt.EdgesDef = edges
	return &gopt
}

func (opt *GraphOptions) IsSmart() {
	opt.Smart = true
	return
}

func (opt *GraphOptions) IsNotSmart() {
	opt.Smart = false
	return
}

// Sets the attribute name that is used to smartly shard the vertices of a graph
func (opt *GraphOptions) SmartGraphAttribute(attribute string) {

	if opt.Options == nil {
		var extOpt GraphExtOptions
		opt.Options = &extOpt
	}

	opt.Options.SmartGraphAttr = attribute
	return
}

// Sets the number of shards for a collection
func (opt *GraphOptions) Shard(num int) {
	if opt.Options == nil {
		var extOpt GraphExtOptions
		opt.Options = &extOpt
	}

	if num <= 0 {
		num = 1
	}

	opt.Options.Shards = num
	return
}

// Set graph orphan (array of additional vertex collections)
func (opt *GraphOptions) setOrphan(orphan []string) {
	opt.Orphan = orphan
	return
}

// Add graph edge to current array of edges
func (opt *GraphOptions) addEdgeDefinition (edge EdgeDefinition) error {
	if edge.Collection == "" {
		return errors.New("Invalid collection")
	}
	opt.EdgesDef = append(opt.EdgesDef, edge)
	return nil
}

// Set graph edges definition (an array of definitions for the edge)
func (opt *GraphOptions) setEdgesDefinition (edges []EdgeDefinition) {
	opt.EdgesDef = edges
	return
}

// Graph structure
type Graph struct {
	Id       		string			 `json:"_id,omitempty"`
	Rev      		string			 `json:"_rev,omitempty"`
	Key      		string           `json:"_key,omitempty"`

	Name    		string           `json:"name"`
	EdgesDef 		[]EdgeDefinition `json:"edgeDefinitions"`
	Orphan   		[]string         `json:"orphanCollections"`
	Smart	 		bool			 `json:"isSmart"`
	Shards	 		int				 `json:"numberOfShards"`
	SmartGraphAttrs	string			 `json:"smartGraphAttribute"`

	db       		*Database
}

type graphResponse struct {
	Error  		bool     `json:"error"`
	Code   		int      `json:"code"`
	Graph  		Graph    `json:"graph"`
	Graphs 		[]Graph  `json:"graphs"`
	Collections	[]string `json:"collections"`
}

// Get graphs list
// Wrapper: GET /_api/gharial
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

// Creates graph
// Wrapper: POST /_api/gharial
func (db *Database) CreateGraph(gopt *GraphOptions) error {
	if gopt.Name == "" {
		return errors.New("Invalid graph name")
	}
	if gopt.EdgesDef == nil {
		return errors.New("Invalid edges")
	}
	res, err := db.send("gharial", "", "POST", gopt, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		return nil
	case 409:
		return errors.New("Conflict creating graph")
	default:
		fmt.Printf("Status: %d\n", res.Status())
		fmt.Printf("Message: %s\n", res.RawText())
		return errors.New("Conflict creating graph")
	}
}

// Drop a graph
// Wrapper: DELETE /_api/gharial/{graph-name}
func (db *Database) DropGraph(name string) error {

	if name == "" {
		return errors.New("Invalid graph name")
	}
	res, err := db.get("gharial", name, "DELETE", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 201, 202:
		return nil
	case 404:
		return errors.New("Graph not found")
	default:
		return nil
	}
}

// Get a graph
// Wrapper: GET /_api/gharial/{graph-name}
func (db *Database) Graph(name string) (*Graph, error) {
	var gr graphResponse
	if name == "" {
		return nil, errors.New("Invalid graph name")
	}
	res, err := db.get("gharial", name, "GET", nil, &gr, &gr)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 200:
		gr.Graph.db = db
		return &gr.Graph, nil
	case 404:
		return nil, errors.New("Graph not found")
	default:
		return nil, errors.New("Invalid graph")
	}
}

// List edge definition
// Wrapper: GET /_api/gharial/{graph-name}/edge
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
		return gr.Collections, nil
	case 404:
		return []string{}, errors.New("Graph not found")
	default:
		return []string{}, errors.New("Invalid graph")
	}
}

// Add edge definition to graph
// Wrapper: POST /_api/gharial/{graph-name}/edge
func (g *Graph) AddEdgeDef(ed *EdgeDefinition) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	var gr graphResponse
	res, err := g.db.send("gharial", g.Name+"/edge", "POST", ed, &gr, &gr)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 201, 202:
		gr.Graph.db = g.db
		*g = gr.Graph
		return nil
	case 400:
		return errors.New("Unable to add edge definition")
	case 404:
		return errors.New("Graph not found")
	default:
		return errors.New("Invalid graph")
	}
}

// Remove an edge definition from the graph
// Wrapper: DELETE /_api/gharial/{graph-name}/edge/{definition-name}
func (g *Graph) RemoveEdgeDef(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}
	if col == "" {
		return errors.New("Invalid edge")
	}

	var gr graphResponse
	res, err := g.db.get("gharial", g.Name+"/edge/"+col, "DELETE", nil, &gr, &gr)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		gr.Graph.db = g.db
		*g = gr.Graph
		return nil
	case 400:
		return errors.New("No edge definition with this name is found in the graph")
	case 404:
		return errors.New("Graph not found")
	default:
		return errors.New("Invalid graph")
	}
}

// Replace an edge definition
// Wrapper: PUT /_api/gharial/{graph-name}/edge/{definition-name}
func (g *Graph) ReplaceEdgeDef(ed *EdgeDefinition) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	var gr graphResponse
	res, err := g.db.send("gharial", g.Name+"/edge/"+ed.Collection, "PUT", ed, &gr, &gr)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		gr.Graph.db = g.db
		*g = gr.Graph
		return nil
	case 400:
		return errors.New("No edge definition with this name is found in the graph")
	case 404:
		return errors.New("Graph not found")
	default:
		return errors.New("Invalid graph")
	}
}

// List vertex collection
// Wrapper: GET /_api/gharial/{graph-name}/vertex
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
		return gr.Collections, nil
	default:
		return []string{}, errors.New("Invalid graph")
	}
}

// Add vertex collection
// Wrapper: POST /_api/gharial/{graph-name}/vertex
func (g *Graph) AddVertexCol(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}

	var pay VertexDefinition
	pay.Collection = col

	var gr graphResponse
	res, err := g.db.send("gharial", g.Name+"/vertex", "POST", pay, &gr, &gr)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		gr.Graph.db = g.db
		*g = gr.Graph
		return nil
	case 404:
		return errors.New("Graph not found")
	default:
		return errors.New("Invalid graph")
	}
}

// Remove vertex collections
// Wrapper: DELETE /_api/gharial/{graph-name}/vertex/{collection-name}
func (g *Graph) RemoveVertexCol(col string) error {
	if g.db == nil {
		return errors.New("Invalid db")
	}
	if col == "" {
		return errors.New("Invalid collection")
	}

	var gr graphResponse
	res, err := g.db.get("gharial", g.Name+"/vertex/"+col, "DELETE", nil, &gr, &gr)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		gr.Graph.db = g.db
		*g = gr.Graph
		return nil
	case 400:
		return errors.New("Cannot remove collection")
	case 404:
		return errors.New("Graph not found")
	default:
		return errors.New("Invalid graph")
	}
}

type graphObj struct {
	V map[string]interface{} 	`json:"vertex"`
	E map[string]interface{} 	`json:"edge"`
}

// Create an egde
// Wrapper: POST /_api/gharial/{graph-name}/edge/{collection-name}
func (g *Graph) CreateEdge(col string, edge interface{}) error {
	if col == "" {
		return errors.New("Invalid collection name")
	}
	var gobj graphObj;
	res, err := g.db.send("gharial", g.Name+"/edge/"+col, "POST", edge, &gobj, &gobj)

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		subParse(gobj.E, edge)
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save edge")
	default:
		return errors.New("DB error")
	}
}

// Remove an edge
// Wrapper: DELETE /_api/gharial/{graph-name}/edge/{collection-name}/{edge-key}
func (g *Graph) RemoveEdge(col string, key string) error {
	// TODO: maybe make Edge struct {EDGE..., PAYLOAD...} and make Delete, Modify and Update methods of Edge
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
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection, graph or document id")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
	}
}

// Get an edge
// Wrapper: GET /_api/gharial/{graph-name}/edge/{collection-name}/{edge-key}
func (g *Graph) GetEdge(col string, key string, edge interface{}) error {
	if col == "" || key == "" {
		return errors.New("Invalid collection or key value")
	}
	var gobj graphObj;
	res, err := g.db.get("gharial", g.Name+"/edge/"+col+"/"+key, "GET", nil, &gobj, &gobj)

	if err != nil {
		return err
	}

	switch res.Status() {
	case 200:
		subParse(gobj.E, edge)
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save edge")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return errors.New("DB error")
	}
}

// Modify an edge
// Wrapper: PATCH /_api/gharial/{graph-name}/edge/{collection-name}/{edge-key}
func (g *Graph) PatchEdge(col string, key string, patch interface{}, edge interface{}) error {
	// TODO: maybe make Edge struct {EDGE..., PAYLOAD...} and make Delete, Modify and Update methods of Edge
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	var gobj graphObj
	res, err := g.db.send("gharial", g.Name+"/edge/"+col+"/"+key, "PATCH", patch, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		subParse(gobj.E, edge)
		return nil
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection ,graph or document id")
	default:
		return nil
	}
}

// Replace an edge
// Wrapper: PUT /_api/gharial/{graph-name}/edge/{collection-name}/{edge-key}
func (g *Graph) ReplaceEdge(col string, key string, edge interface{}) error {
	// TODO: maybe make Edge struct {EDGE..., PAYLOAD...} and make Delete, Modify and Update methods of Edge
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}
	var gobj graphObj

	res, err := g.db.send("gharial", g.Name+"/edge/"+col+"/"+key, "PUT", edge, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		subParse(gobj.E, edge)
		return nil
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection, graph or edge id")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
	}
}

// Create a vertex
// Wrapper: POST /_api/gharial/{graph-name}/vertex/{collection-name}
func (g *Graph) CreateVertex(col string, vertex interface{}) error {
	if col == "" {
		return errors.New("Invalid collection name")
	}
	var gobj graphObj
	res, err := g.db.send("gharial", g.Name+"/vertex/"+col, "POST", vertex, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201, 202:
		subParse(gobj.V, vertex)
		return nil
	case 404:
		return errors.New("Invalid collection or graph to save vertex")
	default:
		return nil
	}
}

// Remove a vertex
// Wrapper: DELETE /_api/gharial/{graph-name}/vertex/{collection-name}/{vertex-key}
// Remove Vertex
func (g *Graph) RemoveVertex(col string, key string) error {
	// TODO: maybe make Vertex struct {VERTEX..., PAYLOAD...} and make Delete, Modify and Update methods of Vertex
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
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection, graph or document id")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
	}
}

// Get a vertex
// GET /_api/gharial/{graph-name}/vertex/{collection-name}/{vertex-key}
func (g *Graph) GetVertex(col string, key string, vertex interface{}) error {
	// TODO: maybe make Vertex struct {VERTEX..., PAYLOAD...} and make Delete, Modify and Update methods of Vertex
	if key == "" || col == "" {
		return errors.New("Invalid key or collection to get")
	}

	var gobj graphObj
	res, err := g.db.get("gharial", g.Name+"/vertex/"+col+"/"+key, "GET", nil, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200:
		subParse(gobj.V, vertex)
		return nil
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection or graph")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
	}

}

// Modify a vertex
// PATCH /_api/gharial/{graph-name}/vertex/{collection-name}/{vertex-key}
func (g *Graph) PatchVertex(col string, key string, patch interface{}, vertex interface{}) error {
	// TODO: maybe make Vertex struct {VERTEX..., PAYLOAD...} and make Delete, Modify and Update methods of Vertex
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	var gobj graphObj
	res, err := g.db.send("gharial", g.Name+"/vertex/"+col+"/"+key, "PATCH", patch, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		subParse(gobj.V, vertex)
		return nil
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection, graph or document id")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
	}
}

// Replace a vertex
// PUT /_api/gharial/{graph-name}/vertex/{collection-name}/{vertex-key}
func (g *Graph) ReplaceVertex(col string, key string, vertex interface{}) error {
	// TODO: maybe make Vertex struct {VERTEX..., PAYLOAD...} and make Delete, Modify and Update methods of Vertex
	if col == "" || key == "" {
		return errors.New("Invalid key or collection")
	}

	var gobj graphObj
	res, err := g.db.send("gharial", g.Name+"/vertex/"+col+"/"+key, "PUT", vertex, &gobj, &gobj)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200, 202:
		subParse(gobj.V, vertex)
		return nil
		// TODO: need to add conditional update
	case 404:
		return errors.New("Invalid collection, graph or document id")
	case 412:
		return errors.New("If-match header is given, but the documents revision is different")
	default:
		return nil
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

// Executes a traversal
// POST /_api/traversal
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