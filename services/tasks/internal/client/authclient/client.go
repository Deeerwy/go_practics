package authclient

import (
	"context"
	"errors"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"tech-ip-sem2/proto/authpb"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Client struct {
	cc     *grpc.ClientConn
	client authpb.AuthServiceClient
}

func New(cc *grpc.ClientConn) *Client {
	return &Client{
		cc:     cc,
		client: authpb.NewAuthServiceClient(cc),
	}
}

// Verify вызывает gRPC-метод AuthService.Verify.
func (c *Client) Verify(ctx context.Context, authorization string) error {
	// В ПЗ токен передаётся в Authorization, здесь ожидаем "Bearer <token>".
	if authorization == "" {
		return ErrUnauthorized
	}

	const prefix = "Bearer "
	token := authorization
	if len(authorization) > len(prefix) && authorization[:len(prefix)] == prefix {
		token = authorization[len(prefix):]
	}

	log.Println("calling grpc verify")

	resp, err := c.client.Verify(ctx, &authpb.VerifyRequest{Token: token})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}
		switch st.Code() {
		case codes.Unauthenticated:
			return ErrUnauthorized
		case codes.PermissionDenied:
			return ErrForbidden
		default:
			return err
		}
	}

	if !resp.Valid {
		return ErrUnauthorized
	}
	return nil
}


