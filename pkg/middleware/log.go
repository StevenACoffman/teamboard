package middleware

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/StevenACoffman/teamboard/pkg/middleware/http2curl"
)

type LoggingRoundTripper struct {
	next   http.RoundTripper
	logger io.Writer
}

func NewLoggingRoundTripper(
	next http.RoundTripper,
	w io.Writer,
) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		next:   next,
		logger: w,
	}
}

func (rt *LoggingRoundTripper) RoundTrip(
	req *http.Request,
) (resp *http.Response, err error) {
	defer func(begin time.Time) {
		var msg string
		body, getResponseBodyErr := GetResponseBody(resp)
		if getResponseBodyErr != nil {
			fmt.Println("unable to get response Body",getResponseBodyErr)
		}
		gotHTTPErr := resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300)
		gotGraphQLErr := false
		if body != "" && strings.Contains(body, "\"errors\":[{\"m") {
			gotGraphQLErr = true
		}
		if gotHTTPErr || gotGraphQLErr{ // only log when there was a problem

			msg = fmt.Sprintf(
				"method=%s host=%s path=%s status_code=%d took=%s\n",
				req.Method,
				req.URL.Host,
				req.URL.Path,
				resp.StatusCode,
				time.Since(begin),
			)
			if err != nil {
				fmt.Fprintf(rt.logger, "%s : %+v\n", msg, err)
			} else {
				fmt.Fprintf(rt.logger, "%s\n", msg)
			}
			command, _ := http2curl.GetCurlCommand(req)
			fmt.Println(command)

			fmt.Println(body)
		}
	}(time.Now())

	return rt.next.RoundTrip(req)
}

// GetResponseBody will read the response body without clobbering it
// so it can be re-read elsewhere
func GetResponseBody(r *http.Response) (string,error) {
	if r.Body == nil {
		return "", nil
	}
	body, err:= ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body), nil
}