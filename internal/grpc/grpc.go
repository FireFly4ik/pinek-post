package grpc

import (
	"post/internal/config"
	"post/internal/db"
	pb "post/internal/proto"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	database *db.AuthDatabase
	envConf  *config.Config
}

func NewPostServer(db *db.AuthDatabase, cfg *config.Config) *PostServiceServer {
	return &PostServiceServer{
		database: db,
		envConf:  cfg,
	}
}
