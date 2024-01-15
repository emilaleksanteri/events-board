import { CfnOutput, Tags } from 'aws-cdk-lib';
import { Construct } from "constructs";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";
import { Bucket } from 'aws-cdk-lib/aws-s3';
import * as events from 'aws-cdk-lib/aws-events';
import { createLambda } from '../../../lib/lambda';
import * as path from "path"


enum LikeRoute {
	BASE = "like",
	POST = "post",
	COMMENT = "comment",
	ID = "{id}",
}

interface LikesProps {
	db_url?: string
	eventBus: events.EventBus
}

export class Likes extends Construct {
	constructor(scope: Construct, id: string, props: LikesProps) {
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


		const api = new RestApi(this, "likesapi", {
			restApiName: "likesapi",
			description: "API for likes",
		})
		Tags.of(api).add("_custom_id_", "likesapi")

		const likeCreateIntegration = new LambdaIntegration(postLikes)

		const base = api.root.addResource(LikeRoute.BASE)
		const posts = base.addResource(LikeRoute.POST)
		const post = posts.addResource(LikeRoute.ID)

		post.addMethod("POST", likeCreateIntegration)

		const comments = base.addResource(LikeRoute.COMMENT)
		const comment = comments.addResource(LikeRoute.ID)
		comment.addMethod("POST", likeCreateIntegration)

		new CfnOutput(this, "GatewayId", { value: api.restApiId })
		new CfnOutput(this, "GatewayUrl", { value: api.url })
		new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
	}
}
