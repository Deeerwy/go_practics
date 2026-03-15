package grpcserver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"tech-ip-sem2/proto/authpb"
)

type AuthService interface {
	Verify(token string) (bool, string)
}

type Server struct {
	authpb.UnimplementedAuthServiceServer
	svc AuthService
}

func NewServer(svc AuthService) *Server {
	return &Server{svc: svc}
}

// Register регистрирует сервер в gRPC-сервере.
func Register(s *grpc.Server, impl *Server) {
	authpb.RegisterAuthServiceServer(s, impl)
}

func (s *Server) Verify(ctx context.Context, req *authpb.VerifyRequest) (*authpb.VerifyResponse, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	valid, subject := s.svc.Verify(req.Token)
	if !valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &authpb.VerifyResponse{
		Valid:   true,
		Subject: subject,
	}, nil
}


