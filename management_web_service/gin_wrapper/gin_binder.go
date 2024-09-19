package gin_wrapper

import "github.com/gin-gonic/gin"

type GinBinder struct {
	c   *gin.Context
	err error
}

func (b *GinBinder) Json(pointer any) *GinBinder {
	if b.err == nil {
		if err := b.c.ShouldBindJSON(pointer); err != nil {
			b.err = err
		}
	}
	return b
}

func (b *GinBinder) Uri(pointer any) *GinBinder {
	if b.err == nil {
		if err := b.c.ShouldBindUri(pointer); err != nil {
			b.err = err
		}
	}
	return b
}

func (b *GinBinder) Bind(pointer any) *GinBinder {
	if b.err == nil {
		if err := b.c.ShouldBind(pointer); err != nil {
			b.err = err
		}
	}
	return b
}

func (b *GinBinder) Query(pointer any) *GinBinder {
	if b.err == nil {
		if err := b.c.ShouldBindQuery(pointer); err != nil {
			b.err = err
		}
	}
	return b
}

func (b *GinBinder) Header(pointer any) *GinBinder {
	if b.err == nil {
		if err := b.c.ShouldBindHeader(pointer); err != nil {
			b.err = err
		}
	}
	return b
}

func (b *GinBinder) Error() error {
	return b.err
}
