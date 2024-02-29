package routes

import (
	"social-network-server/api/handlers"
	"social-network-server/api/middleware"
	"social-network-server/api/views"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.RouterGroup) {
	r.GET("/signup", views.SignupView)
	r.GET("/login", views.LoginView)
	r.POST("/signup", handlers.Signup)
	r.POST("/validate-email", handlers.ExistEmail)
	r.POST("/validate-username", handlers.ExistUsername)
	r.POST("/login", handlers.UserLogin)

	r.Use(middleware.AuthMiddleware())

	r.GET("/home", views.HomeView)
	r.POST("/create-post", handlers.CreateNewPost)
	r.POST("/follow", handlers.Follow)
	r.POST("/unfollow", handlers.Unfollow)
	r.POST("/feed", handlers.Feed)
	r.POST("/loggout", handlers.Logout)
	r.GET("/:username", handlers.RenderProfileTemplate)
	r.POST("/profile", handlers.Profile)
	r.POST("/profile/:username", handlers.AnotherUserProfile)
	r.GET("/edit-profile", views.EditProfileView)
	r.POST("/edit-profile", handlers.EditProfile)
	r.POST("create-message/:username", handlers.CreateNewMessage)
	r.POST("/chat/:username", handlers.Chat)
	r.GET("/chat/:username", views.ChatView)
	r.GET("/websocket/:username", handlers.WebSocketHandler)
}
