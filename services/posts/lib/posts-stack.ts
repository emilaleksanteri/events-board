import {
  CfnOutput, Stack, StackProps, Tags
} from 'aws-cdk-lib';
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

import {
  RestApi,
  LambdaIntegration,
} from "aws-cdk-lib/aws-apigateway";
import { Bucket, IBucket } from 'aws-cdk-lib/aws-s3';
import path = require('path');

enum BaseUrlPaths {
  HEALTH = "healthcheck",
  POSTS = "posts",
  BY_ID = "{id}",
  CREATE_POST = "create",
  UPDATE = "update",
  DELETE = "delete",
}

function createLambda(
  th: Construct,
  funcName: string,
  pathStr: string,
  bucket: IBucket,
  db_url: string,
  description?: string,
): lambda.Function {
  return new lambda.Function(th, funcName, {
    code: lambda.Code.fromBucket(
      bucket,
      path.join(__dirname, pathStr)
    ),
    runtime: lambda.Runtime.GO_1_X,
    handler: "main",
    functionName: funcName,
    description: description ?? `Lambda function for ${funcName}`,
    tracing: lambda.Tracing.ACTIVE,
    environment: {
      DB_ADDRESS: db_url
    }
  })
}

export class PostsStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);
    const db_url = process.env.DB_ADDRESS
    if (!db_url) {
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
      "../lambdas/getPosts",
      hotReloadBucket,
      db_url,
    )

    const lambdaCreate = createLambda(
      this,
      "createPostFunc",
      "../lambdas/postPost",
      hotReloadBucket,
      db_url
    )

    const lambdaUpdate = createLambda(
      this,
      "updatePostFunc",
      "../lambdas/updatePost",
      hotReloadBucket,
      db_url
    )

    const lambdaDelete = createLambda(
      this,
      "deletePostFunc",
      "../lambdas/deletePost",
      hotReloadBucket,
      db_url
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
