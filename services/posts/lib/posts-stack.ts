import { CfnOutput, Tags } from 'aws-cdk-lib';
import { Construct } from "constructs";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";
import { Bucket } from 'aws-cdk-lib/aws-s3';
import * as events from 'aws-cdk-lib/aws-events';
import { createLambda } from '../../../lib/lambda';
import * as path from "path"


enum BaseUrlPaths {
  HEALTH = "healthcheck",
  POSTS = "posts",
  BY_ID = "{id}",
  CREATE_POST = "create",
  UPDATE = "update",
  DELETE = "delete",
}


interface PostsProps {
  db_url?: string
  eventBus: events.EventBus
}

export class Posts extends Construct {
  constructor(scope: Construct, id: string, props: PostsProps) {
    super(scope, id);
    const { eventBus } = props

    if (!props.db_url) {
      throw new Error("DB env var is not set")
    }

    const hotReloadBucket = Bucket.fromBucketName(
      this,
      "HotReloadingBucket",
      "hot-reload"
    )

    const lambdaPosts = createLambda(
      this,
      "getPostsFunc",
      path.join(__dirname, "../lambdas/getPosts"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },
    )

    const lambdaCreate = createLambda(
      this,
      "createPostFunc",
      path.join(__dirname, "../lambdas/postPost"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url, BUS_NAME: eventBus.eventBusName },
    )
    eventBus.grantPutEventsTo(lambdaCreate)

    const lambdaUpdate = createLambda(
      this,
      "updatePostFunc",
      path.join(__dirname, "../lambdas/updatePost"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },
    )

    const lambdaDelete = createLambda(
      this,
      "deletePostFunc",
      path.join(__dirname, "../lambdas/deletePost"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },

    )

    const api = new RestApi(this, "postsApi", {
      restApiName: "postsApi",
      description: "API for posts",
    })
    Tags.of(api).add("_custom_id_", "postsApi")


    // POSTS (GET)
    const integration = new LambdaIntegration(lambdaPosts)
    api.root.addMethod("GET", integration)

    const posts = api.root.addResource(BaseUrlPaths.POSTS)
    posts.addMethod("GET", integration)

    const post = posts.addResource(BaseUrlPaths.BY_ID)
    post.addMethod("GET", integration)

    const health = posts.addResource(BaseUrlPaths.HEALTH)
    health.addMethod("GET", integration)

    // CREATE (POST)
    const createIntegration = new LambdaIntegration(lambdaCreate)
    const create = api.root.addResource(BaseUrlPaths.CREATE_POST)
    create.addMethod("POST", createIntegration)

    const createHealth = create.addResource(BaseUrlPaths.HEALTH)
    createHealth.addMethod("GET", createIntegration)

    // UPDATE (PUT)
    const updateIntegration = new LambdaIntegration(lambdaUpdate)
    const update = api.root.addResource(BaseUrlPaths.UPDATE)
    const updatePost = update.addResource(BaseUrlPaths.BY_ID)
    updatePost.addMethod("PUT", updateIntegration)

    const updateHealth = update.addResource(BaseUrlPaths.HEALTH)
    updateHealth.addMethod("GET", updateIntegration)

    // DELETE (DELETE)
    const deleteIntegration = new LambdaIntegration(lambdaDelete)
    const deleteResource = api.root.addResource(BaseUrlPaths.DELETE)
    const deletePost = deleteResource.addResource(BaseUrlPaths.BY_ID)
    deletePost.addMethod("DELETE", deleteIntegration)

    const deleteHealth = deleteResource.addResource(BaseUrlPaths.HEALTH)
    deleteHealth.addMethod("GET", deleteIntegration)

    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
    new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
  }
}
