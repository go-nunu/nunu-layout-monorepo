package server

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	nethttp "net/http"
	"nunu-layout-monorepo/app/admin/docs"
	"nunu-layout-monorepo/app/admin/internal/handler"
	"nunu-layout-monorepo/app/admin/internal/middleware"
	"nunu-layout-monorepo/app/admin/web"
	"nunu-layout-monorepo/pkg/jwt"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/server/http"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	e *casbin.SyncedEnforcer,
	adminHandler *handler.AdminHandler,
	userHandler *handler.UserHandler,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)
	// 设置前端静态资源
	s.Use(static.Serve("/", static.EmbedFolder(web.Assets(), "dist")))
	s.NoRoute(func(c *gin.Context) {
		indexPageData, err := web.Assets().ReadFile("dist/index.html")
		if err != nil {
			c.String(nethttp.StatusNotFound, "404 page not found")
			return
		}
		c.Data(nethttp.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
	// swagger doc
	docs.SwaggerInfo.BasePath = "/"
	s.GET("/swagger/*any", newSwaggerHandler())

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/login", adminHandler.Login)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger), middleware.AuthMiddleware(e))
		{
			strictAuthRouter.GET("/users", userHandler.GetUsers)

			strictAuthRouter.GET("/menus", adminHandler.GetMenus)
			strictAuthRouter.GET("/admin/menus", adminHandler.GetAdminMenus)
			strictAuthRouter.POST("/admin/menu", adminHandler.MenuCreate)
			strictAuthRouter.PUT("/admin/menu", adminHandler.MenuUpdate)
			strictAuthRouter.DELETE("/admin/menu", adminHandler.MenuDelete)

			strictAuthRouter.GET("/admin/users", adminHandler.GetAdminUsers)
			strictAuthRouter.GET("/admin/user", adminHandler.GetAdminUser)
			strictAuthRouter.PUT("/admin/user", adminHandler.AdminUserUpdate)
			strictAuthRouter.POST("/admin/user", adminHandler.AdminUserCreate)
			strictAuthRouter.DELETE("/admin/user", adminHandler.AdminUserDelete)
			strictAuthRouter.GET("/admin/user/permissions", adminHandler.GetUserPermissions)
			strictAuthRouter.GET("/admin/role/permissions", adminHandler.GetRolePermissions)
			strictAuthRouter.PUT("/admin/role/permissions", adminHandler.UpdateRolePermission)
			strictAuthRouter.GET("/admin/roles", adminHandler.GetRoles)
			strictAuthRouter.POST("/admin/role", adminHandler.RoleCreate)
			strictAuthRouter.PUT("/admin/role", adminHandler.RoleUpdate)
			strictAuthRouter.DELETE("/admin/role", adminHandler.RoleDelete)

			strictAuthRouter.GET("/admin/apis", adminHandler.GetApis)
			strictAuthRouter.POST("/admin/api", adminHandler.ApiCreate)
			strictAuthRouter.PUT("/admin/api", adminHandler.ApiUpdate)
			strictAuthRouter.DELETE("/admin/api", adminHandler.ApiDelete)

		}
	}
	return s
}

func newSwaggerHandler() gin.HandlerFunc {
	swaggerHandler := ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	)
	return func(ctx *gin.Context) {
		if ctx.Param("any") == "/doc.json" {
			ctx.Data(nethttp.StatusOK, "application/json; charset=utf-8", docs.SwaggerJSON)
			return
		}
		swaggerHandler(ctx)
	}
}
