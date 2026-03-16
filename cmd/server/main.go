package main

import (
	"context"
	"gin-quickstart/api"
	"gin-quickstart/internal/cache"
	"gin-quickstart/internal/config"
	"gin-quickstart/internal/database"
	"gin-quickstart/internal/middleware"
	"gin-quickstart/internal/queue"
	"gin-quickstart/internal/websocket"
	"gin-quickstart/worker"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title Movie Ticket Booking API
// @version 1.0
// @description This is a sample server for a movie ticket booking system.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {

	cfg := config.LoadConfig()

	// connect RabbitMQ
	rabbitMQURL := cfg.RabbitMQURI
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://admin:admin1234@localhost:5672/"
	}
	rabbitMQConnected := false
	if err := queue.InitRabbitMQ(rabbitMQURL); err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
	} else {
		rabbitMQConnected = true
		defer queue.CloseRabbitMQ()
	}

	// connect MongoDB
	mongoClient := database.InitMongoDB(cfg.MongoURI)
	db := mongoClient.Database(cfg.MongoDB)

	// connect Redis
	rdb := cache.InitRedis(cfg.RedisURI)

	// create indexes
	if err := database.CreateIndexes(db); err != nil {
		panic(err)
	}

	//  init firebase auth
	if err := middleware.InitFirebaseAuth(); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// === WebSocket Hub สำหรับ Audit Logs ===
	auditWsManager := websocket.NewAuditLogManager()
	go auditWsManager.Start()

	// === เริ่มการทำงานของ Workers สำหรับดึงข้อมูลจาก Queue ===
	if rabbitMQConnected {
		worker.StartNotificationWorker()
		worker.StartLogWorker(db, auditWsManager)
	} else {
		log.Println("⚠️ Skipping RabbitMQ Workers because connection failed.")
	}
	// ===============================================

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://yourfrontend.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api.Routes(r, db, rdb, auditWsManager)
	// check firebase middleware
	// protected := r.Group("/")
	// protected.Use(middleware.FirebaseAuthMiddleware())
	// protected.GET("/profile", func (c *gin.Context)  {
	//   email := c.GetString("email")
	//   c.JSON(http.StatusOK, gin.H{"message": "You are logged in!", "email": email})
	// })
	// r.GET("/ping", func(c *gin.Context) {
	//   c.JSON(http.StatusOK, gin.H{
	//     "message": "pong",
	//   })
	// })
	// r.GET("/health", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status": "ok",
	// 	})
	// })

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// เริ่มรัน Server ใน Goroutine เพื่อไม่ให้ Block การรอรับ Signal
	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // หลบการทำงานตรงนี้จนกว่าจะได้รับ Signal (เช่นกด Ctrl+C)
	log.Println("Shutting down server...")

	// ให้เวลา Server จัดการ Request ที่ค้างอยู่ 5 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exiting")
}
