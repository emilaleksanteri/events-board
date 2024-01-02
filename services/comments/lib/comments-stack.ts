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
  BY_ID = "{id}",
  CREATE_COMMENT = "create",
  COMMENTS = "comments",
  UPDATE_COMMENT = "update",
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

export class CommentsStack extends Stack {
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

    const postCommentLambda = createLambda(
      this,
      "PostCommentLambda",
      "../lambdas/postComment",
      hotReloadBucket,
      db_url,
    )

    const getCommentLambda = createLambda(
      this,
      "GetCommentLambda",
      "../lambdas/getComment",
      hotReloadBucket,
      db_url,
    )

    const updateCommentLambda = createLambda(
      this,
      "UpdateCommentLambda",
      "../lambdas/updateComment",
      hotReloadBucket,
      db_url,
    )

    const api = new RestApi(this, "commentsapi", {
      restApiName: "commentsapi",
      description: "API for comments",
    })
    Tags.of(api).add("_custom_id_", "commentsapi")


    // COMMENT (POST)
    const postCommentIntegration = new LambdaIntegration(postCommentLambda)
    const create = api.root.addResource(BaseUrlPaths.CREATE_COMMENT)
    create.addMethod("POST", postCommentIntegration)

    const createSubCommetn = create.addResource(BaseUrlPaths.BY_ID)
    createSubCommetn.addMethod("POST", postCommentIntegration)

    const createHealth = create.addResource(BaseUrlPaths.HEALTH)
    createHealth.addMethod("GET", postCommentIntegration)

    // COMMENT (GET)
    const getCommentIntegration = new LambdaIntegration(getCommentLambda)
    const comments = api.root.addResource(BaseUrlPaths.COMMENTS)

    const getComment = comments.addResource(BaseUrlPaths.BY_ID)
    getComment.addMethod("GET", getCommentIntegration)

    const getHealth = comments.addResource(BaseUrlPaths.HEALTH)
    getHealth.addMethod("GET", getCommentIntegration)

    // COMMENT (PUT)
    const updateCommentIntegration = new LambdaIntegration(updateCommentLambda)
    const update = api.root.addResource(BaseUrlPaths.UPDATE_COMMENT)

    const updateComment = update.addResource(BaseUrlPaths.BY_ID)
    updateComment.addMethod("PUT", updateCommentIntegration)

    const updateHealth = update.addResource(BaseUrlPaths.HEALTH)
    updateHealth.addMethod("GET", updateCommentIntegration)

    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
    new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
  }
}
