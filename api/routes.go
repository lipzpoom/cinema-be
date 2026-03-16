package api

import (
	"gin-quickstart/internal/handlers"
	"gin-quickstart/internal/middleware"
	"gin-quickstart/internal/repositories"
	"gin-quickstart/internal/services"
	"gin-quickstart/internal/websocket"
	"gin-quickstart/pkg/redislock"
	"net/http"
	"time"

	_ "gin-quickstart/docs" // Swagger docs

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
)

func Routes(r *gin.Engine, db *mongo.Database, rdb *redis.Client, auditWsManager *websocket.AuditLogManager) {
	authHandler := &handlers.AuthHandler{DB: db}

	// === WebSocket Hub === //
	wsManager := websocket.NewClientManager()
	go wsManager.Start()

	// === Redis Locker === //
	locker := redislock.NewRedisLocker(rdb)

	// === Dependency Injection === //
	movieRepo := repositories.NewMovieRepository(db)
	movieSvc := services.NewMovieService(movieRepo)
	movieHandler := handlers.NewMovieHandler(movieSvc)

	showtimeRepo := repositories.NewShowtimeRepository(db)
	showtimeSvc := services.NewShowtimeService(showtimeRepo, movieRepo)
	showtimeHandler := handlers.NewShowtimeHandler(showtimeSvc)

	bookingRepo := repositories.NewBookingRepository(db)

	seatRepo := repositories.NewSeatRepository(db)
	seatSvc := services.NewSeatService(seatRepo, bookingRepo)
	seatHandler := handlers.NewSeatHandler(seatSvc)

	bookingSvc := services.NewBookingService(bookingRepo, seatRepo, locker)
	bookingHandler := handlers.NewBookingHandler(bookingSvc)

	userRepo := repositories.NewUserRepository(db)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	auditLogRepo := repositories.NewAuditLogRepository(db)
	auditLogSvc := services.NewAuditLogService(auditLogRepo)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogSvc)

	// public routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "system": "Booking Ticket System", "version": "1.0", "timeStamp": time.Now()})
	})

	// WebSocket endpoint for real-time seats
	r.GET("/ws/seats", websocket.ServeWS(wsManager))

	// WebSocket endpoint for real-time audit logs
	if auditWsManager != nil {
		r.GET("/ws/auditlogs", websocket.ServeAuditWS(auditWsManager))
	}

	apiGroup := r.Group("/api")

	// Movies
	movies := apiGroup.Group("/movies")
	{
		movies.GET("", movieHandler.GetAll)
		movies.GET("/:id", movieHandler.GetByID)
		movies.GET("/:id/showtimes", showtimeHandler.GetByMovieID)
		movies.POST("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), movieHandler.Create)
		movies.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), movieHandler.Update)
		movies.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), movieHandler.Delete)
	}

	// Showtimes
	showtimes := apiGroup.Group("/showtimes")
	{
		showtimes.GET("", showtimeHandler.GetAll)
		showtimes.GET("/now-showing", showtimeHandler.GetNowShowing)
		showtimes.GET("/:id", showtimeHandler.GetByID)
		showtimes.POST("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), showtimeHandler.Create)
		showtimes.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), showtimeHandler.Update)
		showtimes.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), showtimeHandler.Delete)
	}

	// Bookings
	bookings := apiGroup.Group("/bookings")
	{
		bookings.GET("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), bookingHandler.GetAll)
		bookings.GET("/:id", middleware.AuthMiddleware(), bookingHandler.GetByID)
		bookings.GET("/user/:user_id", middleware.AuthMiddleware(), bookingHandler.GetByUserID)
		bookings.POST("", middleware.AuthMiddleware(), bookingHandler.Create)
		bookings.POST("/:id/pay", middleware.AuthMiddleware(), bookingHandler.Pay)
		bookings.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), bookingHandler.Update)
		bookings.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), bookingHandler.Delete)
	}

	// Seats
	seats := apiGroup.Group("/seats")
	{
		seats.GET("", seatHandler.GetAll)
		seats.GET("/:id", seatHandler.GetByID)
		seats.POST("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), seatHandler.Create)
		seats.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), seatHandler.Update)
		seats.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), seatHandler.Delete)
	}

	// Users
	users := apiGroup.Group("/users")
	{
		users.GET("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), userHandler.GetAll)
		users.GET("/:id", middleware.AuthMiddleware(), userHandler.GetByID)
		users.POST("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), userHandler.Create)
		users.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), userHandler.Update)
		users.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), userHandler.Delete)
	}

	// Audit Logs
	auditlogs := apiGroup.Group("/auditlogs")
	{
		auditlogs.GET("", middleware.AuthMiddleware(), middleware.RequireRoles("ADMIN"), auditLogHandler.GetAllAuditLogs)
	}

	// protected routes
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/google/login", middleware.FirebaseAuthMiddleware(), authHandler.GoogleLogin)
		authGroup.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
	}

}
