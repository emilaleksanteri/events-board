import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { Comments } from "../services/comments/lib/comments-stack";
import { Posts } from "../services/posts/lib/posts-stack";

export class PubSub extends Stack {
	constructor(scope: Construct, id: string, props?: StackProps) {
		super(scope, id, props);

		const db_url = process.env.DB_ADDRESS;
		new Posts(this, "PostsStack", { db_url: db_url });
		new Comments(this, "CommentsStack", { db_url: db_url, });
	}
}
