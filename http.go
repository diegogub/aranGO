package aranGO

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	genlog "github.com/hnakamur/gentleman-log"
	"gopkg.in/h2non/gentleman.v1"
	c "gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugins/auth"
)

type httpClient struct {
	cli *gentleman.Client
}

func newHTTPClient(user, password string, log bool) *httpClient {
	cli := gentleman.New()
	if user != "" {
		cli.Use(auth.Basic(user, password))
	}
	if log {
		cli.Use(genlog.Log(genlog.Config{LogFunc: logFunc}))
	}
	return &httpClient{cli: cli}
}

func (c *httpClient) Delete(url string, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method: "DELETE",
		url:    url,
		result: result,
		errMsg: errMsg,
	})
}

func (c *httpClient) Get(url string, p map[string]string, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method: "GET",
		url:    url,
		params: p,
		result: result,
		errMsg: errMsg,
	})
}

func (c *httpClient) Head(url string, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method: "HEAD",
		url:    url,
		result: result,
		errMsg: errMsg,
	})
}

func (c *httpClient) Options(url string, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method: "OPTIONS",
		url:    url,
		result: result,
		errMsg: errMsg,
	})
}

func (c *httpClient) Patch(url string, payload, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method:  "PATCH",
		url:     url,
		payload: payload,
		result:  result,
		errMsg:  errMsg,
	})
}

func (c *httpClient) Post(url string, payload, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method:  "POST",
		url:     url,
		payload: payload,
		result:  result,
		errMsg:  errMsg,
	})
}

func (c *httpClient) Put(url string, payload, result, errMsg interface{}) (*response, error) {
	return c.send(&request{
		method:  "PUT",
		url:     url,
		payload: payload,
		result:  result,
		errMsg:  errMsg,
	})
}

func (c *httpClient) send(r *request) (*response, error) {
	genReq := c.cli.Request().Method(r.method).URL(r.url)
	if r.params != nil {
		genReq = genReq.Params(r.params)
	}
	if r.payload != nil {
		genReq = genReq.JSON(r.payload)
	}
	genRes, err := genReq.Send()
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	b := genRes.Bytes()
	if b != nil {
		if genRes.StatusCode < http.StatusMultipleChoices {
			if r.result != nil {
				err = json.Unmarshal(b, r.result)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal result: %v", err)
				}
			}
		} else {
			if r.errMsg != nil {
				err = json.Unmarshal(b, r.errMsg)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal error message: %v", err)
				}
			}
		}
	}

	return &response{
		rawResponse: genRes.RawResponse,
	}, nil
}

type request struct {
	method  string
	url     string
	params  map[string]string
	payload interface{}
	result  interface{}
	errMsg  interface{}
}

type response struct {
	rawResponse *http.Response
}

func (r *response) Status() int {
	return r.rawResponse.StatusCode
}

func logHeader(h http.Header) {
	if len(h) == 0 {
		return
	}
	keys := make([]string, len(h))
	for k, _ := range h {
		keys = append(keys, k)
	}
	log.Println("Header:")
	for _, k := range keys {
		for _, v := range h[k] {
			log.Printf("%s: %s", k, v)
		}
	}
}

func logBody(label string, body []byte) error {
	out := bytes.NewBufferString(label)
	if len(body) > 0 {
		out.WriteByte('\n')
		err := json.Indent(out, body, "", "  ")
		if err != nil {
			return err
		}
	}
	log.Println(out.String())
	return nil
}

func logFunc(ctx *c.Context, req *http.Request, res *http.Response, reqBody, resBody []byte) error {
	log.Println("--------------------------------------------------------------------------------")
	log.Println("REQUEST")
	log.Println("--------------------------------------------------------------------------------")
	log.Printf("%s %s", req.Method, req.URL)
	logHeader(req.Header)
	err := logBody("Payload:", reqBody)
	if err != nil {
		return err
	}

	log.Println("--------------------------------------------------------------------------------")
	log.Println("RESPONSE")
	log.Println("--------------------------------------------------------------------------------")
	log.Printf("Status: %d", res.StatusCode)
	logHeader(res.Header)
	err = logBody("Body:", resBody)
	if err != nil {
		return err
	}
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.Contains(file, "diegogub/aranGO") && !strings.Contains(file, "diegogub/aranGO/http.go") {
			log.Printf("Caller: %s:%d", filepath.Base(file), line)
			for i++; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				if !strings.Contains(file, "diegogub/aranGO") {
					break
				}
				log.Printf("Caller: %s:%d", filepath.Base(file), line)
			}
			break
		}
	}
	return nil
}
