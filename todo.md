# todo eh
- break current structure into microservcies
- post service for making posts
    - if img send img to s3 bucket, store key in db w post
    - push event to event bridge
- like service for liking posts
    - on like update in db, push event to event bridge
- comment service for commenting on posts
    - on comment, update db, push event to event bridge
- auth service for authentication
    - on auth, make session, save in dynamoDB
- feed service for feed related stuff
    - get feed based on followers etc
- notifications service for real time notification
    - websocket api w api gateway, listens to event bridge
    - separate gateway from once used by other services?
- api for as api gateway
    - use api gateway to interact w services
    - non cold start gateway?


