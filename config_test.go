package aranGO

import (
	"fmt"
	"log"
	"os"
)

// Configure to start testing
var (
	TestCollection = "TestCollection"
	TestDoc        DocTest
	TestDbName     = "TestDbName"
	TestUsername   = "TestUsername"
	TestPassword   = "TestPassword"
	TestString     = "test string"
	verbose        = false
	TestServer     = "http://localhost:8529"
	s              *Session
)

func init() {
	if wercker := os.Getenv("WERCKER"); wercker == "true" {
		log.Printf("Found WERCKER env!")
		arangoPort := os.Getenv("ARANGODB_PORT_8529_TCP_PORT")
		arangoIP := os.Getenv("ARANGODB_PORT_8529_TCP_ADDR")
		TestServer = fmt.Sprintf("http://%s:%s", arangoIP, arangoPort)
	}
	log.Printf("using TestServer at %s", TestServer)
}

// document to test
type DocTest struct {
	Document // arango Document to save id, key, rev
	Text     string
}
