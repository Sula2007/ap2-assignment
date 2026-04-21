package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pborder "github.com/Sula2007/ap2-generated/order"

	"order-service/internal/repository"
	grpcserver "order-service/internal/transport/grpc"
	"order-service/internal/transport/handler"
	"order-service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading from environment")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to order database")

	paymentAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	paymentClient, err := repository.NewPaymentGRPCClient(paymentAddr)
	if err != nil {
		log.Fatalf("Failed to connect to payment gRPC server at %s: %v", paymentAddr, err)
	}
	log.Printf("Connected to payment gRPC server at %s", paymentAddr)

	orderRepo := repository.NewOrderRepository(db)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, paymentClient)
	orderHandler := handler.NewOrderHandler(orderUsecase)

	go func() {
		grpcPort := os.Getenv("ORDER_GRPC_PORT")
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on :%s : %v", grpcPort, err)
		}
		s := grpc.NewServer()
		pborder.RegisterOrderServiceServer(s, grpcserver.NewOrderGRPCServer(db))
		log.Printf("Order gRPC streaming server started on :%s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)
	}

	httpPort := os.Getenv("HTTP_PORT")
	log.Printf("Order HTTP server started on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, router); err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}