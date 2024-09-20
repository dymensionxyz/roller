package management_web_service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleWebIndex(c *gin.Context) {
	w := wrapGin(c)
	cfg := w.Config()

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"host": c.Request.Host,
		"ip":   cfg.IP,
		"port": cfg.Port,
	})
}
