package main

import (
	"net/http"

	"github.com/amtoaer/go-7-days/gee"
)

func main() {
	client := gee.New()
	client.Get("/get", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"username": "amtoaer",
			"password": "test",
		})
	})
	client.Post("/post", func(c *gee.Context) {
		value := c.PostForm("name")
		c.String(http.StatusOK, "hello,%s!", value)
	})
	client.Run(":9999")
}
