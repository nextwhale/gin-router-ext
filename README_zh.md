[`英文`](README.md) | **中文** 

GinRouterExt是适用于Gin框架的库，对Gin的路由功能进行了扩展，支持可配置属性的路由定义。
### 特点
- 保留并扩展了原gin.Engine、Gin.RouterGroup的方法，对RouterGroup的路由方法重写以记录自定义的路由属性。
- 从中间件获取当前路由的配置信息，以及在server启动后导出全部路由配置信息。
- 增加基于Group路径的404页面方法。例如，综合性网站缺省404为html页面，而API模块的缺省404页则可能输出json格式。

### 安装

```
go get -u github.com/nextwhale/gin-router-ext@latest
```

## 基本使用介绍

### 创建gin.Engine对象
在gin框架中，使用`gin.New()`方法创建gin.Engine对象， 作为框架app实例；

使用本库时，则需要使用`ginrouterext.New()`方法创建`EngineExt`对象来代替，EngineExt是gin.Engine的扩展结构体.

```go
// 创建一个全新的实例：
var app = ginrouterext.New(nil)
// 或者，代入已有的gin.Engine实例：
var engExisted = &gin.New()
var app = ginrouterext.New(engExisted)
```

### 为路由设置自定义属性
在添加REST路由后，紧跟使用Set方法设置自定义属性.

例如：
```go
groupAPI := app.Group("/api")
groupAPI.GET("/contacts", handler).Set("Getting contacts", true, false, map[string]string{"name_fr":"Obtenir un contact","sitemap":"0"}) 
```

定义好后，您可以从中间件中获取当前匹配的路由配置，并作后续处理。


### 从中间件中获取当前路由配置

我们通常需要从中间件中验证登录状态、访问权限(ACL).  
通过`group.GetRouteSettings(ctx)`来获取当前匹配到的路由.  

举例：
```go
	groupAPI.Use(func(ctx *gin.Context) {
		rs := groupAPI.GetRouteSettings(ctx)
		if rs == nil {
			return
		}
		// 如果当前路由需要先登录才能访问，开始验证：
		if rs.RequiresAuth && ctx.Request.Header.Get("sso_token") == "" {
			ctx.JSON(401, gin.H{"code": "not_sign_in", "message":"You haven't sign in."})
			ctx.Abort()
			return
		}
		// 如果当前路由需要用户获得授权，开始验证
		roleID := ctx.Request.Header.Get("role_id")
		if rs.RequiresACL && !GoACL().IsRoleAllowedUniquely([]string{roleID}, ctx.FullPath()) {
			ctx.JSON(401, gin.H{"code": "API", "message": "You have no access to visit this path"})
			ctx.Abort()
			return
		}
		// 获取额外的自定义配置
		_ = rs.GetExtra("name_fr")
	})
```	

### 获取所有路由配置
我们可以通过调用`group.GetRoutesMap()`方法, 获取某个路由group记录的所有路由配置；
也可以调用`group.GetACLItems()`获取该group下的所有`requiresACL:true`路由作为ACL项。 

例如，在添加管理后台路由时，通常需要控制路由的访问权限（权限部分可以使用我的另一模块 [`go-acl`](https://github.com/nextwhale/go-acl/)）。 通过动态获取该group下的所有路由，列出到后台的权限管理页面即可， 新增的权限项无需手动添加。

例如：
```go
// get all routes
adminRoutes = groupAdmin.GetRoutesMap()
// TODO: save to database, and grant route access to admin
// get all ACL items of path group
adminACLItems := groupAdmin.GetACLItems()
```

## 为路由组自定义404页面
gin框架的gin.Engine.NoRoute()只支持设置一个统一的404处理方法。但本库扩展并支持了为路由group设置单独的404处理函数。  
注意您需要调用Abort相关的方法`ctx.Abort()`或`ctx.AbortWithStatus()`，来取消后续的处理函数冒泡。

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
如果使用过程中遇到问题，欢迎提交issues，并帮助我们改进。

## License
本包基于MIT License 分发, 请查看LISENCE文件获取详细说明。
