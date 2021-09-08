package repositories

import (
	"context"
	Db "micrach/db"

	"github.com/jackc/pgx/v4"
)

type PostsRepository struct{}

var Posts PostsRepository

func (r *PostsRepository) Get(limit, offset int) ([]Post, error) {
	sql := `
		SELECT id, title, text, created_at
		FROM posts
		WHERE
			is_parent = true
			AND is_deleted = false
		ORDER BY updated_at DESC
		LIMIT $1
	`

	rows, err := Db.Pool.Query(context.TODO(), sql, limit)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	postsMap := make(map[int]Post)
	var postIDs []int
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Title, &post.Text, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		postsMap[post.ID] = post
		postIDs = append(postIDs, post.ID)
	}

	filesMap, err := Files.GetByPostIDs(postIDs)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, postID := range postIDs {
		post := postsMap[postID]
		post.Files = filesMap[postID]
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostsRepository) Create(p Post) (int, error) {
	sql := `
		INSERT INTO posts (is_parent, parent_id, title, text, is_sage)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var row pgx.Row
	if p.IsParent {
		row = Db.Pool.QueryRow(
			context.TODO(), sql, p.IsParent, nil, p.Title, p.Text, p.IsSage,
		)
	} else {
		row = Db.Pool.QueryRow(
			context.TODO(), sql, p.IsParent, p.ParentID, p.Title, p.Text, p.IsSage,
		)
	}

	createdPost := new(Post)
	err := row.Scan(&createdPost.ID)
	if err != nil {
		return 0, err
	}

	return createdPost.ID, nil
}

func (r *PostsRepository) GetThreadByPostID(ID int) ([]Post, error) {
	sql := `
		SELECT 
			id,
			title,
			text,
			is_sage,
			created_at,
			is_parent
		FROM posts
		WHERE
				(id = $1 AND is_parent = true) OR parent_id = $1
				AND is_deleted = false
		ORDER BY created_at ASC
	`

	rows, err := Db.Pool.Query(context.TODO(), sql, ID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	postsMap := make(map[int]Post)
	var postIDs []int
	for rows.Next() {
		var post Post
		err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Text,
			&post.IsSage,
			&post.CreatedAt,
			&post.IsParent,
		)
		if err != nil {
			return nil, err
		}

		postsMap[post.ID] = post
		postIDs = append(postIDs, post.ID)
	}

	filesMap, err := Files.GetByPostIDs(postIDs)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, postID := range postIDs {
		post := postsMap[postID]
		post.Files = filesMap[postID]
		posts = append(posts, post)
	}

	return posts, nil
}

// func (r *PostsRepository) IsThreadExists(ID int) ([]Post, error) {
// 	conn, err := Db.Pool.Acquire(context.TODO())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer conn.Release()

// 	sql
// }
