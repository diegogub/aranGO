package aranGO

import (
	"errors"
	"time"
)

type CollectionParameters struct {
	CollectionOptions
	Id      string `json:"cid"`
	Version int    `json:"version"`
	Deleted bool   `json:"deleted"`
}

type CollectionDump struct {
	Parameters CollectionParameters `json:"parameters"`
	Indexes    []Index
}

type ReplicationState struct {
	Running     bool      `json:"running"`
	LastTick    string    `json:"lastLogTick"`
	TotalEvents int64     `json:"totalEvents"`
	Time        time.Time `json:"time"`
}

type ReplicationInventory struct {
	Collections []CollectionDump `json:"collections"`
	State       ReplicationState `json:"state"`
	Tick        string           `json:"tick"`
}

// Returns replication inventory
func (db *Database) Inventory() (*ReplicationInventory, error) {
	var rinv ReplicationInventory

	res, err := db.get("replication", "inventory", "GET", nil, &rinv, &rinv)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 405, 500:
		return nil, errors.New("Error when dumping replication info")
	default:
		return &rinv, nil
	}
}

type ServerInfo struct {
	Id      string `json:"serverId"`
	Version string `json:"version"`
}

type Logger struct {
	State  ReplicationState `json:"state"`
	Server ServerInfo       `json:"server"`
	Client []string         `json:"clients"`
}

func (db *Database) LoggerState() (*Logger, error) {
	var log Logger
	res, err := db.get("replication", "logger-state", "GET", nil, &log, &log)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 405, 500:
		return nil, errors.New("Logger state could not be determined")
	default:
		return &log, nil
	}
}

type ApplierConf struct {
	Endpoint       string `json:"endpoint,omitempty"`
	Database       string `json:"database,omitempty"`
	Username       string `json:"username,omitempty"`
	password       string `json:"password,omitempty"`
	Ssl            int    `json:"sslProtocol,omitempty"`
	ReConnect      int    `json:"maxConnectRetries,omitempty"`
	ConnectTimeout int    `json:"connectTimeOut,omitempty"`
	RequestTimeout int    `json:"requestTimeOut,omitempty"`
	Chunk          int    `json:"chunkSize,omitempty"`
	AutoStart      bool   `json:"autoStart,omitempty"`
	AdaptPolling   bool   `json:"adaptivePolling,omitempty"`
}

type ApplierProgress struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
	Fails   int       `json:"failedConnects"`
}

type ApplierState struct {
	Running       bool            `json:"running"`
	Progress      ApplierProgress `json:"progress"`
	TotalRequests int             `json:"totalRequests"`
	FailConnects  int             `json:"totalFailedConnects"`
	TotalEvents   int             `json:"totalEvents"`
	Time          time.Time       `json:"time"`
}

type Applier struct {
	State    ApplierState `json:"state"`
	Server   ServerInfo   `json:"server"`
	Endpoint string       `json:"endpoint"`
	Database string       `json:"database"`
}

func (db *Database) Applier() (*Applier, error) {
	var appl Applier
	res, err := db.get("replication", "applier-config", "GET", nil, &appl, &appl)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 405, 500:
		return nil, errors.New("Applier state could not be determined")
	default:
		return &appl, nil
	}
}

func (db *Database) ApplierConf() (*ApplierConf, error) {
	var appConf ApplierConf
	res, err := db.get("replication", "applier-config", "GET", nil, &appConf, &appConf)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 405, 500:
		return nil, errors.New("Applier state could not be determined")
	default:
		return &appConf, nil
	}
}

func (db *Database) SetApplierConf(appconf *ApplierConf) error {
	if appconf == nil {
		return errors.New("Invalid config")
	}
	res, err := db.send("replication", "applier-config", "PUT", appconf, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("Configuration is incomplete or malformed or applier running")
	case 405, 500:
		return errors.New("Error occurred while assembling the response.")
	default:
		return nil
	}
}

func (db *Database) StartReplication() error {
	res, err := db.send("replication", "applier-start", "PUT", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 400:
		return errors.New("Invalid applier configuration")
	case 405, 500:
		return errors.New("Error starting replication")
	default:
		return nil
	}
}

func (db *Database) StopReplication() error {
	res, err := db.send("replication", "applier-stop", "PUT", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 405, 500:
		return errors.New("Error stoping replication")
	default:
		return nil
	}
}

func (db *Database) ServerID() string {
	server := map[string]string{}
	_, err := db.get("replication", "server-id", "GET", nil, &server, &server)
	if err != nil {
		return ""
	}
	return server["serverId"]
}
