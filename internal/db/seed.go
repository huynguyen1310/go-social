package db

import (
	"context"
	"fmt"
	"log"

	"github.com/huynguyen1310/social/internal/store"
)

func Seed(store store.Storage) error {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, nil, user); err != nil {
			return err
		}
	}

	posts := generatePosts(1000, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			return err
		}
	}

	comments := generateComments(5000, posts, users)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			return err
		}
	}

	log.Println("seed completed successfully")
	return nil
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := range users {
		user := &store.User{
			Username: fmt.Sprintf("user%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
		}
		user.Password.Set("password")
		users[i] = user
	}
	return users
}

func generatePosts(n int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, n)
	for i := range posts {
		user := users[i%len(users)]
		posts[i] = &store.Post{
			Title:   fmt.Sprintf("post%d", i+1),
			Content: fmt.Sprintf("content%d", i+1),
			UserId:  user.ID,
			Tags:    []string{"go", "postgres", "seed"},
		}
	}
	return posts
}

func generateComments(n int, posts []*store.Post, users []*store.User) []*store.Comment {
	comments := make([]*store.Comment, n)
	for i := range comments {
		user := users[i%len(users)]
		post := posts[i%len(posts)]
		comments[i] = &store.Comment{
			PostID:  post.ID,
			UserID:  user.ID,
			Content: fmt.Sprintf("comment %d from user %d on post %d", i+1, user.ID, post.ID),
		}
	}
	return comments
}
