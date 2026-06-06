package grpc

import (
	"context"
	"database/sql"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/proto/pb"
)

type GrpcServer struct {
	pb.UnimplementedJobBoardServiceServer
	queries *db.Queries
	log     *logger.Logger
}

func NewGrpcServer(queries *db.Queries, log *logger.Logger) *GrpcServer {
	return &GrpcServer{
		queries: queries,
		log:     log,
	}
}

// GetJobDetails returns a job's details including its parent company owner's profile (name and email)
func (s *GrpcServer) GetJobDetails(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	s.log.Infow("gRPC GetJobDetails request received", "jobID", req.JobId)

	jobUUID, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid job id: %v", err)
	}

	// 1. Fetch job
	job, err := s.queries.GetJobByID(ctx, jobUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "job not found")
		}
		s.log.Errorw("failed to fetch job by ID", "error", err, "jobID", jobUUID)
		return nil, status.Errorf(codes.Internal, "internal database error")
	}

	// 2. Fetch company
	company, err := s.queries.GetCompanyByID(ctx, job.CompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "company not found")
		}
		s.log.Errorw("failed to fetch company by ID", "error", err, "companyID", job.CompanyID)
		return nil, status.Errorf(codes.Internal, "internal database error")
	}

	// 3. Fetch company owner/user details
	user, err := s.queries.GetUserByID(ctx, company.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "company owner not found")
		}
		s.log.Errorw("failed to fetch user by ID", "error", err, "userID", company.UserID)
		return nil, status.Errorf(codes.Internal, "internal database error")
	}

	return &pb.GetJobResponse{
		Id:           job.ID.String(),
		Title:        job.Title,
		CompanyId:    company.ID.String(),
		CompanyName:  company.Name,
		CompanyEmail: user.Email,
	}, nil
}

// GetUserDetails returns basic profile details (name and email) of a user
func (s *GrpcServer) GetUserDetails(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.log.Infow("gRPC GetUserDetails request received", "userID", req.UserId)

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	user, err := s.queries.GetUserByID(ctx, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		s.log.Errorw("failed to fetch user by ID", "error", err, "userID", userUUID)
		return nil, status.Errorf(codes.Internal, "internal database error")
	}

	return &pb.GetUserResponse{
		Id:       user.ID.String(),
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

// StartServers registers and runs the gRPC server listening on a specified port
func StartServer(addr string, queries *db.Queries, log *logger.Logger) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	srv := NewGrpcServer(queries, log)
	pb.RegisterJobBoardServiceServer(grpcServer, srv)

	log.Infow("gRPC Server listening", "addr", addr)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Errorw("gRPC server failed to serve", "error", err)
		}
	}()

	return grpcServer, nil
}
