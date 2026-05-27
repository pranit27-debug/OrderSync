package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/varun-karmikanda/trachios/internal/api/handler"
	apimiddleware "github.com/varun-karmikanda/trachios/internal/api/middleware"
	"github.com/varun-karmikanda/trachios/internal/config"
	"github.com/varun-karmikanda/trachios/internal/repository"
	"github.com/varun-karmikanda/trachios/internal/service"
)

func main(){

	cfg, err := config.LoadConfig("./configs")
	if(err != nil){
		log.Fatalf("Could not load configuration: %v", err)
	}

	log.Printf("Trachios starting at port %d...", cfg.Server.Port)
	log.Printf("Postgres Host: %s", cfg.Postgres.Host)
	log.Printf("Redis Host: %s", cfg.Redis.Host)

	// POSTGRESQL CONNECTION
	db, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}
	defer db.Close()
	log.Println("Succcessfully connected to PostgresSQL!")
	
	// REDIS CONNECTION
	rdb, err := repository.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}
	defer rdb.Close()
	log.Println("Succcessfully connected to Redis!")
	
	// 
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	userRepo.SeedAdminUser(cfg)

	orderService := service.NewOrderService(orderRepo, rdb)

	userHandler := handler.NewUserHandler(userRepo, cfg)
	orderHandler := handler.NewOrderHandler(orderService, orderRepo)

	router.Route("/api/v1", func (r chi.Router)  {

		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
		r.Get("/health", func (w http.ResponseWriter, r *http.Request)  {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.AuthMiddleware(cfg.JWT.Secret))

			r.Get("/user", userHandler.GetCurrentUser)

			r.Post("/orders", orderHandler.CreateOrder)
			r.Get("/my-orders", orderHandler.GetMyOrders)
			r.Put("/orders/{orderID}/cancel", orderHandler.CancelOrder)

			r.Get("/orders/{orderID}", orderHandler.GetOrderByID)

			r.Group(func(r chi.Router) {
				r.Use(apimiddleware.AdminOnly)

				r.Get("/orders", orderHandler.GetAllOrders)

				r.Put("/orders/{orderID}/status", orderHandler.AdminUpdateOrderStatus)
			})
		})

	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running on http://localhost:%d", cfg.Server.Port)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}