import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { Comments } from "../services/comments/lib/comments-stack";
import { Posts } from "../services/posts/lib/posts-stack";
import { Notifications } from "../services/notifications/lib/notification";

export class PubSub extends Stack {
	constructor(scope: Construct, id: string, props?: StackProps) {
		super(scope, id, props);

		const regionsToReplicate = ["us-east-1", "us-west-2", "eu-west-1"];
		const region = this.region
		const account = this.account
		const db_url = process.env.DB_ADDRESS;
		new Posts(this, "PostsStack", { db_url: db_url });
		new Comments(this, "CommentsStack", { db_url: db_url, });
		new Notifications(
			this,
			"NotificationsStack",
			{ regionsToReplicate, region, account }
		);
	}
}
