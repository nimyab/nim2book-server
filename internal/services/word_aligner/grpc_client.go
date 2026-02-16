package word_aligner

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go/v4"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Address string
}

type Client struct {
	pb.AlignmentServiceClient
	conn *grpc.ClientConn
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	const operation = "align.NewClient"

	var opts []grpc.DialOption

	if strings.HasSuffix(cfg.Address, ":443") {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Client{
		AlignmentServiceClient: pb.NewAlignmentServiceClient(conn),
		conn:                   conn,
	}, nil
}

func (c *Client) Align(ctx context.Context, in *pb.AlignRequest, opts ...grpc.CallOption) (*pb.AlignResponse, error) {
	return retry.DoWithData(func() (*pb.AlignResponse, error) {
		return c.AlignmentServiceClient.Align(ctx, in, opts...)
	}, retry.Attempts(5))
}

// Close closes the gRPC client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
