package models

import (
	"context"
	"database/sql"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	container      *PostgresContainer
	models         Models
	email          = "testmailbob@pubsub.com"
	username       = "bob-cool"
	profilePicture = "https://lh3.googleusercontent.com/a/default-user=s96-c"
	conn           *sql.DB
	userId         int64
)

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()

	postgres, err := CreatePostgresContainer(ctx)
	if err != nil {
		panic(err)
	}
	container = postgres

	dbConn, err := GetDb(container.ConnectionString)
	if err != nil {
		panic(err)
	}

	models = NewModels(dbConn)
	conn = dbConn

	query := `
		insert into users (email, name, username, profile_picture) values (
		$1, 'bob barry', $2, $3  
		)
		returning id
		`

	insertCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = dbConn.QueryRowContext(insertCtx, query, email, username, profilePicture).Scan(&userId)
	if err != nil {
		panic(err)
	}

})

var _ = AfterSuite(func() {
	ctx := context.Background()
	err := container.Terminate(ctx)
	if err != nil {
		panic(err)
	}
})

var _ = Describe("getting a list of posts", Label("unit"), func() {
	When("there are no posts in the db", func() {
		It("should return an empty slice", func() {
			posts, _, err := models.Posts.List(10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(posts).To(BeEmpty())
		})
	})

	When("there are posts in the db", func() {
		BeforeEach(func() {
			query := `
			insert into posts (body, user_id) values 
			('hello world', $1),
			('hello world 2', $1),
			('hello world 3', $1),
			('hello world 4', $1),
			('hello world 5', $1),
			('hello world 6', $1),
			('hello world 7', $1),
			('hello world 8', $1),
			('hello world 9', $1),
			('hello world 10', $1),
			('hello world 11', $1),
			('hello world 12', $1),
			('hello world 13', $1),
			('hello world 14', $1),
			('hello world 15', $1),
			('hello world 16', $1),
			('hello world 17', $1),
			('hello world 18', $1),
			('hello world 19', $1),
			('hello world 20', $1)
			`

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := conn.ExecContext(ctx, query, userId)
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			query := `
			delete from posts
			`

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := conn.ExecContext(ctx, query)
			if err != nil {
				panic(err)
			}
		})

		It("should return a list of posts", func() {
			posts, _, err := models.Posts.List(10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(posts)).Should(BeNumerically(">", 0))
		})
		It("should have user with profile picture, username and id", func() {
			posts, _, err := models.Posts.List(10, 0)
			Expect(err).ToNot(HaveOccurred())

			Expect(posts[0].Post.User.Id).To(Equal(userId))
			Expect(posts[0].Post.User.Username).To(Equal(username))
			Expect(posts[0].Post.User.ProfilePicture).To(Equal(profilePicture))
		})
		It("should have a body", func() {
			posts, _, err := models.Posts.List(10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(posts[0].Post.Body).ToNot(Equal(""))
		})
		It("should be controlled via pagination", func() {
			posts, _, err := models.Posts.List(10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(posts)).To(Equal(10))

			finalId := posts[len(posts)-1].Post.Id

			posts, _, err = models.Posts.List(10, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(posts[len(posts)-1].Post.Id).ToNot(Equal(finalId))

			posts, _, err = models.Posts.List(1, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(posts)).To(Equal(1))
		})

		When("10 posts are taken, metadata should reflect that", func() {
			It("page size is 10", func() {
				_, metadata, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.PageSize).To(Equal(10))
			})
		})

		When("posts have no comments", func() {
			It("post metadata should have 0 comments indicated", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Metadata.CommentsCount).To(Equal(0))
			})
			It("post metadata should show last comment as empty string", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Metadata.LatestComment).To(Equal(""))
			})
			It("post metadata last comment at should be empty time", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Metadata.LastCommentAt).To(Equal(time.Time{}))
			})
			It("post comments should be nil", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Post.Comments).To(BeNil())
			})
		})

		When("posts have comments", func() {
			BeforeEach(func() {
				var postsIds []int64
				query := `
					select id from posts
				`

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				rows, err := conn.QueryContext(ctx, query)
				if err != nil {
					panic(err)
				}

				defer rows.Close()

				for rows.Next() {
					var postId int64
					err := rows.Scan(&postId)
					if err != nil {
						panic(err)
					}

					postsIds = append(postsIds, postId)
				}

				if err = rows.Err(); err != nil {
					panic(err)
				}

				for _, postId := range postsIds {
					query := `
					insert into comments (body, user_id, post_id, path) values 
					('hello world', $1, $2, '0')
					`

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					_, err := conn.ExecContext(ctx, query, userId, postId)
					if err != nil {
						panic(err)
					}
				}
			})

			AfterEach(func() {
				query := `
				delete from comments
				`

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err := conn.ExecContext(ctx, query)
				if err != nil {
					panic(err)
				}
			})

			It("post metadata should show num of comments", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Metadata.CommentsCount).To(Equal(1))
			})
			It("post metadata should show last comment body", func() {
				posts, _, err := models.Posts.List(10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(posts[0].Metadata.LatestComment).To(Equal("hello world"))
			})
		})
	})
})
