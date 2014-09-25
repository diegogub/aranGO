package aranGO

// Configure to start testing
var (
	TestCollection = ""
	TestDoc        DocTest
	TestDbName     = ""
	TestUsername   = ""
	TestPassword   = ""
	TestString     = "test string"
	verbose        = false
	TestServer     = "http://localhost:8529"
	s              *Session
)

// document to test
type DocTest struct {
	Document // arango Document to save id, key, rev
	Text     string
}
