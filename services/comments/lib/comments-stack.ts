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

    const api = new RestApi(this, "commentsApi", {
      restApiName: "commentsApi",
      description: "API for comments",
    })
    Tags.of(api).add("_custom_id_", "commentsApi")

    const postCommentLambda = createLambda(
      this,
      "PostCommentLambda",
      "../lambdas/postComment",
      hotReloadBucket,
      db_url,
    )

    // COMMENT (POST)
    const postCommentIntegration = new LambdaIntegration(postCommentLambda)
    const create = api.root.addResource(BaseUrlPaths.CREATE_COMMENT)
    create.addMethod("POST", postCommentIntegration)

    const createSubCommetn = create.addResource(BaseUrlPaths.BY_ID)
    createSubCommetn.addMethod("POST", postCommentIntegration)

    const createHealth = create.addResource(BaseUrlPaths.HEALTH)
    createHealth.addMethod("GET", postCommentIntegration)


    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
  }
}
