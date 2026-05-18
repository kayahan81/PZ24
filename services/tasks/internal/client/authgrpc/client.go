package authgrpc

import (
	"context"
	"time"

	"tech-ip-sem2/services/auth/pkg/authpb"
	"tech-ip-sem2/shared/middleware"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
	log    *zap.Logger
}

func NewClient(addr string, log *zap.Logger) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
		log:    log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) VerifyToken(ctx context.Context, token string) (bool, string, error) {
	requestID := middleware.GetRequestID(ctx)
	log := c.log.With(zap.String("request_id", requestID), zap.String("component", "auth_client"))

	log.Debug("calling auth.Verify via gRPC")

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := c.client.Verify(ctx, &authpb.VerifyRequest{
		Token: token,
	})
	duration := time.Since(start)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.DeadlineExceeded:
				log.Error("auth timeout", zap.Duration("duration_ms", duration))
				return false, "", ErrTimeout
			case codes.Unavailable:
				log.Error("auth unavailable", zap.Duration("duration_ms", duration))
				return false, "", ErrUnavailable
			default:
				log.Error("auth error", zap.Error(err), zap.String("code", st.Code().String()))
			}
		} else {
			log.Error("auth error", zap.Error(err))
		}
		return false, "", err
	}

	log.Info("auth response",
		zap.Bool("valid", resp.Valid),
		zap.String("subject", resp.Subject),
		zap.Duration("duration_ms", duration),
	)

	if !resp.Valid {
		return false, "", ErrUnauthorized
	}

	return true, resp.Subject, nil
}

var (
	ErrUnauthorized = &AuthError{Code: "UNAUTHORIZED", Message: "invalid token"}
	ErrTimeout      = &AuthError{Code: "TIMEOUT", Message: "auth service timeout"}
	ErrUnavailable  = &AuthError{Code: "UNAVAILABLE", Message: "auth service unavailable"}
)

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
