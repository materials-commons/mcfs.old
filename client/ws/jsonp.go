package ws

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"net/http"
)

type jsonpResponseWriter struct {
	writer   http.ResponseWriter
	callback string
}

func (j *jsonpResponseWriter) Header() http.Header {
	return j.writer.Header()
}

func (j *jsonpResponseWriter) WriteHeader(status int) {
	j.writer.WriteHeader(status)
}

func (j *jsonpResponseWriter) Write(bytes []byte) (int, error) {
	if j.callback != "" {
		bytes = []byte(fmt.Sprintf("%s(%s)", j.callback, bytes))
	}
	return j.writer.Write(bytes)
}

func newJSONPResponseWriter(httpWriter http.ResponseWriter, callback string) *jsonpResponseWriter {
	jsonpResponseWriter := new(jsonpResponseWriter)
	jsonpResponseWriter.writer = httpWriter
	jsonpResponseWriter.callback = callback
	return jsonpResponseWriter
}

// JSONPFilter implements JSONP handling. It looks for a callback argument and modifies the
// returned response to wrap it in the callback.
func JSONPFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	callback := req.Request.FormValue("callback")
	jsonpResponseWriter := newJSONPResponseWriter(resp.ResponseWriter, callback)
	resp.ResponseWriter = jsonpResponseWriter
	chain.ProcessFilter(req, resp)
}
