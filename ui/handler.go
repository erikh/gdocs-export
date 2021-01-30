package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/erikh/gdocs-export/pkg/converters"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"github.com/erikh/gdocs-export/pkg/oauth2"
	"github.com/erikh/gdocs-export/pkg/util"
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
	params, err := c.FormParams()
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	url := params.Get("url")
	format := params.Get("format")
	preview := params.Get("preview") == "on"
	download := strings.HasPrefix(params.Get("action"), "Download")

	client := oauth2.GetClient()

	if download {
		tar, err := util.MakeTarFromGDoc(client, url, format)
		if err != nil {
			c.Logger().Error(err)
			return err
		}
		defer tar.Close()

		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", "output.tar.gz"))
		return c.Stream(http.StatusOK, "application/gzip", tar)
	}

	docID, err := util.ParseDocsURL(url)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	doc, err := util.DownloadDoc(client, docID)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	res, err := converters.Convert(format, doc, downloader.Manifest{})
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	var ct string
	switch format {
	case "html":
		ct = "text/html"
	case "md":
		ct = "text/plain"
	default:
		c.Logger().Error("invalid format")
		return fmt.Errorf("invalid format")
	}

	if !preview {
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", "output."+format))
	}

	return c.Blob(http.StatusOK, ct, []byte(res))
}
