import {
  CfnOutput, Stack, StackProps,
} from 'aws-cdk-lib';
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

import {
  RestApi,
  LambdaIntegration,
  EndpointType,
  MethodLoggingLevel,
} from "aws-cdk-lib/aws-apigateway";


export class PostsStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const lambdaFunc = new lambda.Function(this, "postsFunc", {
      code: lambda.Code.fromAsset("./app/bin"),
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: "main",
    })

    const api = new RestApi(this, "postsApi", {
      defaultCorsPreflightOptions: {
        allowHeaders: [
          "Content-Type",
          "X-Amz-Date",
          "Authorization",
          "X-Api-Key",
        ],
        allowMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allowCredentials: true,
        allowOrigins: ["*"],
      },
      deployOptions: {
        loggingLevel: MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
      },
      endpointConfiguration: {
        types: [EndpointType.REGIONAL],
      }
    });

    const integration = new LambdaIntegration(lambdaFunc)
    api.root.addMethod("GET", integration)


    new CfnOutput(this, "GatewayId", { value: api.restApiId })
    new CfnOutput(this, "GatewayUrl", { value: api.url })
    new CfnOutput(this, "LambdaArn", { value: lambdaFunc.functionArn })
    new CfnOutput(this, "LambdaName", { value: lambdaFunc.functionName })
    new CfnOutput(this, "LambdaVersion", { value: lambdaFunc.currentVersion.version })

  }
}
