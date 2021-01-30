package main

import (
	"net/http"

	_ "github.com/erikh/gdocs-export/ui/fs"
	"github.com/labstack/echo/v4"
	"github.com/rakyll/statik/fs"
)

func main() {
	e := echo.New()
	e.GET("/", serveIndex)
	e.POST("/", convertURL)
	e.Start("localhost:4000")
}

func serveIndex(c echo.Context) error {
	fs, err := fs.New()
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	file, err := fs.Open("/index.html")
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	defer file.Close()

	return c.Stream(http.StatusOK, "text/html", file)
}

func convertURL(c echo.Context) error {
	return nil
}
