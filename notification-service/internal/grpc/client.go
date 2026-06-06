package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kidus-yoseph1/job-board-platform/proto/pb"
)

// JobBoardClient wraps the generated gRPC client for JobBoardService.
type JobBoardClient struct {
	client pb.JobBoardServiceClient
	conn   *grpc.ClientConn
}

// NewJobBoardClient dials the remote job-service gRPC server and returns a client wrapper.
func NewJobBoardClient(addr string) (*JobBoardClient, error) {
	// Connect to gRPC server without TLS/encryption for simplicity in development
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &JobBoardClient{
		client: pb.NewJobBoardServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying gRPC TCP network connection.
func (c *JobBoardClient) Close() error {
	return c.conn.Close()
}

// GetJobDetails queries the job-service for job information and company email/details.
func (c *JobBoardClient) GetJobDetails(ctx context.Context, jobID string) (*pb.GetJobResponse, error) {
	return c.client.GetJobDetails(ctx, &pb.GetJobRequest{JobId: jobID})
}

// GetUserDetails queries the job-service for user profile details like full name and email.
func (c *JobBoardClient) GetUserDetails(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
	return c.client.GetUserDetails(ctx, &pb.GetUserRequest{UserId: userID})
}
