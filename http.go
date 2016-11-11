package aranGO

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	genlog "github.com/hnakamur/gentleman-log"
	"gopkg.in/h2non/gentleman.v1"
	c "gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugins/auth"
)

const (
	batchRequestBoundary = "XXXsubpartXXX"
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
		err = saveResponse(bytes.NewReader(b), genRes.StatusCode, r)
		if err != nil {
			return nil, err
		}
	}

	return &response{
		rawResponse: genRes.RawResponse,
	}, nil
}

func saveResponse(reader io.Reader, statusCode int, r *request) error {
	if statusCode < http.StatusMultipleChoices {
		if r.result != nil {
			decoder := json.NewDecoder(reader)
			err := decoder.Decode(r.result)
			if err != nil {
				return fmt.Errorf("failed to unmarshal result: %v", err)
			}
		}
	} else {
		if r.errMsg != nil {
			decoder := json.NewDecoder(reader)
			err := decoder.Decode(r.errMsg)
			if err != nil {
				return fmt.Errorf("failed to unmarshal error message: %v", err)
			}
		}
	}

	return nil
}

func (c *httpClient) BatchPost(url, batchUrl string, payloads, results, errs []interface{}) ([]response, error) {
	requests := make([]request, 0, len(payloads))

	for idx, _ := range payloads {
		// Errors and results are saved to the payload
		requests = append(requests, request{
			method:  "POST",
			url:     url,
			payload: payloads[idx],
			result:  &results[idx],
			errMsg:  &errs[idx],
		})
	}

	return c.sendBatch(batchUrl, requests)
}

// Support for batch requests as described here:
// 	https://docs.arangodb.com/3.0/HTTP/BatchRequest/
func (c *httpClient) sendBatch(batchUrl string, requests []request) ([]response, error) {
	if len(requests) == 0 {
		return nil, errors.New("Empty requests sequence")
	}

	// Generate multipart data for the array of requests
	data := bytes.NewBuffer([]byte{})
	httpRequests, err := generateBatchRequests(data, requests)
	if err != nil {
		return nil, err
	}

	// Setup HTTP request and send it
	genReq := c.cli.Request().Method("POST").URL(batchUrl)
	genReq.SetHeader("Content-Type", "multipart/form-data; boundary="+batchRequestBoundary)
	genReq.Body(data)

	genRes, err := genReq.Send()
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if genRes.StatusCode >= http.StatusMultipleChoices {
		// We failed to process batch of requests, notify every requestor
		// (we shouldn't get multipart response in this case)
		b := genRes.Bytes()
		if b != nil {
			reader := bytes.NewReader(b)
			for idx, _ := range requests {
				saveResponse(reader, genRes.StatusCode, &requests[idx])
			}
		}
		return nil, fmt.Errorf("failed to process request: %v", err)
	}

	// Parse multipart response
	httpResponses := make([]response, 0, len(requests))
	mpReader := multipart.NewReader(bytes.NewReader(genRes.Bytes()), batchRequestBoundary)
	for {
		part, err := mpReader.NextPart()
		if err == io.EOF {
			break
		}

		contentId, err := strconv.Atoi(part.Header.Get("Content-Id"))
		if contentId > len(httpRequests) || contentId <= 0 {
			err = errors.New("out of range")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to process response, invalid Content-Id: %v", err)
		}

		httpResponse, err := http.ReadResponse(bufio.NewReader(part), httpRequests[contentId-1])
		if err != nil {
			return nil, fmt.Errorf("failed to process response: %v", err)
		}

		saveResponse(httpResponse.Body, httpResponse.StatusCode, &requests[contentId-1])
		httpResponses = append(httpResponses, response{rawResponse: httpResponse})
	}

	return httpResponses, nil
}

func generateBatchRequests(data *bytes.Buffer, requests []request) ([]*http.Request, error) {
	httpRequests := make([]*http.Request, 0, len(requests))

	mpWriter := multipart.NewWriter(data)
	mpWriter.SetBoundary(batchRequestBoundary)
	for contentId, r := range requests {
		httpRequest, err := writeMultiPartRequest(mpWriter, contentId+1, &r)
		if err != nil {
			return nil, fmt.Errorf("failed to create multi-part: %v", err)
		}

		httpRequests = append(httpRequests, httpRequest)
	}

	err := mpWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multi-part: %v", err)
	}

	return httpRequests, nil
}

func generateBatchUrl(r *request) *url.URL {
	// Generate url with params, but without host/scheme
	// (they are omitted in batch requests)
	url, err := url.Parse(r.url)
	if err != nil {
		return nil
	}

	if r.params != nil {
		query := url.Query()
		for key, value := range r.params {
			query.Add(key, value)
		}
		url.RawQuery = query.Encode()
	}

	url.Scheme = ""
	url.Host = ""

	return url
}

func writeMultiPartRequest(mp *multipart.Writer, contentId int, r *request) (*http.Request, error) {
	header := make(textproto.MIMEHeader)
	header.Add("Content-Type", "application/x-arango-batchpart")
	header.Add("Content-Id", strconv.Itoa(contentId))

	pw, err := mp.CreatePart(header)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method:     r.method,
		URL:        generateBatchUrl(r),
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	req.Write(pw)

	if r.payload != nil {
		jsonEncoder := json.NewEncoder(pw)
		jsonEncoder.Encode(r.payload)
	}

	return req, nil
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

func logBody(label, contentType string, body []byte) (err error) {
	out := bytes.NewBufferString(label)
	if len(body) > 0 {
		out.WriteByte('\n')

		if strings.HasPrefix(contentType, "application/json") {
			err = json.Indent(out, body, "", "  ")
		} else {
			out.Write(body)
		}
	}

	if err == nil {
		log.Println(out.String())
	}
	return
}

func logFunc(ctx *c.Context, req *http.Request, res *http.Response, reqBody, resBody []byte) error {
	log.Println("--------------------------------------------------------------------------------")
	log.Println("REQUEST")
	log.Println("--------------------------------------------------------------------------------")
	log.Printf("%s %s", req.Method, req.URL)
	logHeader(req.Header)

	err := logBody("Payload:", req.Header.Get("Content-Type"), reqBody)
	if err != nil {
		return err
	}

	log.Println("--------------------------------------------------------------------------------")
	log.Println("RESPONSE")
	log.Println("--------------------------------------------------------------------------------")
	log.Printf("Status: %d", res.StatusCode)
	logHeader(res.Header)
	err = logBody("Body:", req.Header.Get("Content-Type"), resBody)
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
