package grpc

import (
	"post/internal/config"
	pb "post/internal/proto"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	envConf *config.Config
}

func NewPostServer(cfg *config.Config) *PostServiceServer {
	return &PostServiceServer{
		envConf: cfg,
	}
}
