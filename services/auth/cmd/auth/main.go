package main

import (
	"net"
	"net/http"
	"os"

	authgrpc "tech-ip-sem2/services/auth/internal/grpc"
	"tech-ip-sem2/services/auth/internal/handler"
	"tech-ip-sem2/services/auth/pkg/authpb"
	"tech-ip-sem2/shared/logger"
	"tech-ip-sem2/shared/middleware"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log := logger.New("auth")
	defer log.Sync()

	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	log.Info("starting auth service",
		zap.String("http_port", httpPort),
		zap.String("grpc_port", grpcPort),
	)

	go startGRPCServer(grpcPort, log)

	startHTTPServer(httpPort, log)
}

func startGRPCServer(port string, log *zap.Logger) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	authpb.RegisterAuthServiceServer(s, authgrpc.NewAuthServer(log))

	log.Info("gRPC server listening", zap.String("port", port))
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve gRPC", zap.Error(err))
	}
}

func startHTTPServer(port string, log *zap.Logger) {
	mux := http.NewServeMux()
	authHandler := handler.NewAuthHandler(log)

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /v1/auth/verify", authHandler.Verify)

	handler := middleware.RequestIDMiddleware(
		middleware.AccessLogMiddleware(log)(mux),
	)

	log.Info("HTTP server listening", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("failed to serve HTTP", zap.Error(err))
	}
}
