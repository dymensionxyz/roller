package gin_wrapper

import (
	webtypes "github.com/dymensionxyz/roller/management_web_service/types"
	"github.com/gin-gonic/gin"
)

type ginWrapperType int8

const (
	GwtDefault ginWrapperType = iota
)

type GinWrapper struct {
	c     *gin.Context
	wType ginWrapperType
}

func WrapGin(c *gin.Context) GinWrapper {
	return GinWrapper{
		c:     c,
		wType: GwtDefault,
	}
}

func (w GinWrapper) Gin() *gin.Context {
	return w.c
}

func (w GinWrapper) Binder() *GinBinder {
	return &GinBinder{
		c:   w.c,
		err: nil,
	}
}

func (w GinWrapper) Config() webtypes.Config {
	return w.c.MustGet(KeyGinConfig).(webtypes.Config)
}
