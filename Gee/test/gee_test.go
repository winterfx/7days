package test

import (
	"gee"
	"testing"
)

func TestRun(t *testing.T) {
	engine := gee.New()
	engine.GET("/", func(c *gee.Context) {
		c.JSON(200, map[string]interface{}{
			"name": "winter",
			"sex":  "male",
		})
	})
	engine.GET("/book", func(c *gee.Context) {
		c.JSON(200, map[string]interface{}{
			"book": "all",
		})
	})
	engine.GET("/book/english", func(c *gee.Context) {
		c.JSON(200, map[string]interface{}{
			"book": "english",
		})
	})
	engine.GET("/book/math", func(c *gee.Context) {
		c.JSON(200, map[string]interface{}{
			"book": "math",
		})
	})
	engine.GET("/book/math", func(c *gee.Context) {
		c.JSON(200, map[string]interface{}{
			"book": "math2",
		})
	})
	if err := engine.Run("127.0.0.1:8080"); err != nil {
		panic(err.Error())
	}

}
