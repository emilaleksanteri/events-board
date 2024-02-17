import { CfnOutput, Tags } from 'aws-cdk-lib';
import { Construct } from "constructs";
import { RestApi, LambdaIntegration } from "aws-cdk-lib/aws-apigateway";
import { Bucket } from 'aws-cdk-lib/aws-s3';
import * as events from 'aws-cdk-lib/aws-events';
import { createLambda } from '../../../lib/lambda';
import * as path from "path"



interface SocialProps {
	db_url?: string
	eventBus: events.EventBus
}

export class Social extends Construct {
	constructor(scope: Construct, id: string, props: SocialProps) {
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


		const follow = createLambda(
			this,
			"follow",
			path.join(__dirname, "../lambdas/follow"),
			hotReloadBucket,
			{ DB_ADDRESS: props.db_url },
		)

		eventBus.grantPutEventsTo(follow)

		const api = new RestApi(this, "likesapi", {
			restApiName: "likesapi",
			description: "API for likes",
		})
		Tags.of(api).add("_custom_id_", "likesapi")



		new CfnOutput(this, "GatewayId", { value: api.restApiId })
		new CfnOutput(this, "GatewayUrl", { value: api.url })
		new CfnOutput(this, "GatewayEndPoints", { value: "\n" + api.methods.join("\n") })
	}
}
