package repository

import (
	"context"

	pb "github.com/Sula2007/ap2-generated/payment"
	"order-service/internal/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient interface {
	AuthorizePayment(req *domain.PaymentRequest) (*domain.PaymentResponse, error)
}

type paymentGRPCClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentGRPCClient(addr string) (PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &paymentGRPCClient{
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

func (c *paymentGRPCClient) AuthorizePayment(req *domain.PaymentRequest) (*domain.PaymentResponse, error) {
	resp, err := c.client.ProcessPayment(context.Background(), &pb.PaymentRequest{
		OrderId: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &domain.PaymentResponse{
		TransactionID: resp.TransactionId,
		Status:        resp.Status,
	}, nil
}
