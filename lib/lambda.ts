import type { IBucket } from "aws-cdk-lib/aws-s3";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Duration } from 'aws-cdk-lib';
import { Construct } from "constructs";

export function createLambda(
	th: Construct,
	funcName: string,
	pathStr: string,
	bucket: IBucket,
	environment: Record<string, string>,
	description?: string,
): lambda.Function {
	return new lambda.Function(th, funcName, {
		code: lambda.Code.fromBucket(
			bucket,
			pathStr,
		),
		runtime: lambda.Runtime.GO_1_X,
		handler: "main",
		functionName: funcName,
		description: description ?? `Lambda function for ${funcName}`,
		tracing: lambda.Tracing.ACTIVE,
		timeout: Duration.seconds(120),
		memorySize: 256,
		environment
	})
}
