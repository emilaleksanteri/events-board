# for likes
- like for posts
- like for comments
- post & remove like comment lambda
- post & remove like post lambda
    - like db table -> 
        1. user_id -> 1:1 on user
        2. like_type (comment, post) 
        3. comment_id -> relation to comments, nullable
        4. post_id -> relation to posts, nullable
        5. liked_at

- get likes lambda:
    - get likes for a post (list users)
    - get likes for a comment (list users)

- include total likes for each comment and post query on comments and posts service
