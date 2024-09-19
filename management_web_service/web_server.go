package management_web_service

import (
	"fmt"
	"github.com/dymensionxyz/roller/management_web_service/gin_wrapper"
	webtypes "github.com/dymensionxyz/roller/management_web_service/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	statikfs "github.com/rakyll/statik/fs"
	"html/template"
	"net/http"
)

func StartManagementWebService(cfg webtypes.Config) {
	binding := fmt.Sprintf("%s:%d", cfg.IP, cfg.Port)

	statikFS, err := statikfs.New()
	if err != nil {
		panic(errors.Wrap(err, "failed to create statik FS"))
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(gin_wrapper.KeyGinConfig, cfg)
	})
	/* How to implement middleware
	r.Use(func(c *gin.Context) {
		c.Next()
	})
	*/

	const (
		engineDelimsLeft  = "{[{"
		engineDelimsRight = "}]}"
	)
	r.Delims(engineDelimsLeft, engineDelimsRight)
	r.SetHTMLTemplate(
		template.Must(
			template.
				New("").
				Delims(engineDelimsLeft, engineDelimsRight).
				Funcs(nil).
				ParseFS(
					webtypes.WrapHttpFsToOsFs(statikFS),
					"/index.tmpl",
				),
		),
	)

	// Resources
	r.GET("/resources/*file", func(c *gin.Context) {
		http.FileServer(statikFS).ServeHTTP(c.Writer, c.Request)
	})

	// Web
	r.GET("/", HandleWebIndex)

	fmt.Println("INF: starting Web service at", binding)

	if err := r.Run(binding); err != nil {
		panic(errors.Wrap(err, "failed to start web service"))
	}
}

// wrap and return gin Context as a GinWrapper class with enhanced utilities
func wrapGin(c *gin.Context) gin_wrapper.GinWrapper {
	return gin_wrapper.WrapGin(c)
}
