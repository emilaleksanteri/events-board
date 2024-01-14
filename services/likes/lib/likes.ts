import { CfnOutput, Tags } from 'aws-cdk-lib';
import { Construct } from "constructs";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";
import { Bucket } from 'aws-cdk-lib/aws-s3';
import * as events from 'aws-cdk-lib/aws-events';
import { createLambda } from '../../../lib/lambda';
import * as path from "path"


enum LikeRoute {
	POST_LIKE = "like/post/{id}",
	COMMENT_LIKE = "like/comment/{id}",
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

		const postLikes = createLambda(
			this,
			"postLikes",
			path.join(__dirname, "../lambdas/postLike"),
			hotReloadBucket,
			{ DB_ADDRESS: props.db_url, BUS_NAME: eventBus.eventBusName },
		)
		eventBus.grantPutEventsTo(postLikes)


		const api = new RestApi(this, "likesApi", {
			restApiName: "likesApi",
			description: "API for likes",
		})
		Tags.of(api).add("_custom_id_", "likesApi")

		const likeCreateIntegration = new LambdaIntegration(postLikes)

		const postLike = api.root.addResource(LikeRoute.POST_LIKE)
		postLike.addMethod("POST", likeCreateIntegration)

		const commentLike = api.root.addResource(LikeRoute.COMMENT_LIKE)
		commentLike.addMethod("POST", likeCreateIntegration)
	}
}
