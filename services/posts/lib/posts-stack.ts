import {
  CfnOutput, Stack, StackProps
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
  POST = "{id}",
  CREATE_POST = "create",
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

    const api = new RestApi(this, "postsApi");

    // POSTS (GET)
    const integration = new LambdaIntegration(lambdaPosts)
    api.root.addMethod("GET", integration)

    const posts = api.root.addResource(BaseUrlPaths.POSTS)
    posts.addMethod("GET", integration)

    const post = posts.addResource(BaseUrlPaths.POST)
    post.addMethod("GET", integration)

    const health = posts.addResource(BaseUrlPaths.HEALTH)
    health.addMethod("GET", integration)

    // CREATE (POST)
    const createIntegration = new LambdaIntegration(lambdaCreate)
    const create = api.root.addResource(BaseUrlPaths.CREATE_POST)
    create.addMethod("POST", createIntegration)

    const createHealth = create.addResource(BaseUrlPaths.HEALTH)
    createHealth.addMethod("GET", createIntegration)

    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayEndPoints", { value: api.methods.join("\n") })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
  }
}
