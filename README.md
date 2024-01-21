# About project
Real time feed application built on aws infrastructure in microservices architecture (work in progress) using [localstack](https://www.localstack.cloud/)

# Dev
Make sure to have [cdklocal](https://github.com/localstack/aws-cdk-local), [docker](https://docs.docker.com/engine/install/) and [migrate](https://github.com/golang-migrate/migrate) installed
for testing make sure to have [ginkgo](https://github.com/onsi/ginkgo) working and meet the requirements for [testcontainers](https://golang.testcontainers.org/)

Having [awslocal](https://github.com/localstack/awscli-local) can also help with local dev to browse resources within localstack

1. Run ```make help``` to see all available commands
2. Copy ```.envrc.example``` into an ```.envrc``` file and add secrets
3. If on a fresh db ```cd ./migrations && make db/migrations/up``` to get db set up
4. Run ```make start``` to start localstack dev environment
5. Run ```make build/lambdas && make build/infra``` to get resources ready
6. Run ```make bootstrap && make deploy``` to deploy localstack
