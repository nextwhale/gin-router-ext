package main

import (
	"fmt"
	. "ginrouterextexample/init"

	"github.com/gin-gonic/gin"
)

func handleAdmin(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"code": "API", "message": "hello /admin"})
}
func handleWeb(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(200, `<div>A sample page</div>`)
}

func init() {
	app := App()
	app.NoRoute(func(ctx *gin.Context) {
		ctx.HTML(200, "404.html", gin.H{"title": "404 Not Found"})
	})
	// add a new path group
	groupAdmin := app.Group("/admin")
	{
		// Get route setting
		groupAdmin.Use(func(ctx *gin.Context) {
			ssoToken := ctx.Request.Header.Get("user_id")
			roleID := ctx.Request.Header.Get("role_id")
			rs := groupAdmin.GetRouteSettings(ctx)

			fmt.Print("rs : ", rs)
			if rs == nil {
				return
			}
			// 如果当前路由需要先登录才能访问，开始验证：
			if rs.RequiresAuth && ssoToken == "" {
				ctx.JSON(401, gin.H{"code": "not_sign_in", "message": "You haven't sign in."})
				ctx.Abort()
				return
			}
			if rs.RequiresACL && !ACL().IsRoleAllowedUniquely([]string{roleID}, ctx.FullPath()) {
				ctx.JSON(401, gin.H{"code": "API", "message": "You have no access to visit this path"})
				ctx.Abort()
				return
			}
		})

		groupAdmin.GET("/hello", handleAdmin).Set("Saying Hello", false, false, nil)
		groupAdmin.GET("/contacts", handleAdmin).Set("Getting contacts", true, true, nil)
		groupAdmin.GET("/article/list", handleAdmin).Set("Article list", true, true, map[string]string{"showInSitemap": "1", "name_en": "Article List"})
		groupAdmin.PUT("/article/edit/:id", handleAdmin).Set("Article editting", true, true, map[string]string{"showInSitemap": "1", "name_ja": "ビデオモデレーター"})
		groupAdmin.DELETE("/article/del/:id", handleAdmin).Set("Article deletting", true, true, map[string]string{"showInSitemap": "0", "logPrint": "1"})
		groupAdmin.GET("/video/list", handleAdmin).Set("Video list", true, true, map[string]string{"showInSitemap": "1", "logPrint": "0"})
		groupAdmin.PUT("/video/edit/:id", handleAdmin).Set("Video editting", true, true, map[string]string{"showInSitemap": "0", "logPrint": "1"})
		groupAdmin.DELETE("/video/del/:id", handleAdmin).Set("Video deletting", true, true, map[string]string{"showInSitemap": "1", "logPrint": "1"})


		// add a 404 handler for group /admin
		groupAdmin.NoRouteByGroup(func(c *gin.Context) {
			c.JSON(404, gin.H{
				"code": "404_not_found",
				"msg":  "Path is 404 not found under /admin",
			})
			c.Abort()
		})
	}

	gWeb := app.Group("/web")
	{
		gWeb.GET("/about", handleWeb).Set("About us", false, false, map[string]string{"sitemap": "1"})
		gWeb.PUT("/user/account", handleWeb).Set("Modify account", true, false, nil)

		// customize a 404 handler for this group
		gWeb.NoRouteByGroup(func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"code": "not found", "message": "This API was not found"})
			ctx.Abort()
		})
	}

}

func main() {
	app := App()

	groupAdmin := app.Group("/admin")
	fmt.Println("ACL items: ", groupAdmin.GetACLItems())

	if nil != app.Run(":8000") {
		panic("Gin server: Running error")
	}
}


// To test this example, use curl with:
// Method not found:            curl -X GET http://127.0.0.1:8000/admin/article/del/18
// Not signed in:               curl -X DELETE http://127.0.0.1:8000/admin/article/del/18
// Signed in & Have no access:  curl -X DELETE -H 'user_id: 10' http://127.0.0.1:8000/admin/article/del/18
// Signed in & Have access:     curl -X DELETE -H 'user_id: 10' -H 'role_id: editor' http://127.0.0.1:8000/admin/article/del/18
