package gin_mw

import (
	"errors"
	"github.com/gin-gonic/gin"
	"path"
)

type routerItf interface {
	BasePath() string
	Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
}

type Routes[T any] []func(a T) (m, p string, hf gin.HandlerFunc)

func (r Routes[T]) Set(a T, router routerItf) ([][2]string, error) {
	set := make(map[string]struct{})
	var endpoints [][2]string
	for _, route := range r {
		m, p, hf := route(a)
		key := m + router.BasePath() + p
		if _, ok := set[key]; ok {
			return nil, errors.New("duplicate route: " + m + " " + path.Join(router.BasePath(), p))
		}
		set[key] = struct{}{}
		router.Handle(m, p, hf)
		endpoints = append(endpoints, [2]string{m, path.Join(router.BasePath(), p)})
	}
	return endpoints, nil
}
