package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (c *Context) PostForm(key string) string {
	return c.request.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

func (c *Context) Status(code int) {
	c.writer.WriteHeader(code)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Context-Type", "text/plain")
	c.Status(code)
	c.writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, content H) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.writer)
	if err := encoder.Encode(content); err != nil {
		http.Error(c.writer, err.Error(), code)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.writer.Write([]byte(html))
}

func (c *Context) notFound() {
	c.Status(http.StatusNotFound)
	c.writer.Write([]byte("404 NOT FOUND"))
}
