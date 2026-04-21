package grpc

import (
	"context"
	"time"

	pb "github.com/Sula2007/ap2-generated/payment"
	"payment-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	usecase usecase.PaymentUsecase
}

func NewPaymentGRPCServer(uc usecase.PaymentUsecase) *PaymentGRPCServer {
	return &PaymentGRPCServer{usecase: uc}
}

func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := s.usecase.AuthorizePayment(req.OrderId, req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process payment: %v", err)
	}

	return &pb.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
		ProcessedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (s *PaymentGRPCServer) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payments, err := s.usecase.ListPayments(req.MinAmount, req.MaxAmount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var result []*pb.PaymentResponse
	for _, p := range payments {
		result = append(result, &pb.PaymentResponse{
			TransactionId: p.TransactionID,
			Status:        p.Status,
			ProcessedAt:   timestamppb.New(p.CreatedAt),
		})
	}

	return &pb.ListPaymentsResponse{Payments: result}, nil
}
