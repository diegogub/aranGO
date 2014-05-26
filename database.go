package aranGO

import(
  "net/url"
  "errors"
  nap "github.com/jmcvetta/napping"
)

// Database information to build requests
type Database struct {
  Name    string
  Log     bool
  Unsafe  bool
  Auth    bool

  host    string
  baseURL string
  sess    *nap.Session
  user    string
  password string
}

func NewDatabase(name string,host string,username string,password string) (*Database,error){
  var db Database
  if name != ""{
    db.Name = name
  }else{
    return nil,errors.New("Invalid DB name")
  }

  // check valid host
  urlhost , err:= url.Parse(host)
  if err != nil{
    return nil,err
  }

  if urlhost.Host == "" {
    return nil,errors.New("Invalid DB host")
  }

  if urlhost.Scheme != "http" && urlhost.Scheme != "https"{
    return nil,errors.New("Invalid url scheme")
  }

  db.host = host
  db.baseURL = db.host + "/_db/" + db.Name + "/_api/"
  db.user = username
  db.password = password

  return &db,nil
}

// Do a request to test if the database is up or usern authorized to use it
func (d *Database) Ping() error{
  if d.Auth {
    if d.sess == nil{
      var s nap.Session
      s.Log = d.Log
      s.UnsafeBasicAuth = d.Unsafe
      s.Userinfo = url.UserPassword(d.user,d.password)
      d.sess = &s // set session to use
    }
    url := d.host + "/_db/" + d.Name + "/_api/version"
    resp , err := d.sess.Get(url,nil,nil,nil)

    if err != nil{
      return err
    }
    // check if 200
    switch resp.Status(){
      case 200:
        return nil
      default:
        return errors.New("Invalid DB ,host or auth data to connect")
    }

  }else{
    var s nap.Session
    s.Log = d.Log
    s.UnsafeBasicAuth = d.Unsafe
    d.sess = &s // set session to use

    url := d.host + "/_db/" + d.Name + "/_api/version"
    resp , err := nap.Get(url,nil,nil,nil)

    if err != nil{
      return err
    }
    // check if 200
    switch resp.Status(){
      case 200:
        return nil
      default:
        return errors.New("Invalid DB or host to connect")
    }
  }
  return nil
}

func (d *Database) get(resource string,id string,method string,param *nap.Params, result, err interface{}) (*nap.Response,error){
  url := d.buildRequest(resource,id)
  var r *nap.Response
  var e error

  switch method{
    case "OPTIONS":
      r,e = d.sess.Options(url,result,err)
    case "HEAD":
      r,e = d.sess.Head(url,result,err)
    case "DELETE":
      r,e = d.sess.Delete(url,result,err)
    default:
      r,e = d.sess.Get(url,param,result,err)
  }

  return r,e
}

func (d *Database) send(resource string,id string, method string,payload ,result, err interface{}) (*nap.Response,error){
  url := d.buildRequest(resource,id)
  var r *nap.Response
  var e error

  switch method {
    case "POST":
      r,e = d.sess.Post(url,payload,result,err)
    case "PUT":
      r,e = d.sess.Put(url,payload,result,err)
    case "PATCH":
      r,e = d.sess.Patch(url,payload,result,err)
  }
  return r,e
}

func (db Database) buildRequest(t string,id string) string{
  var r string
  if id == "" {
    r = db.baseURL + t
  }else{
    r = db.baseURL + t + "/" + id
  }
  return r
}

type DatabaseResult struct{
  Result []string `json:"result"`
  Error  bool     `json:"error"`
  Code   int      `json:"code"`
}

// Returns Collection attached to current Database
func (db Database) Col(name string) *Collection{
  var col Collection
  // need to validate this more
  if name != ""{
    col.name = name
  }else{
    col.name = "test"
  }
  col.db = &db
  return &col
}
