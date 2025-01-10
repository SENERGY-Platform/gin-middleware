package gin_mw

import (
	"errors"
	"github.com/gin-gonic/gin"
	"path"
)

type loggerItf interface {
	Debug(v ...any)
}

type routerItf interface {
	BasePath() string
	Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
}

type Routes[T any] []func(a T) (m, p string, hf gin.HandlerFunc)

func (r Routes[T]) Set(a T, router routerItf, logger loggerItf) error {
	set := make(map[string]struct{})
	for _, route := range r {
		m, p, hf := route(a)
		key := m + router.BasePath() + p
		if _, ok := set[key]; ok {
			return errors.New("duplicate route: " + m + " " + path.Join(router.BasePath(), p))
		}
		set[key] = struct{}{}
		router.Handle(m, p, hf)
		if logger != nil {
			logger.Debug("set route: " + m + " " + path.Join(router.BasePath(), p))
		}
	}
	return nil
}

func (r *Routes[T]) Append(routes Routes[T]) {
	*r = append(*r, routes...)
}
