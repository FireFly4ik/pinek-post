package grpc

import (
	"context"
	"post/internal/config"
	"post/internal/db"
	pb "post/internal/proto"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	database *db.PostDatabase
	envConf  *config.Config
}

func NewPostServer(db *db.PostDatabase, cfg *config.Config) *PostServiceServer {
	return &PostServiceServer{
		database: db,
		envConf:  cfg,
	}
}

func (p *PostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	postId, err := p.database.CreatePost(req.PostId, req.UserId, req.Title, description)
	if err != nil {
		return nil, err
	}

	return &pb.CreatePostResponse{
		PostId: postId,
	}, nil
}

func (p *PostServiceServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	var description, title *string
	if req.Description != "" {
		description = &req.Description
	}
	if req.Title != "" {
		title = &req.Title
	}

	err := p.database.UpdatePost(req.PostId, req.UserId, title, description)
	if err != nil {
		return nil, err
	}

	return &pb.UpdatePostResponse{
		Message: "successfully updated post",
	}, nil
}

func (p *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	postId, userId, title, description, extension, tagsIds, err := p.database.GetPost(req.PostId)
	if err != nil {
		return nil, err
	}

	tags := make([]*pb.Tag, len(tagsIds))
	for i, tagId := range tagsIds {
		tags[i] = &pb.Tag{TagId: tagId[0], Name: tagId[1]}
	}

	resp := &pb.Post{
		PostId:    postId,
		UserId:    userId,
		Title:     title,
		Extension: extension,
		Tags:      tags,
	}

	if description != nil {
		resp.Description = *description
	}

	return &pb.GetPostResponse{
		Post: resp,
	}, nil
}

func (p *PostServiceServer) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	postIds, userIds, titles, descriptions, extensions, tagsIds, err := p.database.GetPosts(req.PostIds)
	if err != nil {
		return nil, err
	}

	posts := make([]*pb.Post, len(postIds))
	for i := range postIds {
		tags := make([]*pb.Tag, len(tagsIds[i]))
		for j, tagId := range tagsIds[i] {
			tags[j] = &pb.Tag{TagId: tagId[0], Name: tagId[1]}
		}

		posts[i] = &pb.Post{
			PostId:    postIds[i],
			UserId:    userIds[i],
			Title:     titles[i],
			Extension: extensions[i],
			Tags:      tags,
		}

		if descriptions[i] != nil {
			posts[i].Description = *descriptions[i]
		}
	}

	return &pb.GetPostsResponse{
		Posts: posts,
	}, nil
}

func (p *PostServiceServer) SearchPosts(ctx context.Context, req *pb.SearchPostsRequest) (*pb.SearchPostsResponse, error) {
	var tagIds, userIds []string
	if req.TagId != nil && len(req.TagId) > 0 {
		tagIds = req.TagId
	}

	if req.UserId != nil && len(req.UserId) > 0 {
		userIds = req.UserId
	}

	var query *string
	if req.Query != "" {
		query = &req.Query
	}

	postIds, userIdsRes, titles, descriptions, extensions, tagsIdsResp, err := p.database.SearchPosts(query, userIds, tagIds, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	posts := make([]*pb.Post, len(postIds))
	for i := range postIds {
		tags := make([]*pb.Tag, len(tagsIdsResp))
		for j, tagId := range tagsIdsResp[i] {
			if len(tagId) == 0 {
				continue
			}
			tags[j] = &pb.Tag{TagId: tagId[0], Name: tagId[1]}
		}

		posts[i] = &pb.Post{
			PostId:    postIds[i],
			UserId:    userIdsRes[i],
			Title:     titles[i],
			Extension: extensions[i],
			Tags:      tags,
		}

		if descriptions[i] != nil {
			posts[i].Description = *descriptions[i]
		}
	}

	return &pb.SearchPostsResponse{
		Posts: posts,
	}, nil
}

func (p *PostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	err := p.database.DeletePost(req.PostId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.DeletePostResponse{
		Message: "successfully deleted post",
	}, nil
}

func (p *PostServiceServer) CreateBoard(ctx context.Context, req *pb.CreateBoardRequest) (*pb.CreateBoardResponse, error) {
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	boardId, err := p.database.CreateBoard(req.UserId, req.Name, description)
	if err != nil {
		return nil, err
	}

	return &pb.CreateBoardResponse{
		BoardId: boardId,
	}, nil
}

func (p *PostServiceServer) UpdateBoard(ctx context.Context, req *pb.UpdateBoardRequest) (*pb.UpdateBoardResponse, error) {
	var description, name *string
	if req.Description != "" {
		description = &req.Description
	}
	if req.Name != "" {
		name = &req.Name
	}

	err := p.database.UpdateBoard(req.BoardId, req.UserId, name, description)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateBoardResponse{
		Message: "successfully updated board",
	}, nil
}

func (p *PostServiceServer) GetBoard(ctx context.Context, req *pb.GetBoardRequest) (*pb.GetBoardResponse, error) {
	boardId, userId, name, description, pinnedPostIds, err := p.database.GetBoard(req.BoardId)
	if err != nil {
		return nil, err
	}

	posts := make([]*pb.Post, len(pinnedPostIds))
	for i, postId := range pinnedPostIds {
		posts[i] = &pb.Post{PostId: postId[0], UserId: postId[1], Title: postId[2], Description: postId[3], Extension: postId[4]}
	}

	resp := &pb.Board{
		BoardId: boardId,
		UserId:  userId,
		Name:    name,
		Posts:   posts,
	}

	if description != nil {
		resp.Description = *description
	}

	return &pb.GetBoardResponse{
		Board: resp,
	}, nil
}

func (p *PostServiceServer) GetBoards(ctx context.Context, req *pb.GetBoardsRequest) (*pb.GetBoardsResponse, error) {
	boardIds, userIds, names, descriptions, pinnedPostIds, err := p.database.GetBoards(req.BoardIds)
	if err != nil {
		return nil, err
	}

	boards := make([]*pb.Board, len(boardIds))
	for i := range boardIds {
		post := make([]*pb.Post, len(pinnedPostIds[i]))
		for j, postId := range pinnedPostIds[i] {
			post[j] = &pb.Post{PostId: postId[0], UserId: postId[1], Title: postId[2], Description: postId[3], Extension: postId[4]}
		}

		boards[i] = &pb.Board{
			BoardId: boardIds[i],
			UserId:  userIds[i],
			Name:    names[i],
			Posts:   post,
		}

		if descriptions[i] != nil {
			boards[i].Description = *descriptions[i]
		}
	}
	return &pb.GetBoardsResponse{
		Boards: boards,
	}, nil
}

func (p *PostServiceServer) SearchBoards(ctx context.Context, req *pb.SearchBoardsRequest) (*pb.SearchBoardsResponse, error) {
	var query *string
	if req.Query != "" {
		query = &req.Query
	}

	var userIds []string
	if req.UserId != nil && len(req.UserId) > 0 {
		userIds = req.UserId
	}

	var postIds []string
	if req.PostId != nil && len(req.PostId) > 0 {
		postIds = req.PostId
	}

	boardIds, userIdsRes, names, descriptions, pinnedPostIds, err := p.database.SearchBoards(query, userIds, postIds, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	boards := make([]*pb.Board, len(boardIds))
	for i := range boardIds {
		post := make([]*pb.Post, len(pinnedPostIds[i]))
		for j, postId := range pinnedPostIds[i] {
			post[j] = &pb.Post{PostId: postId[0], UserId: postId[1], Title: postId[2], Description: postId[3], Extension: postId[4]}
		}

		boards[i] = &pb.Board{
			BoardId: boardIds[i],
			UserId:  userIdsRes[i],
			Name:    names[i],
			Posts:   post,
		}

		if descriptions[i] != nil {
			boards[i].Description = *descriptions[i]
		}
	}

	return &pb.SearchBoardsResponse{
		Boards: boards,
	}, nil
}

func (p *PostServiceServer) DeleteBoard(ctx context.Context, req *pb.DeleteBoardRequest) (*pb.DeleteBoardResponse, error) {
	err := p.database.DeleteBoard(req.BoardId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteBoardResponse{
		Message: "successfully deleted board",
	}, nil
}

func (p *PostServiceServer) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.CreateTagResponse, error) {
	tagId, err := p.database.CreateTag(req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTagResponse{
		TagId: tagId,
	}, nil
}

func (p *PostServiceServer) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	tagId, name, err := p.database.GetTag(req.TagId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTagResponse{
		TagId: tagId,
		Name:  name,
	}, nil

}

func (p *PostServiceServer) SearchTags(ctx context.Context, req *pb.SearchTagsRequest) (*pb.SearchTagsResponse, error) {
	var query *string
	if req.Query != "" {
		query = &req.Query
	}

	tagIds, names, err := p.database.SearchTags(query, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	tags := make([]*pb.Tag, len(tagIds))
	for i := range tagIds {
		tags[i] = &pb.Tag{
			TagId: tagIds[i],
			Name:  names[i],
		}
	}

	return &pb.SearchTagsResponse{
		Tags: tags,
	}, nil
}

func (p *PostServiceServer) PinPostToBoard(ctx context.Context, req *pb.PinPostToBoardRequest) (*pb.PinPostToBoardResponse, error) {
	err := p.database.PinPostToBoard(req.PostId, req.BoardId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.PinPostToBoardResponse{
		Message: "successfully pinned post to board",
	}, nil
}

func (p *PostServiceServer) UnpinPostFromBoard(ctx context.Context, req *pb.UnpinPostFromBoardRequest) (*pb.UnpinPostFromBoardResponse, error) {
	err := p.database.UnpinPostFromBoard(req.PostId, req.BoardId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.UnpinPostFromBoardResponse{
		Message: "successfully unpinned post from board",
	}, nil
}

func (p *PostServiceServer) AddTagToPost(ctx context.Context, req *pb.AddTagToPostRequest) (*pb.AddTagToPostResponse, error) {
	err := p.database.AddTagToPost(req.PostId, req.TagId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.AddTagToPostResponse{
		Message: "successfully added tag to post",
	}, nil
}

func (p *PostServiceServer) RemoveTagFromPost(ctx context.Context, req *pb.RemoveTagFromPostRequest) (*pb.RemoveTagFromPostResponse, error) {
	err := p.database.RemoveTagFromPost(req.PostId, req.TagId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.RemoveTagFromPostResponse{
		Message: "successfully removed tag from post",
	}, nil
}
