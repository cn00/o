package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/pkg/errors"
	)

const (
	maxRetry      = 120
	retryInterval = 60 * time.Second
)

var (
	httpGetClient  = &http.Client{Timeout: time.Duration(650 * time.Second)}
	httpPostClient = &http.Client{Timeout: time.Duration(60 * time.Minute)}
)

var (
	App       *cli.App
	AppSecret string
)

func HttpGet(url string, v interface{}) error {
	// client do error 対策 get仅
	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	var err error
	retry(func() bool {
		err = httpGet(url, v)
		return tryAgain(err)
	})
	return err
}

func HttpPost(url string, body []byte, v interface{}) error {
	var err error
	retry(func() bool {
		http.DefaultTransport.(*http.Transport).DisableKeepAlives = false
		err = httpPost(url, body, v) // POST失败时不知道服务器的处理是否成功，所以不重试
		return tryAgain(err)
	})
	return err
}

func retry(f func() bool) {
	trial := maxRetry
	for trial > 0 {
		trial--
		again := f()
		if !again || trial <= 0 {
			break
		}
		time.Sleep(retryInterval)
	}
}

func tryAgain(err error) bool {
	if err, ok := errors.Cause(err).(*httpError); ok {
		if err.StatusCode >= http.StatusInternalServerError {
			log.Println("[INFO] API Server is temporary unavailable:", err)
			log.Println("[INFO] Retrying...")
			return true
		}
	}
	return false
}

func httpGet(url string, v interface{}) error {
	log.Println("[INFO] httpGet:", url)
	req, err := newRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := httpGetClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "client do error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := handleResponseError(resp)
		return errors.Wrap(err, "response error")
	}

	return readBody(resp.Body, v)
}

func httpPost(url string, body []byte, v interface{}) error {
	log.Println("[INFO] httpPost:", url, len(body))
	req, err := newRequest("POST", url, body)
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	resp, err := httpPostClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "client do error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := handleResponseError(resp)
		return errors.Wrap(err, "response error")
	}

	return readBody(resp.Body, v)
}

func newRequest(method, url string, body []byte) (*http.Request, error) {
	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, errors.Wrap(err, "new request error")
	}

	//if AppSecret == "" {
	//	return nil, errors.Errorf("invalid app secret: %v", AppSecret)
	//}

	req.Header.Set("X-Octo-Key", AppSecret)
	req.Header.Set("X-Octo-Cli-Version", App.Version)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

type httpError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("httpError: %+v", *e)
}

func handleResponseError(resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	return &httpError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}
}

func readBody(r io.Reader, v interface{}) error {
	if v == nil {
		_, err := io.Copy(ioutil.Discard, r)
		return errors.Wrap(err, "discard error")
	}
	err := json.NewDecoder(r).Decode(v)
	return errors.Wrap(err, "decode error")
}

func Fatal(err error) {
	log.Println("%+v\n", err)
}

func NewTabwriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 5, 1, ' ', 0)
}
