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

	pb "github.com/Sula2007/ap2-generated/payment"

	"payment-service/internal/repository"
	grpcserver "payment-service/internal/transport/grpc"
	httphandler "payment-service/internal/transport/http/handler"
	"payment-service/internal/usecase"
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
	log.Println("Connected to payment database")

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)

	go func() {
		grpcPort := os.Getenv("GRPC_PORT")
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on :%s : %v", grpcPort, err)
		}
		s := grpc.NewServer(
			grpc.UnaryInterceptor(grpcserver.LoggingInterceptor),
		)
		pb.RegisterPaymentServiceServer(s, grpcserver.NewPaymentGRPCServer(paymentUsecase))
		log.Printf("Payment gRPC server started on :%s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	paymentHandler := httphandler.NewPaymentHandler(paymentUsecase)

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.POST("/payments", paymentHandler.CreatePayment)
		api.GET("/payments/:order_id", paymentHandler.GetPayment)
	}

	httpPort := os.Getenv("HTTP_PORT")
	log.Printf("Payment HTTP server started on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, router); err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}