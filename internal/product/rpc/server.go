package rpc

import (
	"context"
	"database/sql"
	"net"

	pb "github.com/nilesh/shopcore-contracts/shopcorepb/shopcorepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedProductServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) GetProductDetails(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	var id, categoryID, category string
	var price int64
	err := s.db.QueryRowContext(ctx, `SELECT p.id,p.price,COALESCE(p.category_id::text,''),COALESCE(c.name,'') FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.id=$1 AND p.active=true`, req.GetProductId()).Scan(&id, &price, &categoryID, &category)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "could not load product")
	}
	return &pb.ProductResponse{ProductId: id, Price: int32(price), VendorId: categoryID, Categroy: category}, nil
}

func Listen(address string, db *sql.DB) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterProductServiceServer(server, NewServer(db))
	return server.Serve(listener)
}
