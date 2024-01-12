import { CfnOutput, Tags } from 'aws-cdk-lib';
import { Construct } from "constructs";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";
import { Bucket } from 'aws-cdk-lib/aws-s3'
import { createLambda } from '../../../lib/lambda';
import * as path from "path"
import * as events from 'aws-cdk-lib/aws-events';

enum BaseUrlPaths {
  HEALTH = "healthcheck",
  BY_ID = "{id}",
  CREATE_COMMENT = "create",
  COMMENTS = "comments",
  UPDATE_COMMENT = "update",
  DELETE_COMMENT = "delete",
}

interface CommentsProps {
  db_url?: string
  eventBus: events.EventBus
}

export class Comments extends Construct {
  constructor(scope: Construct, id: string, props: CommentsProps) {
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

    const postCommentLambda = createLambda(
      this,
      "PostCommentLambda",
      path.join(__dirname, "../lambdas/postComment"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url, BUS_NAME: eventBus.eventBusName },
    )
    eventBus.grantPutEventsTo(postCommentLambda)

    const getCommentLambda = createLambda(
      this,
      "GetCommentLambda",
      path.join(__dirname, "../lambdas/getComment"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },
    )

    const updateCommentLambda = createLambda(
      this,
      "UpdateCommentLambda",
      path.join(__dirname, "../lambdas/updateComment"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },
    )

    const deleteCommentLambda = createLambda(
      this,
      "DeleteCommentLambda",
      path.join(__dirname, "../lambdas/deleteComment"),
      hotReloadBucket,
      { DB_ADDRESS: props.db_url },
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

    // COMMENT (DELETE)
    const deleteCommentIntegration = new LambdaIntegration(deleteCommentLambda)
    const remove = api.root.addResource(BaseUrlPaths.DELETE_COMMENT)

    const deleteComment = remove.addResource(BaseUrlPaths.BY_ID)
    deleteComment.addMethod("DELETE", deleteCommentIntegration)

    const deleteHealth = remove.addResource(BaseUrlPaths.HEALTH)
    deleteHealth.addMethod("GET", deleteCommentIntegration)

    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
    new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
  }
}
