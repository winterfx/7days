package test

import (
	"gee"
	"net/http"
	"testing"
)

func TestRouter(t *testing.T) {
	engine := gee.New()

	engine.GET("/", func(c *gee.Context) {
		c.JSON(200, "success")
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

}
func TestGroup(t *testing.T) {
	engine := gee.New()
	g1 := engine.Group("/book")
	{
		g1.GET("/english", func(c *gee.Context) {
			c.JSON(http.StatusOK, map[string]interface{}{
				"book category": "english",
			})
		})
		g1.GET("/math", func(c *gee.Context) {
			c.JSON(http.StatusOK, map[string]interface{}{
				"book category": ",math",
			})
		})

	}
	g2 := engine.Group("/movie")
	{
		g2.GET("/action", func(c *gee.Context) {
			c.String(http.StatusOK, "action movies")
		})
	}
}
func TestRecovery(t *testing.T) {
	engine := gee.New()
	engine.Use(gee.Recovery())
	engine.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})
	if err := engine.Run("127.0.0.1:8080"); err != nil {
		panic(err.Error())
	}
}
