package aranGO

import (
	"errors"
	nap "github.com/diegogub/napping"
	"net/url"
	"regexp"
)

type Session struct {
	host string
	safe bool
	nap  *nap.Session
	dbs  Databases
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Active   bool   `json:"active"`
	Extra    string `json:"extra"`
}

type Databases struct {
	List []string `json:"result" `
}

type auxCurrentDB struct {
	Db Database `json:"result"`
}

// Connects to Database
func Connect(host, user, password string, log bool) (*Session, error) {
	var sess Session
	var s nap.Session
	var dbs Databases
	var err error
	var request string
	s.Log = log
	// default unsafe
	s.UnsafeBasicAuth = true

	if user != "" {
		s.Userinfo = url.UserPassword(user, password)
	}

	request = host + "/_db/_system/_api/version"
	resp, err := s.Get(request, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	// check if 200
	switch resp.Status() {
	case 200:
		// load Databases
		request = host + "/_api/database/user"
		_, err = s.Get(request, nil, &dbs, nil)
		sess.dbs.List = dbs.List

		if err != nil {
			return nil, err
		}
		sess.nap = &s
		sess.host = host
		return &sess, nil
	default:
		return nil, errors.New("Invalid host or auth data to connect")
	}

}

// Show current database
func (s *Session) CurrentDB() (*Database, error) {
	var db auxCurrentDB
	sdb := s.DB("_system")

	res, err := sdb.get("database", "current", "GET", nil, &db, &db)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 200:
		return &db.Db, nil
	case 400:
		return nil, errors.New("Invalid request")
	case 404:
		return nil, errors.New("Database not found")
	default:
		return nil, errors.New("Invalid error code")
	}
}

// List available databases
func (s *Session) AvailableDBs() ([]string, error) {
	var dbs Databases
	db := s.DB("_system")
	if db == nil {
		return nil, errors.New("invalid db")
	}

	res, err := db.get("database", "user", "GET", nil, &dbs, &dbs)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 200:
		return dbs.List, nil
	case 400:
		return nil, errors.New("request is invalid")
	default:
		return dbs.List, nil
	}

}

// Create database
func (s *Session) CreateDB(name string, users []User) error {
	body := make(map[string]interface{})
	// validate name
	reg, err := regexp.Compile(`^[A-z]+[0-9\-_]*`)

	if err != nil {
		return err
	}
	if !reg.MatchString(name) {
		return errors.New("Invalid database name")
	}
	if err != nil {
		return err
	}

	body["name"] = name
	if users != nil && len(users) > 0 {
		body["users"] = users
	}
	// use _system database
	sdb := s.DB("_system")

	res, err := sdb.send("database", "", "POST", &body, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201:
		// update Databases
		var dbs Databases
		request := s.host + "/_api/database/user"
		_, err = s.nap.Get(request, nil, &dbs, nil)
		s.dbs.List = dbs.List
		return nil
	case 400:
		return errors.New("Request parameters are invalid or database already exist")
	case 403:
		return errors.New("Must be _system database")
	case 409:
		return errors.New("Database with the specified name already exists")
	default:
		// update Databases
		var dbs Databases
		request := s.host + "/_api/database/user"
		_, err = s.nap.Get(request, nil, &dbs, nil)
		s.dbs.List = dbs.List
		return nil
	}

}

//Drops database
func (s *Session) DropDB(name string) error {
	// use _system database
	sdb := s.DB("_system")

	res, err := sdb.get("database", name, "DELETE", nil, nil, nil)
	if err != nil {
		return err
	}

	switch res.Status() {
	case 201:
		return nil
	case 400:
		return errors.New("Request is malformed")
	case 403:
		return errors.New("Request was not executed in the _system database.")
	case 404:
		return errors.New("Database could not be found")
	default:
		return nil
	}
}

// DB returns database
func (s *Session) DB(name string) *Database {
	var db Database
	var found bool
	if name != "" {
		db.Name = name
	} else {
		return nil
	}

	for _, dbname := range s.dbs.List {
		if dbname == name {
			found = true
			break
		}
	}
	if found {
		db.baseURL = s.host + "/_db/" + db.Name + "/_api/"
		db.sess = s
		// load collections
		Collections(&db)
	} else {
		panic("Invalid DB")
	}

	return &db

}

func (s *Session) Safe(safe bool) {
	s.safe = safe
	return
}
