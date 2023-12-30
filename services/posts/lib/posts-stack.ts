import {
  CfnOutput, Stack, StackProps
} from 'aws-cdk-lib';
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

import {
  RestApi,
  LambdaIntegration,
} from "aws-cdk-lib/aws-apigateway";


export class PostsStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const lambdaFunc = new lambda.Function(this, "postsFunc", {
      code: lambda.Code.fromAsset("app"),
      runtime: lambda.Runtime.GO_1_X,
      handler: "main",
      functionName: "postsFunc",
      description: "Posts function",
      tracing: lambda.Tracing.ACTIVE,
    })

    const api = new RestApi(this, "postsApi");

    const integration = new LambdaIntegration(lambdaFunc)
    api.root.addMethod("GET", integration)

    const health = api.root.addResource("health")
    health.addMethod("GET", integration)

    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
    new CfnOutput(this, "LambdaArn", { value: lambdaFunc.functionArn })
    new CfnOutput(this, "LambdaName", { value: lambdaFunc.functionName })
    new CfnOutput(this, "LambdaVersion", { value: lambdaFunc.currentVersion.version })

  }
}
