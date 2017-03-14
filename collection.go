package aranGO

// TODO Must Implement revision control
import (
	"errors"

	nap "github.com/diegogub/napping"
)

// Options to create collection
type CollectionOptions struct {
	Name        string                 `json:"name"`
	Type        uint                   `json:"type"`
	Sync        bool                   `json:"waitForSync,omitempty"`
	Compact     bool                   `json:"doCompact,omitempty"`
	JournalSize int                    `json:"journalSize,omitempty"`
	System      bool                   `json:"isSystem,omitempty"`
	Volatile    bool                   `json:"isVolatile,omitempty"`
	Keys        map[string]interface{} `json:"keyOptions,omitempty"`
	// Count
	Count int64 `json:"count"`
	// Cluster
	Shards    int      `json:"numberOfShards,omitempty"`
	ShardKeys []string `json:"shardKeys,omitempty"`
}

func NewCollectionOptions(name string, sync bool) *CollectionOptions {
	var copt CollectionOptions
	copt.Name = name
	copt.Sync = sync
	return &copt
}

func (opt *CollectionOptions) IsEdge() {
	opt.Type = 3
	return
}

func (opt *CollectionOptions) IsDocument() {
	opt.Type = 2
	return
}

// Sets always-sync to true
func (opt *CollectionOptions) MustSync() {
	opt.Sync = true
	return
}

// Sets if collection must be Volatile.
func (opt *CollectionOptions) IsVolatile() {
	opt.Volatile = true
	return
}

// Sets custom journal size
func (opt *CollectionOptions) Journal(size int) {
	if size <= 0 {
		size = 1
	}
	// convert to kb
	size = size * 1024 * 1024
	opt.JournalSize = size
}

// Sets the number of shards for a collection
func (opt *CollectionOptions) Shard(num int) {
	if num <= 0 {
		num = 1
	}

	opt.Shards = num
}

func (opt *CollectionOptions) ShardKey(keys []string) {
	if len(keys) == 0 {
		return
	}
	if opt.ShardKeys == nil {
		opt.ShardKeys = make([]string, 0)
	}
	opt.ShardKeys = keys
}

// Basic Collection struct
type Collection struct {
	db     *Database `json:"db"`
	Name   string    `json:"name"`
	System bool      `json:"isSystem"`
	Status int       `json:"status"`
	// 3 = Edges , 2 =  Documents
	Type     int    `json:"type"`
	policy   string `json:"-"`
	revision bool   `json:"-"`
}

// Load collection
func (col *Collection) Load() error {
	// set count to false to speed up loading
	payload := map[string]bool{"count": false}
	res, err := col.db.send("collection", col.Name+"/load", "PUT", payload, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400, 404:
		return errors.New("Invalid collection to load")
	default:
		return nil
	}
}

//Count all documents in collection
func (col *Collection) Count() int64 {
	var cop CollectionOptions
	res, err := col.db.get("collection", col.Name+"/count", "GET", nil, &cop, &cop)
	if err != nil {
		return 0
	}

	switch res.Status() {
	case 400, 404:
		return 0
	default:
		return cop.Count
	}
}

// Save saves doc into collection, doc should have Document Embedded to retrieve error and Key later.
func (col *Collection) Save(doc interface{}) error {
	var err error
	var res *nap.Response

	if col.Type == 2 {
		res, err = col.db.send("document?collection="+col.Name, "", "POST", doc, &doc, &doc)
	} else {
		return errors.New("Trying to save doc into EdgeCollection")
	}

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201:
		return nil
	case 202:
		return nil
	case 400:
		return errors.New("Invalid document json")
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

// Save Edge into Edges collection
func (col *Collection) SaveEdge(doc map[string]interface{}, from string, to string) error {
	var err error
	var res *nap.Response
	
	doc["_from"] = from
	doc["_to"] = to

	if col.Type == 3 {
		res, err = col.db.send("document?collection="+col.Name, "", "POST", doc, &doc, &doc)
	} else {
		return errors.New("Trying to save document into Edge-Collection")
	}

	if err != nil {
		return err
	}

	if res.Status() != 201 && res.Status() != 202 {
		return errors.New("Unable to save document error")
	}

	return nil

}

//Get vertex relations
func (col *Collection) Edges(start string, direction string, result interface{}) error {
	if start == "" {
		return errors.New("Invalid start vertex")
	}
	if direction != "in" && direction != "out" {
		direction = "any"
	}

	if col.Type == 2 {
		return errors.New("Invalid edge collection: " + col.Name)
	}

	res, err := col.db.get("edges", col.Name+"?vertex="+start+"&direction="+direction, "GET", nil, &result, &result)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 200:
		return nil
	default:
		return errors.New("failed to get edges")
	}
}

// Relate documents in edge collection
func (col *Collection) Relate(from string, to string, label  map[string]interface{}) error {
	if col.Type == 2 {
		return errors.New("Invalid collection to add Edge: " + col.Name)
	}
	if from == "" || to == "" {
		return errors.New("from or to documents don't exist")
	}
	if label == nil {
	    label = make(map[string]interface{})
	}

	return col.SaveEdge(label, from, to)
}

//Get Document
func (col *Collection) Get(key string, doc interface{}) error {
	var err error

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		_, err = col.db.get("document", col.Name+"/"+key, "GET", nil, &doc, &doc)
	} else {
		_, err = col.db.get("edge", col.Name+"/"+key, "GET", nil, &doc, &doc)
	}

	if err != nil {
		return err
	}

	return nil
}

// Replace document
func (col *Collection) Replace(key string, doc interface{}) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.send("document", col.Name+"/"+key, "PUT", doc, &doc, &doc)
	} else {
		res, err = col.db.send("edge", col.Name+"/"+key, "PUT", doc, &doc, &doc)
	}

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201:
		return nil
	case 202:
		return nil
	case 400:
		return errors.New("Invalid json")
	case 404:
		return errors.New("Collection or document was not found")
	default:
		return nil
	}
}

func (col *Collection) Patch(key string, doc interface{}) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.send("document", col.Name+"/"+key, "PATCH", doc, &doc, &doc)
	} else {
		res, err = col.db.send("edge", col.Name+"/"+key+"?rev=", "PATCH", doc, &doc, &doc)
	}

	if err != nil {
		return err
	}

	switch res.Status() {
	case 201:
		return nil
	case 202:
		return nil
	case 400:
		return errors.New("Body does not contain a valid JSON representation of a document.")
	case 404:
		return errors.New("Collection or document was not found")
	default:
		return nil
	}
}

func (col *Collection) Delete(key string) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.get("document", col.Name+"/"+key, "DELETE", nil, nil, nil)
	} else {
		res, err = col.db.get("edge", col.Name+"/"+key, "DELETE", nil, nil, nil)
	}
	if err != nil {
		return err
	}

	switch res.Status() {
	case 202, 200:
		return nil
	default:
		return errors.New("Document don't exist or revision error")

	}
}

// Get list of collections from any database
func Collections(db *Database) error {
	var err error
	var res *nap.Response

	// get all non-system collections
	res, err = db.get("collection?excludeSystem=true", "", "GET", nil, db, db)
	if err != nil {
		return err
	}

	if res.Status() == 200 {
		var result struct {
			Result []Collection `json:"result"`
		}
		if err = res.Unmarshal(&result); err != nil {
			return err
		}
		for i, _ := range result.Result {
			result.Result[i].db = db
			db.Collections = append(db.Collections, result.Result[i])
		}
		return nil
	} else {
		return errors.New("Failed to retrieve collections from Database")
	}
}

// check if a key is unique
func (c *Collection) Unique(key string, value interface{}, update bool, index string) (bool, error) {
	var cur *Cursor
	var err error
	switch index {
	case "hash":
		// must implement other simple query function s
	case "skip-list":

	default:
		cur, err = c.Example(map[string]interface{}{key: value}, 0, 2)
	}
	if err != nil {
		return false, err
	}

	var result map[string]interface{}
	result = make(map[string]interface{})

	if !update {
		if cur.Amount > 0 {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		if cur.Amount == 0 {
			return false, nil
		} else {
			if cur.Amount == 1 {
				cur.FetchOne(&result)
				if result[key].(string) == value {
					return true, nil
				} else {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
		return true, nil
	}
}

// lSimple Queries
//helper structs
type singleDoc struct {
	Doc interface{} `json:"document"`
}

//
func (c *Collection) All(skip, limit int) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	query := map[string]interface{}{"collection": c.Name, "skip": skip, "limit": limit}
	res, err := c.db.send("simple", "all", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

//Simple query by example
func (c *Collection) Example(doc interface{}, skip, limit int) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	query := map[string]interface{}{"collection": c.Name, "example": doc, "skip": skip, "limit": limit}
	res, err := c.db.send("simple", "by-example", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

// Returns first document in example query
func (c *Collection) First(example, doc interface{}) error {
	var d singleDoc
	d.Doc = doc
	query := map[string]interface{}{"collection": c.Name, "example": example}
	// sernd request
	res, err := c.db.send("simple", "first-example", "PUT", query, &d, &doc)

	if err != nil {
		return err
	}

	if res.Status() == 200 {
		return nil
	} else {
		return errors.New("Failed to execute query")
	}
}

//Coditional query using skiplist index
func (c *Collection) ConditionSkipList(condition string, skip int, limit int, index string) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	if condition == "" {
		return nil, errors.New("Invalid conditions")
	}
	query := map[string]interface{}{"collection": c.Name, "index": index, "condition": condition, "skip": skip, "limit": limit}
	res, err := c.db.send("simple", "by-condition-skiplist", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

//Coditional query using bitarray index
func (c *Collection) ConditionBitArray(condition string, skip int, limit int, index string) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	if condition == "" {
		return nil, errors.New("Invalid conditions")
	}
	query := map[string]interface{}{"collection": c.Name, "index": index, "condition": condition, "skip": skip, "limit": limit}
	res, err := c.db.send("simple", "by-condition-bitarray", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

//Return random number
func (c *Collection) Any(doc interface{}) error {
	var d singleDoc
	d.Doc = doc
	query := map[string]interface{}{"collection": c.Name}
	res, err := c.db.send("simple", "any", "PUT", query, &d, &doc)

	if err != nil {
		return err
	}

	if res.Status() == 200 {
		return nil
	} else {
		return errors.New("Failed to execute query")
	}
}

//Get all indexs
func (c *Collection) Indexes() (map[string]Index, error) {
	var indexes Indexes
	_, err := c.db.get("index?collection="+c.Name, "", "GET", nil, &indexes, &indexes)
	if err != nil {
		return indexes.IndexMap, err
	}

	return indexes.IndexMap, err
}

// Delete Index
func (c *Collection) DeleteIndex(id string) error {
	if id == "" {
		return errors.New("Invalid id")
	}

	res, err := c.db.get("index", "id", "DELETE", nil, nil, nil)
	if err != nil {
		return err
	}

	if res.Status() == 404 {
		return errors.New("Index does not exist")
	}

	return nil
}

//Create cap constraint
func (c *Collection) SetCap(size int64, bysize int64) error {
	if size == 0 && bysize == 0 {
		return errors.New("Invalid num and byte size")
	}

	if bysize > 0 && bysize < 16384 {
		return errors.New("Invalid byte size, must be at least 16384")
	}

	capindex := map[string]interface{}{"type": "cap", "size": size, "byteSize": bysize}

	res, err := c.db.send("index?collection="+c.Name, "", "POST", &capindex, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("Invalid size or byte-size values")
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

func (c *Collection) CreateHash(unique bool, fields ...string) error {
	hashindex := map[string]interface{}{"type": "hash", "unique": unique, "fields": fields}

	res, err := c.db.send("index?collection="+c.Name, "", "POST", &hashindex, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("There are documents violating the uniqueness")
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

func (c *Collection) CreateSkipList(unique bool, fields ...string) error {
	skiplist := map[string]interface{}{"type": "skiplist", "unique": unique, "fields": fields}

	res, err := c.db.send("index?collection="+c.Name, "", "POST", &skiplist, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("There are documents violating the uniqueness")
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

func (c *Collection) CreateGeoIndex(unique bool, geojson bool, fields ...string) error {
	geoindex := map[string]interface{}{"type": "geo", "geoJson": geojson, "unique": unique, "fields": fields}

	res, err := c.db.send("index?collection="+c.Name, "", "POST", &geoindex, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("There are documents violating the uniqueness")
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

func (c *Collection) Near(lat float64, lon float64, distance bool, geo string, skip, limit int) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	var query map[string]interface{}
	if distance {
		query = map[string]interface{}{"collection": c.Name, "latitude": lat, "longitude": lon, "distance": "distance", "skip": skip, "limit": limit}
	} else {
		query = map[string]interface{}{"collection": c.Name, "latitude": lat, "longitude": lon, "skip": skip, "limit": limit}
	}

	if len(geo) > 0 {
		query["geo"] = geo
	}

	res, err := c.db.send("simple", "near", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

func (c *Collection) WithIn(radius float64, lat float64, lon float64, distance bool, geo string, skip, limit int) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}
	var query map[string]interface{}

	if distance {
		query = map[string]interface{}{"collection": c.Name, "radius": radius, "latitude": lat, "longitude": lon, "distance": "distance", "skip": skip, "limit": limit}
	} else {
		query = map[string]interface{}{"collection": c.Name, "radius": radius, "latitude": lat, "longitude": lon, "skip": skip, "limit": limit}
	}

	if len(geo) > 0 {
		query["geo"] = geo
	}

	res, err := c.db.send("simple", "near", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

func (c *Collection) CreateFullText(min int, fields ...string) error {
	full := map[string]interface{}{"type": "fulltext", "minLength": min, "fields": fields}

	res, err := c.db.send("index?collection="+c.Name, "", "POST", &full, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 404:
		return errors.New("Collection does not exist")
	default:
		return nil
	}
}

func (c *Collection) FullText(q string, atr string, skip, limit int) (*Cursor, error) {
	var cur Cursor
	if skip < 0 || limit < 0 {
		return nil, errors.New("Invalid skip or limit")
	}

	query := map[string]interface{}{"collection": c.Name, "query": q, "attribute": atr, "skip": skip, "limit": limit}
	res, err := c.db.send("simple", "fulltext", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

type Indexes struct {
	Indexes  []Index
	IndexMap map[string]Index `json:"identifiers"`
}

type Index struct {
	Id        string   `json:"id"`
	Type      string   `json:"type"`
	Unique    bool     `json:"unique"`
	MinLength int      `json:"minLength"`
	Fields    []string `json:"fields"`
	Size      int64    `json:"size"`
}
