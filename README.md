**英文** | [`中文`](README_zh.md)

GinRouterExt is a library suitable for the Gin framework, which extends the Gin router, and supports configurable attributes wihle defining routes.

### Features
- Retained the origin methods of gin.Engine and gin.RouterGroup, rewrited the REST methods of RouterGroup to set custom attributes for each route.

- Supports retrieving the custom attributes of each route from middleware, and exporting those routes after defined.

- Customize different 404 pages for different group path.


### Installation

```
go get -u github.com/nextwheel/gin-router-ext@latest
```

## Basic usage

### Create gin Engine object
In the gin framework, we use ` gin.New() ` to create gin Engine object as app instance;

When using this package, we use `ginrouterext.New() ` to create an `EngineExt` as instance instead. `EngineExt` is an extended struct of gin.Engine.

```go
// create an new instance
var app = ginrouterext.New(nil)
// or create instance with existed gin.Engine object
var engExisted = &gin.New()
var app = ginrouterext.New(engExisted)
```

### Set custom attributes for route
Use method `group.Set()` to define custom attributes for each route following the REST method. 
> The params of group.Set():  
> group.Set(routeName, requiresAuth, requiresACL, extraMap) 

e.g:
```go
groupAPI := app.Group("/api")
groupAPI.GET("/contacts", handler).Set("Getting contacts", true, false, map[string]string{"name_fr":"Obtenir un contact","sitemap":"0"}) 
```

After that, we can retrieve the matched route setting in middlewares.

### Get the current route setting from middleware

Typically, we use middleware to verify login status and access permissions (ACLs) from middleware. By invoking `group.GetRouteSettings(ctx) ` in middleware can get the matched route setting.

e.g:
```go
	groupAPI.Use(func(ctx *gin.Context) {
		rs := groupAPI.GetRouteSettings(ctx)
		if rs == nil {
			return
		}
		// Verify if user is signed in if route demands it
		if rs.RequiresAuth && ctx.Request.Header.Get("sso_token") == "" {
			ctx.JSON(401, gin.H{"code": "not_sign_in", "message":"You haven't sign in."})
			ctx.Abort()
			return
		}
		// Verify if user is granted access to visit this path if route demands it
		roleID := ctx.Request.Header.Get("role_id")
		if rs.RequiresACL && !GoACL().IsRoleAllowedUniquely([]string{roleID}, ctx.FullPath()) {
			ctx.JSON(401, gin.H{"code": "API", "message": "You have no access to visit this path"})
			ctx.Abort()
			return
		}
		// To get extra attibute which was defined while adding this route
		_ = rs.GetExtra("name_fr")
	})
```	

### Get all routing configurations
We could retrieve all routing settings by calling `group.GetRoutesMap()`, and get those `requiresACL:true` routes as ACL by calling `group.GetACLItems()`.

For example, while developing a administration panel, some routes must be visited with access granted. We could define these setting while adding routes, then we can see these access items on administration panel.
(Using ACL, I recommend my another package [`go-acl`](https://github.com/nextwhale/go-acl/))

e.g:
```go
// get all routes
adminRoutes = groupAdmin.GetRoutesMap()
// TODO: save to database, and grant route access to admin
// get all ACL items of path group
adminACLItems := groupAdmin.GetACLItems()
```

## Customizing the 404 page for routing groups
In gin framework, only global 404 handlers can be added. However, with this package, we can add different 404 handler for each routing group, respectively.

Note that `ctx.Abort()` or `ctx.AbortWithStatus()` must be called in the group 404 handler.

```go
// default 404 handler
app.NoRoute(default404Handler)

// add a 404 handler for group /admin
groupAdmin.NoRouteByGroup(func(c *gin.Context) {
	c.JSON(404, gin.H{
		"code": "404_not_found",
		"msg": "Path is 404 not found under /admin",
	})
	c.Abort()
})

// Define another 404 for group /api
groupAPI.NoRouteByGroup(func(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(404, `<div>404 Not Found</div>`)
	c.Abort()
})
```

## Note
if you encountered any issue, just feel free to post it to look for help. 
And I strongly encourage contributing to this project.

## License
Distributed under MIT License, please see license file in code for more details.

