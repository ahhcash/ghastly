package server

import (
	"context"
	"fmt"
	db2 "github.com/ahhcash/ghastlydb/db"
	pb "github.com/ahhcash/ghastlydb/grpc/gen/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type GhastlyServer struct {
	pb.UnimplementedGhastlyDBServer
	db *db2.DB
}

func NewGhastlyServer(db *db2.DB) *GhastlyServer {
	return &GhastlyServer{
		db: db,
	}
}

func (s *GhastlyServer) Put(_ context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	if err := s.db.Put(req.Key, req.Value); err != nil {
		return &pb.PutResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.PutResponse{Success: true}, nil
}

func (s *GhastlyServer) Get(_ context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, err := s.db.Get(req.Key)
	if err != nil {
		return &pb.GetResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	return &pb.GetResponse{
		Value: value,
		Found: true,
	}, nil
}

func (s *GhastlyServer) Delete(_ context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.db.Delete(req.Key)
	if err != nil {
		return &pb.DeleteResponse{
			Success: false,
			Error:   fmt.Sprintf("%v", err),
		}, err
	}

	return &pb.DeleteResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *GhastlyServer) Exists(_ context.Context, req *pb.ExistsRequest) (*pb.ExistsResponse, error) {
	exists := s.db.Exists(req.Key)
	return &pb.ExistsResponse{Exists: exists}, nil
}

func (s *GhastlyServer) Search(_ context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	results, err := s.db.Search(req.Query)
	if err != nil {
		return &pb.SearchResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	pbResults := make([]*pb.SearchResult, 0, len(results))
	for _, r := range results {
		if req.ScoreThreshold > 0 && r.Score < float64(req.ScoreThreshold) {
			continue
		}
		pbResults = append(pbResults, &pb.SearchResult{
			Key:   r.Key,
			Value: r.Value,
			Score: float32(r.Score),
		})
	}

	if req.Limit > 0 && int32(len(pbResults)) > req.Limit {
		pbResults = pbResults[:req.Limit]
	}

	return &pb.SearchResponse{Results: pbResults}, nil
}

func (s *GhastlyServer) GetConfig(_ context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	return &pb.GetConfigResponse{
		Config: &pb.DatabaseConfig{
			MemtableSizeBytes:          0,
			DataDirectory:              s.db.DBConfig.Path,
			DefaultSimilarityMetric:    s.db.DBConfig.Metric,
			DefaultSimilarityThreshold: 0,
			EmbeddingModel:             s.db.DBConfig.EmbeddingModel,
		},
	}, nil
}

func (s *GhastlyServer) BulkPut(stream pb.GhastlyDB_BulkPutServer) error {
	var processed int32
	var failed []string

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&pb.BulkPutResponse{
					ProcessedCount: processed,
					FailedKeys:     failed,
				})
			}
			return err
		}

		if err := s.db.Put(req.Key, req.Value); err != nil {
			failed = append(failed, req.Key)
		} else {
			processed++
		}
	}
}

func (s *GhastlyServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}
