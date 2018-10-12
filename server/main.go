package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"ill.fi/superpower/server/api"
	"ill.fi/superpower/server/web"
	"net/http"
	"os"
	"runtime"
)

func clearLockFiles() {
	dbs := []string{"files", "users", "invites"}
	for _, f := range dbs {
		e := os.Remove(fmt.Sprintf("db/def/%s/LOCK", f))
		if e != nil {
			fmt.Println(e.Error())
		}
	}
}

func AdminRequired(db *api.DBHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sesh := sessions.Default(ctx)
		user := sesh.Get("user")
		if user == nil {
			ctx.JSON(http.StatusBadRequest, api.Response{
				Error:    true,
				Response: "Invalid session token",
			})
		} else {
			acc := api.GetAccountByID(user.(int), db.Users)
			if acc.Administrator {
				ctx.Next()
			} else {
				ctx.JSON(http.StatusBadRequest, api.Response{
					Error:    true,
					Response: "Not an administrator",
				})
			}
		}
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user")
		if user == nil {
			ctx.JSON(http.StatusBadRequest, api.Response{
				Error:    true,
				Response: "Invalid session token",
			})
		} else {
			ctx.Next()
		}
	}
}

func main() {

	runtime.GOMAXPROCS(128)
	clearLockFiles()

	gin.DisableConsoleColor()

	dblu := flag.String("db-u-path", "db/def/users", "the file location for the user database")
	dblf := flag.String("db-f-path", "db/def/files", "the file location for the file database")
	dbli := flag.String("db-i-path", "db/def/invites", "the file location for the invite database")
	flag.Parse()

	db := api.CreateDBHandler(*dblu, *dblf, *dbli)
	defer db.Close()
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./views", true)))
	r.Use(gin.Recovery())

	r.Use(sessions.Sessions("ill-session", sessions.NewCookieStore([]byte("penis"))))
	SetupAPIRoutes(r, db)
	r.Run()
}

func WithDB(f func(*gin.Context, *api.DBHandler), db *api.DBHandler) func(*gin.Context) {
	return func(ctx *gin.Context) {
		f(ctx, db)
	}
}

func WithDBMiddleware(f func(*api.DBHandler) gin.HandlerFunc, db *api.DBHandler) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return f(db)
	}
}

func SetupAPIRoutes(eng *gin.Engine, db *api.DBHandler) {
	apiGroup := eng.Group("/api/user")
	{
		apiGroup.GET("/list", WithDB(web.ListUsersHandler, db))
		apiGroup.GET("/info/:id", WithDB(web.UserInfoHandler, db))
		apiGroup.POST("/create", WithDB(web.UserCreationHandler, db))
	}
	auth := eng.Group("/api/auth")
	{
		auth.POST("/login", WithDB(web.LoginHandler, db))
		auth.GET("/logout", WithDB(web.LogoutHandler, db))
		auth.GET("/check", WithDB(web.AuthenticationCheckHandler, db))
	}
	admin := eng.Group("/admin")
	admin.Use(WithDBMiddleware(AdminRequired, db)())
	{
		admin.DELETE("/delete")
	}
}

func SetupPrivateRoutes(eng *gin.Engine, db *api.DBHandler) {
	priv := eng.Group("/home")
	{
		// delete file
		// upload file
	}
	priv.Use(AuthRequired())
}

/*
func SetupViewRoutes(eng *gin.Engine) {
	view := eng.Group("/")
	{
		view.GET("/", web.IndexHandler)
		view.GET("/faq", web.FAQHandler)
	}
}
*/
