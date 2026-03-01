package models

import (
	"database/sql"
	"time"
)

//--------------------------------------------------------------------------------------|

type User struct {
	ID           int            `db:"id" json:"id"`
	Username     string         `db:"username" json:"username"`
	Email        string         `db:"email" json:"email"`
	FirstName    string         `db:"first_name" json:"first_name"`
	LastName     string         `db:"last_name" json:"last_name"`
	Age          int            `db:"age" json:"age"`
	Gender       string         `db:"gender" json:"gender"`
	AvatarPath   sql.NullString `db:"avatar_path" json:"avatar_path,omitempty"`
	PasswordHash string         `db:"password_hash" json:"-"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	Role         string         `db:"role" json:"role"`
}

//--------------------------------------------------------------------------------------|

type Post struct {
	ID         int            `db:"id" json:"id"`
	UserID     int            `db:"user_id" json:"user_id"`
	Username   string         `db:"username" json:"username"`
	AvatarPath sql.NullString `db:"avatar_path" json:"avatar_path,omitempty"`
	Title      string         `db:"title" json:"title"`
	Body       string         `db:"body" json:"body"`
	ImagePath  sql.NullString `db:"image_path" json:"image_path,omitempty"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	Likes      int            `db:"likes" json:"likes"`
	Dislikes   int            `db:"dislikes" json:"dislikes"`
	UserLike   sql.NullInt64  `db:"user_like" json:"user_like"`
	Categories []Category     `json:"categories"`
}

//--------------------------------------------------------------------------------------|

type Comment struct {
	ID         int            `db:"id" json:"id"`
	PostID     int            `db:"post_id" json:"post_id"`
	UserID     int            `db:"user_id" json:"user_id"`
	ViewerID   int            `db:"viewer_id" json:"viewer_id"`
	Username   string         `db:"username" json:"username"`
	AvatarPath sql.NullString `db:"avatar_path" json:"avatar_path,omitempty"`
	ParentID   sql.NullInt64  `db:"parent_id" json:"parent_id,omitempty"`
	Body       string         `db:"body" json:"body"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	Likes      int            `db:"likes" json:"likes"`
	Dislikes   int            `db:"dislikes" json:"dislikes"`
	UserLike   sql.NullInt64  `db:"user_like" json:"user_like"`
	Depth      int            `json:"depth"`
	MaxDepth   int            `json:"max_depth"`
	Children   []*Comment     `json:"children"`
}

//--------------------------------------------------------------------------------------|

type Category struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

//--------------------------------------------------------------------------------------|

type Session struct {
	ID        string    `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

//--------------------------------------------------------------------------------------|

type Message struct {
	ID         int            `db:"id" json:"id"`
	SenderID   int            `db:"sender_id" json:"sender_id"`
	ReceiverID int            `db:"receiver_id" json:"receiver_id"`
	Body       string         `db:"body" json:"body"`
	ImagePath  sql.NullString `db:"image_path" json:"image_path,omitempty"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

//--------------------------------------------------------------------------------------|

type Profile struct {
	ID           int            `json:"id"`
	Username     string         `json:"username"`
	FirstName    string         `json:"first_name"`
	LastName     string         `json:"last_name"`
	Age          int            `json:"age"`
	Gender       string         `json:"gender"`
	AvatarPath   sql.NullString `json:"avatar_path,omitempty"`
	Email        string         `json:"email,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	PostCount    int            `json:"post_count"`
	CommentCount int            `json:"comment_count"`
	LikeCount    int            `json:"like_count"`
}
