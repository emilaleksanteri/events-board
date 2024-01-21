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
	postId         int64
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

	dbConn, err := OpenDB(container.ConnectionString)
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

	query = `
		insert into posts (body, user_id) values (
		$1, $2
		)
		returning id
	`

	err = dbConn.QueryRowContext(insertCtx, query, "this is a post", userId).Scan(&postId)
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

var _ = Describe("Get comment", Label("unit"), func() {
	When("there are no comments in the db", func() {
		It("should return not found", func() {
			_, err := models.Comments.GetComment(99999, 10, 0)
			Expect(err).To(MatchError(ErrRecordNotFound))
		})
	})

	When("there are comments in the db", func() {
		commentCount := 20
		commentBody := "this is a comment"
		commentIds := make([]int64, commentCount)
		BeforeEach(func() {
			for i := 0; i < commentCount; i++ {
				query := `
					insert into comments (body, user_id, post_id, path) values (
					$1, $2, $3, '0'
					)
					returning id
				`

				var commentId int64
				insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				err := conn.QueryRowContext(insertCtx, query, commentBody, userId, postId).Scan(&commentId)
				if err != nil {
					panic(err)
				}

				commentIds[i] = commentId
			}
		})

		AfterEach(func() {
			query := `
				delete from comments
			`

			insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := conn.ExecContext(insertCtx, query)
			if err != nil {
				panic(err)
			}
		})

		It("should include the comment body", func() {
			comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(comment.Body).To(Equal(commentBody))
		})
		It("should include a user with username, profile pic and id", func() {
			comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(comment.User.Username).To(Equal(username))
			Expect(comment.User.ProfilePicture).To(Equal(profilePicture))
			Expect(comment.User.Id).To(Equal(userId))
		})

		When("a comment has no sub comments", func() {
			It("should have sub comments as an empty slice", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments).To(BeEmpty())
			})
			It("should have num of sub comments as 0", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.NumOfSubComments).To(Equal(0))
			})
		})

		When("a comment has sub comments", func() {
			subCommentCount := 10
			subCommentBody := "this is a sub comment"
			subCommentIds := make([]int64, subCommentCount)
			BeforeEach(func() {
				for i := 0; i < subCommentCount; i++ {
					query := `
						insert into comments (body, user_id, post_id, path) values (
						$1, $2, $3, $4::text::ltree
						)
						returning id
					`

					var commentId int64
					insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					err := conn.QueryRowContext(insertCtx, query, subCommentBody, userId, postId, commentIds[0]).Scan(&commentId)
					if err != nil {
						panic(err)
					}

					subCommentIds[i] = commentId
				}
			})

			AfterEach(func() {
				query := `
				delete from comments where path <@ $1::text::ltree
				`

				insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err := conn.ExecContext(insertCtx, query, commentIds[0])
				if err != nil {
					panic(err)
				}
			})

			It("should have sub comments", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments).ToNot(BeEmpty())
			})
			It("should have num of sub comments as the number of sub comments", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.NumOfSubComments).To(Equal(subCommentCount))
			})
			It("should have sub comments with the correct body", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments[0].Body).To(Equal(subCommentBody))
			})
			It("should have sub comments with the correct user", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments[0].User.Username).To(Equal(username))
				Expect(comment.SubComments[0].User.ProfilePicture).To(Equal(profilePicture))
				Expect(comment.SubComments[0].User.Id).To(Equal(userId))
			})
			It("should return sub comments always in the same order if state does not change in the db", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())

				comment2, err := models.Comments.GetComment(commentIds[0], 10, 0)
				Expect(err).ToNot(HaveOccurred())

				Expect(comment.SubComments).To(Equal(comment2.SubComments))
			})
			It("should be possible to use pagination on the sub comments", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 5, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments).To(HaveLen(5))

				lastId := comment.SubComments[len(comment.SubComments)-1].Id
				comment, err = models.Comments.GetComment(commentIds[0], 5, 5)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments).To(HaveLen(5))
				Expect(comment.SubComments[len(comment.SubComments)-1]).ToNot(Equal(lastId))
			})
			It("should be possible to just return the parent comment with pagination set to 0", func() {
				comment, err := models.Comments.GetComment(commentIds[0], 0, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(comment.SubComments).To(BeEmpty())
			})
		})
	})
})
