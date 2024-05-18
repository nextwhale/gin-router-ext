// Copyright 2024 Shaotschaw Teng(github.com/nextwhale). All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ginrouterext

import (
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// define route methods
type HTTPMethod string

const HTTPMethodGET HTTPMethod = "GET"
const HTTPMethodPOST HTTPMethod = "POST"
const HTTPMethodPUT HTTPMethod = "PUT"
const HTTPMethodDELETE HTTPMethod = "DELETE"
const HTTPMethodHEAD HTTPMethod = "HEAD"
const HTTPMethodPATCH HTTPMethod = "PATCH"
const HTTPMethodOPTIONS HTTPMethod = "OPTIONS"

// /////////////////////
// Settings of single route
// /////////////////////
type RouteSetting struct {
	// Order NO. by it been addedf
	No int
	// route method
	Method HTTPMethod
	// route path
	Path string
	// Route name, typically been used to describe route
	Name string
	// Is route requires authentication?
	RequiresAuth bool
	// Is route requires access control?
	RequiresACL bool
	// any custom params, tipically struct data, used when route matched
	Extra map[string]string
}
func (ri *RouteSetting) GetExtra(key string) string{
	return ri.Extra[key]
}

// /////////////////////
// EngineExt inherits from gin.Engine
// /////////////////////
type EngineExt struct {
	*gin.Engine
	rge            *RouterGroupExt
	groupsCreated  map[string]*RouterGroupExt
	groupsNoRoute  map[string][]gin.HandlerFunc
	defaultNoRoute []gin.HandlerFunc
}

// Create a new group on root, or get the existed group which has same path
func (e *EngineExt) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	if relativePath == "" || relativePath == "/" {
		return e.rge
	}
	return e.rge.Group(relativePath, handlers...)
}

// To set group 404 handlers and default 404 handlers
// group 404 handlers only work when route matches group path prefix
func (e *EngineExt) NoRoute(handlers ...gin.HandlerFunc) {
	if len(handlers) > 0 {
		e.defaultNoRoute = handlers
	} else {
		handlers = e.defaultNoRoute
	}
	handlerList := make([]gin.HandlerFunc, 0, len(handlers)+len(e.groupsNoRoute))
	for groupPath, funcs := range e.groupsNoRoute {
		for _, fn := range funcs {
			handler := func(c *gin.Context) {
				if strings.HasPrefix(strings.ToLower(c.Request.URL.Path), groupPath) {
					fn(c)
				}
			}
			handlerList = append(handlerList, handler)
		}
	}
	handlerList = append(handlerList, handlers...)
	e.Engine.NoRoute(handlerList...)
}

// /////////////////////
// RouterGroupEx
// RouterGroupEx inherits gin.RouterGroup, it provides extra route setting manipulators as well
// /////////////////////
type RouterGroupExt struct {
	*gin.RouterGroup
	eng           *EngineExt
	routesIndexed map[string]*RouteSetting
	lastRouteInfo *RouteSetting
}

// retrieve to root group
func (r *RouterGroupExt) RootGroup() *RouterGroupExt {
	return r.eng.rge
}

// create a new group, or get the existed group which has same path
func (r *RouterGroupExt) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	fp := joinPaths(r.BasePath(), relativePath)
	if _, ok := r.eng.groupsCreated[fp]; !ok {
		r.eng.groupsCreated[fp] = &RouterGroupExt{
			RouterGroup:   r.RouterGroup.Group(relativePath, handlers...),
			eng:           r.eng,
			routesIndexed: make(map[string]*RouteSetting),
		}
	}
	return r.eng.groupsCreated[fp]
}

// register no route handlers for current group path
func (r *RouterGroupExt) NoRouteByGroup(handlers ...gin.HandlerFunc) {
	r.eng.groupsNoRoute[r.BasePath()] = handlers
	r.eng.NoRoute() // update root group no route handlers
}

func (r *RouterGroupExt) GET(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodGET, relativePath, handlers)
}
func (r *RouterGroupExt) POST(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodPOST, relativePath, handlers)
}
func (r *RouterGroupExt) PUT(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodPUT, relativePath, handlers)
}
func (r *RouterGroupExt) DELETE(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodDELETE, relativePath, handlers)
}
func (r *RouterGroupExt) HEAD(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodHEAD, relativePath, handlers)
}
func (r *RouterGroupExt) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) *RouterGroupExt {
	return r.addRoute(HTTPMethodOPTIONS, relativePath, handlers)
}

// add single route
func (r *RouterGroupExt) addRoute(method HTTPMethod, path string, handlers []gin.HandlerFunc) *RouterGroupExt {
	switch method {
	case HTTPMethodGET:
		r.RouterGroup.GET(path, handlers...)
	case HTTPMethodPOST:
		r.RouterGroup.POST(path, handlers...)
	case HTTPMethodPUT:
		r.RouterGroup.PUT(path, handlers...)
	case HTTPMethodDELETE:
		r.RouterGroup.DELETE(path, handlers...)
	case HTTPMethodHEAD:
		r.RouterGroup.HEAD(path, handlers...)
	case HTTPMethodPATCH:
		r.RouterGroup.PATCH(path, handlers...)
	case HTTPMethodOPTIONS:
		r.RouterGroup.OPTIONS(path, handlers...)
	default:
		panic("Route method is not supported:" + method)
	}

	// save route info
	rk := joinRouteIndex(string(method), r.BasePath(), path)
	r.routesIndexed[rk] = &RouteSetting{
		No:     len(r.routesIndexed) + 1,
		Method: method,
		Path:   path,
	}
	r.lastRouteInfo = r.routesIndexed[rk]

	return r
}

// set setting fields for last route
// Note： This method shhould be used after GET/POST/PUT/DELETE/HEAD/OPTIONS only
func (r *RouterGroupExt) Set(name string, requiresAuth bool, requiresACL bool, extra map[string]string) *RouterGroupExt {
	if r.lastRouteInfo == nil {
		return r
	}
	r.lastRouteInfo.Name = name
	r.lastRouteInfo.RequiresAuth = requiresAuth
	r.lastRouteInfo.RequiresACL = requiresACL
	r.lastRouteInfo.Extra = extra
	return r
}

// Get current requested RouteSetting .
// e.g: Using in middlewares for verifying
func (r *RouterGroupExt) GetRouteSettings(c *gin.Context) *RouteSetting {
	return r.GetRouteSettingsByPath(c.Request.Method, c.FullPath())
}

// Get single route info by method and path
func (r *RouterGroupExt) GetRouteSettingsByPath(method string, fullPath string) *RouteSetting {
	rk := joinRouteIndex(method, "", fullPath)
	ri, ok := r.routesIndexed[rk]
	if ok {
		return ri
	} else {
		return nil
	}
}

// Get the map of all routes added to this router instance
func (r *RouterGroupExt) GetRoutesMap() map[string]*RouteSetting {
	return r.routesIndexed
}

// Get access items
// Those routes that set requiresACL:true will be returned
func (r *RouterGroupExt) GetACLItems() map[string]*RouteSetting{
	al := map[string]*RouteSetting{}
	for ind, setting := range r.routesIndexed {
	    if setting.RequiresACL {
			al[ind] = setting
		}
	}
	return al
}

// New EngineExt instance with passing in gin.Engine or creating with gin.New() instance
func New(eng *gin.Engine) *EngineExt {
	if eng == nil {
		eng = gin.New()
	}
	e := &EngineExt{
		Engine: eng,
		rge: &RouterGroupExt{
			RouterGroup:   &eng.RouterGroup,
			routesIndexed: make(map[string]*RouteSetting),
		},
		groupsCreated:  make(map[string]*RouterGroupExt),
		groupsNoRoute:  make(map[string][]gin.HandlerFunc),
		defaultNoRoute: []gin.HandlerFunc{},
	}
	e.rge.eng = e
	return e
}

// helper function to join two path
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}
	finalPath := path.Join(absolutePath, relativePath)
	if relativePath[len(relativePath)-1] == '/' && finalPath[len(finalPath)-1] != '/' {
		return finalPath + "/"
	}
	return finalPath
}


// join a unique restful route index
func joinRouteIndex(method, basePath, relativePath string) string {
	fp := joinPaths(basePath, relativePath)
	return method + " " + fp
}
