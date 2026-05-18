package grpc

import (
	"context"
	"strings"

	"tech-ip-sem2/services/auth/pkg/authpb"

	"go.uber.org/zap"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	log *zap.Logger
}

func NewAuthServer(log *zap.Logger) *AuthServer {
	return &AuthServer{
		log: log,
	}
}

func (s *AuthServer) Verify(ctx context.Context, req *authpb.VerifyRequest) (*authpb.VerifyResponse, error) {

	log := s.log

	token := req.GetToken()
	const validToken = "demo-token"

	log.Debug("gRPC verify called", zap.String("token_present", "true"))

	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		log.Warn("gRPC: missing or invalid auth header")
		return &authpb.VerifyResponse{
			Valid: false,
			Error: "unauthorized",
		}, nil
	}

	actualToken := strings.TrimPrefix(token, "Bearer ")
	if actualToken == validToken {
		log.Info("gRPC: token verified", zap.String("subject", "student"))
		return &authpb.VerifyResponse{
			Valid:   true,
			Subject: "student",
		}, nil
	}

	log.Warn("gRPC: invalid token")
	return &authpb.VerifyResponse{
		Valid: false,
		Error: "unauthorized",
	}, nil
}
