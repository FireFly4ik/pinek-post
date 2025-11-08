package db

import (
	"errors"
	"github.com/google/uuid"
	"strings"
)

var (
	ErrCannotParseUUID    = errors.New("cannot parse uuid")
	ErrPostNotFound       = errors.New("post not found")
	ErrBoardNotFound      = errors.New("board not found")
	ErrTagAlreadyExists   = errors.New("tag already exists")
	ErrTagNotFound        = errors.New("tag not found")
	ErrModelAlreadyExists = errors.New("model already exists")
)

func (d *PostDatabase) CreatePost(postId, userId, title string, description *string) (string, error) {
	postIdSplitted := strings.Split(postId, ".")

	postIdParsed, err := uuid.Parse(postIdSplitted[0])
	if err != nil {
		return "", ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return "", ErrCannotParseUUID
	}

	post := &Post{
		ID:          postIdParsed,
		UserID:      userIdParsed,
		Title:       title,
		Description: description,
		Extension:   postIdSplitted[1],
	}

	tx := d.Database.Begin()

	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", err
	}

	return postId, nil
}

func (d *PostDatabase) UpdatePost(postId, userId string, title, description *string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tx := d.Database.Begin()

	post := &Post{}

	if err := tx.Where("id = ? AND user_id = ?", postIdParsed, userIdParsed).First(post).Error; err != nil {
		tx.Rollback()
		return errors.Join(ErrPostNotFound, err)
	}

	if title != nil {
		post.Title = *title
	}

	if description != nil {
		post.Description = description
	}

	if err := tx.Save(post).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) GetPost(postId string) (string, string, string, *string, string, [][]string, error) {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return "", "", "", nil, "", nil, ErrCannotParseUUID
	}

	post := &Post{}
	if err := d.Database.Where("id = ?", postIdParsed).First(post).Error; err != nil {
		return "", "", "", nil, "", nil, errors.Join(ErrPostNotFound, err)
	}

	tags := []Tag{}
	if err := d.Database.Model(PostTag{}).Joins("JOIN tags ON post_tags.tag_id = tags.id").Where("post_tags.post_id = ?", postIdParsed).Find(&tags).Error; err != nil {
		return "", "", "", nil, "", nil, err
	}

	tagsStrings := make([][]string, len(tags))
	for i, tag := range tags {
		tagsStrings[i] = []string{tag.ID.String(), tag.Name}
	}

	return post.ID.String(), post.UserID.String(), post.Title, post.Description, post.Extension, tagsStrings, nil
}

func (d *PostDatabase) GetPosts(postIds []string) ([]string, []string, []string, []*string, []string, [][][]string, error) {
	postIdsResp := make([]string, len(postIds))
	userIds := make([]string, len(postIds))
	titles := make([]string, len(postIds))
	descriptions := make([]*string, len(postIds))
	extensions := make([]string, len(postIds))
	tags := make([][][]string, len(postIds))

	for i := range postIds {
		postId, userId, title, description, extension, tag, err := d.GetPost(postIds[i])
		if err == nil {
			postIdsResp[i] = postId
			userIds[i] = userId
			titles[i] = title
			descriptions[i] = description
			extensions[i] = extension
			tags[i] = tag
		}
	}

	return postIds, userIds, titles, descriptions, extensions, tags, nil
}

func (d *PostDatabase) SearchPosts(query *string, userIdsSearch, tagIdsSearch []string, limit, offset int) ([]string, []string, []string, []*string, []string, [][][]string, error) {
	dbQuery := d.Database.Model(&Post{})

	if query != nil {
		dbQuery = dbQuery.Where("title ILIKE ?", "%"+*query+"%")
	}

	if userIdsSearch != nil {
		userIdsParsed := []uuid.UUID{}
		for _, userId := range userIdsSearch {
			userIdParsed, err := uuid.Parse(userId)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, ErrCannotParseUUID
			}
			userIdsParsed = append(userIdsParsed, userIdParsed)
		}

		dbQuery = dbQuery.Where("user_id IN ?", userIdsParsed)
	}

	if tagIdsSearch != nil {
		tagIdsParsed := []uuid.UUID{}
		for _, tagId := range tagIdsSearch {
			tagIdParsed, err := uuid.Parse(tagId)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, ErrCannotParseUUID
			}
			tagIdsParsed = append(tagIdsParsed, tagIdParsed)
		}

		dbQuery = dbQuery.Joins("JOIN post_tags ON posts.id = post_tags.post_id").Where("post_tags.tag_id IN ?", tagIdsParsed)
	}

	posts := []Post{}
	if err := dbQuery.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	postIds := make([]string, len(posts))
	userIds := make([]string, len(posts))
	titles := make([]string, len(posts))
	descriptions := make([]*string, len(posts))
	extensions := make([]string, len(posts))
	tags := make([][][]string, len(posts))

	for i, post := range posts {
		postIds[i] = post.ID.String()
		userIds[i] = post.UserID.String()
		titles[i] = post.Title
		descriptions[i] = post.Description
		extensions[i] = post.Extension
		tags[i] = [][]string{}

		tagModels := []Tag{}
		if err := d.Database.Model(PostTag{}).Joins("JOIN tags ON post_tags.tag_id = tags.id").Where("post_tags.post_id = ?", post.ID.String()).Find(&tagModels).Error; err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}

		if len(tagModels) == 0 {
			continue
		}

		tagsOfPost := make([][]string, len(tagModels))
		for j, tag := range tagModels {
			tagsOfPost[j] = []string{tag.ID.String(), tag.Name}
		}

		tags[i] = tagsOfPost
	}

	return postIds, userIds, titles, descriptions, extensions, tags, nil
}

func (d *PostDatabase) DeletePost(postId, userId string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tx := d.Database.Begin()

	if err := tx.Where("id = ? AND user_id = ?", postIdParsed, userIdParsed).Delete(&Post{}).Error; err != nil {
		tx.Rollback()
		return errors.Join(ErrPostNotFound, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) CreateBoard(userId, name string, description *string) (string, error) {
	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return "", ErrCannotParseUUID
	}

	board := &Board{
		UserID:      userIdParsed,
		Name:        name,
		Description: description,
	}

	tx := d.Database.Begin()

	if err := tx.Create(board).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", err
	}

	return board.ID.String(), nil
}

func (d *PostDatabase) UpdateBoard(boardId, userId string, name, description *string) error {
	boardIdParsed, err := uuid.Parse(boardId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tx := d.Database.Begin()

	board := &Board{}

	if err := tx.Where("id = ? AND user_id = ?", boardIdParsed, userIdParsed).First(board).Error; err != nil {
		tx.Rollback()
		return errors.Join(ErrBoardNotFound, err)
	}

	if name != nil {
		board.Name = *name
	}

	if description != nil {
		board.Description = description
	}

	if err := tx.Save(board).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) GetBoard(boardId string) (string, string, string, *string, [][]string, error) {
	boardIdParsed, err := uuid.Parse(boardId)
	if err != nil {
		return "", "", "", nil, nil, ErrCannotParseUUID
	}

	board := &Board{}
	if err := d.Database.Where("id = ?", boardIdParsed).First(board).Error; err != nil {
		return "", "", "", nil, nil, errors.Join(ErrBoardNotFound, err)
	}

	posts := []Post{}
	if err := d.Database.Where("board_id = ?", boardIdParsed).Find(&posts).Error; err != nil {
		return "", "", "", nil, nil, err
	}

	postIds := make([][]string, len(posts))
	for i, post := range posts {
		postIds[i] = []string{post.ID.String(), post.UserID.String(), post.Title, *post.Description, post.Extension}
	}

	return board.ID.String(), board.UserID.String(), board.Name, board.Description, postIds, nil
}

func (d *PostDatabase) GetBoards(boardIds []string) ([]string, []string, []string, []*string, [][][]string, error) {
	userIds := make([]string, len(boardIds))
	names := make([]string, len(boardIds))
	descriptions := make([]*string, len(boardIds))
	postsIds := make([][][]string, len(boardIds))

	for i := range boardIds {
		boardId, userId, name, description, postIds, err := d.GetBoard(boardIds[i])
		if err == nil {
			boardIds[i] = boardId
			userIds[i] = userId
			names[i] = name
			descriptions[i] = description
			postsIds[i] = postIds
		}
	}

	return boardIds, userIds, names, descriptions, postsIds, nil
}

func (d *PostDatabase) SearchBoards(query *string, userId []string, limit, offset int) ([]string, []string, []string, []*string, [][][]string, error) {
	dbQuery := d.Database.Model(&Board{}).Joins("LEFT JOIN board_posts ON boards.id = board_posts.board_id")

	if query != nil {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+*query+"%")
	}

	if userId != nil {
		userIdsParsed := []uuid.UUID{}
		for _, uId := range userId {
			userIdParsed, err := uuid.Parse(uId)
			if err != nil {
				return nil, nil, nil, nil, nil, ErrCannotParseUUID
			}
			userIdsParsed = append(userIdsParsed, userIdParsed)
		}

		dbQuery = dbQuery.Where("user_id IN ?", userIdsParsed)
	}

	boards := []Board{}
	if err := dbQuery.Limit(limit).Offset(offset).Find(&boards).Error; err != nil {
		return nil, nil, nil, nil, nil, err
	}

	boardIds := make([]string, len(boards))
	userIds := make([]string, len(boards))
	names := make([]string, len(boards))
	descriptions := make([]*string, len(boards))
	postsIds := make([][][]string, len(boards))

	for i, board := range boards {
		boardIds[i] = board.ID.String()
		userIds[i] = board.UserID.String()
		names[i] = board.Name
		descriptions[i] = board.Description
		postModels := []Post{}
		if err := d.Database.Where("board_id = ?", board.ID).Find(&postModels).Error; err != nil {
			return nil, nil, nil, nil, nil, err
		}

		postIds := make([][]string, len(postModels))
		for i, post := range postModels {
			postIds[i] = []string{post.ID.String(), post.UserID.String(), post.Title, *post.Description, post.Extension}
		}

		postsIds[i] = postIds
	}

	return boardIds, userIds, names, descriptions, postsIds, nil
}

func (d *PostDatabase) DeleteBoard(boardId, userId string) error {
	boardIdParsed, err := uuid.Parse(boardId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tx := d.Database.Begin()

	if err := tx.Where("id = ? AND user_id = ?", boardIdParsed, userIdParsed).Delete(&Board{}).Error; err != nil {
		tx.Rollback()
		return errors.Join(ErrBoardNotFound, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) CreateTag(name string) (string, error) {
	tag := &Tag{
		Name: name,
	}

	tx := d.Database.Begin()

	alreadyExistsTag := &Tag{}
	if err := tx.Where("name = ?", name).First(alreadyExistsTag).Error; err == nil {
		tx.Rollback()
		return "", err
	}

	if alreadyExistsTag.ID != uuid.Nil {
		tx.Rollback()
		return "", ErrTagAlreadyExists
	}

	if err := tx.Create(tag).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", err
	}

	return tag.ID.String(), nil
}

func (d *PostDatabase) GetTag(tagId string) (string, string, error) {
	tagIdParsed, err := uuid.Parse(tagId)
	if err != nil {
		return "", "", ErrCannotParseUUID
	}

	tag := &Tag{}
	if err := d.Database.Where("id = ?", tagIdParsed).First(tag).Error; err != nil {
		return "", "", errors.Join(ErrTagNotFound, err)
	}

	return tag.ID.String(), tag.Name, nil
}

func (d *PostDatabase) SearchTags(query *string, limit, offset int) ([]string, []string, error) {
	dbQuery := d.Database.Model(&Tag{})

	if query != nil {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+*query+"%")
	}

	tags := []Tag{}
	if err := dbQuery.Limit(limit).Offset(offset).Find(&tags).Error; err != nil {
		return nil, nil, err
	}

	tagIds := make([]string, len(tags))
	names := make([]string, len(tags))

	for i, tag := range tags {
		tagIds[i] = tag.ID.String()
		names[i] = tag.Name
	}

	return tagIds, names, nil
}

func (d *PostDatabase) PinPostToBoard(postId, boardId, userId string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	boardIdParsed, err := uuid.Parse(boardId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	board := &Board{}
	if err := d.Database.Where("id = ? AND user_id = ?", boardIdParsed, userIdParsed).First(board).Error; err != nil {
		return errors.Join(ErrBoardNotFound, err)
	}

	boardPost := &BoardPost{
		BoardID: boardIdParsed,
		PostID:  postIdParsed,
	}

	alreadyExistsModel := &BoardPost{}
	if err := d.Database.Where("board_id = ? AND post_id = ?", boardIdParsed, postIdParsed).First(alreadyExistsModel).Error; err == nil {
		return nil
	}

	if alreadyExistsModel.BoardID != uuid.Nil && alreadyExistsModel.PostID != uuid.Nil {
		return ErrModelAlreadyExists
	}

	tx := d.Database.Begin()

	if err := tx.Create(boardPost).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) UnpinPostFromBoard(postId, boardId, userId string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	boardIdParsed, err := uuid.Parse(boardId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tx := d.Database.Begin()

	if err := tx.Joins("JOIN boards ON board_posts.board_id = boards.id").Error; err != nil {
		tx.Rollback()
		return errors.Join(ErrBoardNotFound, err)
	}

	if err := tx.Where("board_id = ? AND post_id = ? AND boards.user_id = ?", boardIdParsed, postIdParsed, userIdParsed).Delete(&BoardPost{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) AddTagToPost(postId, tagId, userId string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tagIdParsed, err := uuid.Parse(tagId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	post := &Post{}
	if err := d.Database.Where("id = ? AND user_id = ?", postIdParsed, userIdParsed).First(post).Error; err != nil {
		return errors.Join(ErrPostNotFound, err)
	}

	postTag := &PostTag{
		PostID: postIdParsed,
		TagID:  tagIdParsed,
	}

	alreadyExistsModel := &PostTag{}
	if err := d.Database.Where("post_id = ? AND tag_id = ?", postIdParsed, tagIdParsed).First(alreadyExistsModel).Error; err == nil {
		return nil
	}

	if alreadyExistsModel.PostID != uuid.Nil && alreadyExistsModel.TagID != uuid.Nil {
		return ErrModelAlreadyExists
	}

	tx := d.Database.Begin()

	if err := tx.Create(postTag).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (d *PostDatabase) RemoveTagFromPost(postId, tagId, userId string) error {
	postIdParsed, err := uuid.Parse(postId)
	if err != nil {
		return ErrCannotParseUUID
	}

	tagIdParsed, err := uuid.Parse(tagId)
	if err != nil {
		return ErrCannotParseUUID
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		return ErrCannotParseUUID
	}

	post := &Post{}
	if err := d.Database.Where("id = ? AND user_id = ?", postIdParsed, userIdParsed).First(post).Error; err != nil {
		return errors.Join(ErrPostNotFound, err)
	}

	tx := d.Database.Begin()

	if err := tx.Where("post_id = ? AND tag_id = ?", postIdParsed, tagIdParsed).Delete(&PostTag{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
