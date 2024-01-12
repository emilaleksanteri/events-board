# todo eh
- break current structure into microservcies -> start with most basic functionality
- post service for making posts
    - if img send img to s3 bucket, store key in db w post
    - push event to event bridge

- social connection service -> follow, unfollow
    - makes/deletes connections on social graph
    - add follower and followers count on user table

- like service for liking posts
    - on like update in db, push event to event bridge

- comment service for commenting on posts
    - on comment, update db, push event to event bridge

- auth service for authentication
    - on auth, make session, save in dynamoDB

- feed service for feed related stuff
    - get feed based on followers etc

- notifications service for real time notification
    - add notification types -> push notification on post comment and post
    - social graph to find who needs to be notified -> any node connected to the post user id will be notidied, how?:
        - websocket sessions stored in dynamoDB, contains connection id and user id (only signed in users can have live)
        - on event push include pushed id, search postgresdb for connected nodes, get their ids and filter from dynamo
        - based on these ids, get their connectionID and push data to socket
        
- api for as api gateway
    - use api gateway to interact w services
    - non cold start gateway?

- centralized logging service
