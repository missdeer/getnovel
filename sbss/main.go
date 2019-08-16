package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func getBookSource(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusOK, gin.H{
			"result":  "error",
			"message": "host name required",
		})
		return
	}

	// read from database
}

func updateBookSource(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result":  "error",
			"message": fmt.Sprintf("%s", err.Error()),
		})
		return
	}

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result":  "error",
			"message": fmt.Sprintf("%s", err.Error()),
		})
		return
	}

	// save to database

	c.JSON(http.StatusOK, gin.H{
		"result": "OK",
	})
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home.tmpl", gin.H{
		"title": "简易书源服务(Simple Book Source Service)",
	})
}

func main() {
	addr := ":8091"
	if bind, ok := os.LookupEnv("BIND"); ok {
		addr = bind
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", homePage)
	r.GET("/bs/:host", getBookSource)
	r.POST("/update", updateBookSource)

	log.Fatal(r.Run(addr))
}
