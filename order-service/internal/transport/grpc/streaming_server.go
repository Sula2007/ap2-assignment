package grpc

import (
	"database/sql"
	"log"
	"time"

	pb "github.com/Sula2007/ap2-generated/order"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	db *sql.DB
}

func NewOrderGRPCServer(db *sql.DB) *OrderGRPCServer {
	return &OrderGRPCServer{db: db}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(
	req *pb.OrderRequest,
	stream pb.OrderService_SubscribeToOrderUpdatesServer,
) error {
	log.Printf("[Stream] Client subscribed to order: %s", req.OrderId)

	lastStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("[Stream] Client disconnected from order: %s", req.OrderId)
			return nil
		default:
		}

		var currentStatus string
		err := s.db.QueryRowContext(
			stream.Context(),
			"SELECT status FROM orders WHERE id = $1",
			req.OrderId,
		).Scan(&currentStatus)

		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "order %s not found", req.OrderId)
		}
		if err != nil {
			return status.Errorf(codes.Internal, "db error: %v", err)
		}

		if currentStatus != lastStatus {
			lastStatus = currentStatus
			if err := stream.Send(&pb.OrderStatusUpdate{
				OrderId:   req.OrderId,
				Status:    currentStatus,
				UpdatedAt: timestamppb.New(time.Now()),
			}); err != nil {
				return err
			}
			log.Printf("[Stream] Sent update: order=%s status=%s", req.OrderId, currentStatus)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
