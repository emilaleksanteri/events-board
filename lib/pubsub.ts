import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { Comments } from "../services/comments/lib/comments-stack";
import { Posts } from "../services/posts/lib/posts-stack";
import { Notifications } from "../services/notifications/lib/notification";
import { Likes } from "../services/likes/lib/likes";
import * as events from 'aws-cdk-lib/aws-events';

export class PubSub extends Stack {
	constructor(scope: Construct, id: string, props?: StackProps) {
		super(scope, id, props);

		const regionsToReplicate = ["us-east-1", "us-west-2"];
		const region = this.region
		const account = this.account
		const db_url = process.env.DB_ADDRESS;
		const isProd = process.env.IS_PROD === "true";

		const eventBus = new events.EventBus(this, "NotificationsEventBus", {
			eventBusName: "notifications",
		})

		new Posts(this, "PostsStack", { db_url: db_url, eventBus });
		new Comments(this, "CommentsStack", { db_url: db_url, eventBus });
		/**
		new Notifications(
			this,
			"NotificationsStack",
			{ regionsToReplicate, region, account, isProd, db_url, eventBus }
		);
		*/
		new Likes(this, "LikesStack", { db_url: db_url, eventBus });
	}
}
