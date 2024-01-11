"use strict"
import {
	CfnOutput, Duration
} from 'aws-cdk-lib';
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

import { WebSocketLambdaIntegration } from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import * as apigw2 from 'aws-cdk-lib/aws-apigatewayv2';
import { Bucket, IBucket } from 'aws-cdk-lib/aws-s3'
import * as path from "path"
import * as events from 'aws-cdk-lib/aws-events';
import { AttributeType, Table } from 'aws-cdk-lib/aws-dynamodb';
import { Effect, PolicyStatement, Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';
import { EventBus, LambdaFunction } from 'aws-cdk-lib/aws-events-targets';

type LambdaEnv = Record<string, string>

function createLambda(
	th: Construct,
	funcName: string,
	pathStr: string,
	bucket: IBucket,
	lambdaEnv: LambdaEnv,
	description?: string,
): lambda.Function {
	return new lambda.Function(th, funcName, {
		code: lambda.Code.fromBucket(
			bucket,
			path.join(__dirname, pathStr)
		),
		runtime: lambda.Runtime.GO_1_X,
		handler: "main",
		functionName: funcName,
		description: description ?? `Lambda function for ${funcName}`,
		tracing: lambda.Tracing.ACTIVE,
		environment: lambdaEnv,
		timeout: Duration.seconds(120),
	})
}

interface NotifciationsProps {
	regionsToReplicate: string[],
	region: string,
	account: string,
	isProd: boolean,
	db_url?: string,
	eventBus: events.EventBus
}

export class Notifications extends Construct {
	constructor(scope: Construct, id: string, props: NotifciationsProps) {
		super(scope, id);
		const { eventBus } = props
		if (!props.db_url) {
			throw new Error("DB URL must be provided")
		}

		const hotReloadBucket = Bucket.fromBucketName(
			this,
			"HotReloadingBucket",
			"hot-reload"
		)

		const table = new Table(this, "NotificationsTable", {
			tableName: "notifications",
			partitionKey: {
				name: "userId",
				type: AttributeType.NUMBER
			},
			sortKey: {
				name: "connectionId",
				type: AttributeType.STRING
			}
		})

		const connLambda = createLambda(
			this,
			"ConnectionHandler",
			"../lambdas/connectionHandler",
			hotReloadBucket,
			{ TABLE_NAME: table.tableName }
		)
		table.grantFullAccess(connLambda)


		const api = new apigw2.WebSocketApi(this, "NotificationsApi", {
			apiName: "NotificationsApi",
			description: "API for notifications",
			connectRouteOptions: {
				integration: new WebSocketLambdaIntegration(
					"conInt",
					connLambda
				)
			},
			disconnectRouteOptions: {
				integration: new WebSocketLambdaIntegration(
					"disconnInt",
					connLambda
				)
			},
		})

		const wsStage = new apigw2.WebSocketStage(this, "NotificationsStage", {
			webSocketApi: api,
			stageName: "notifications",
			autoDeploy: true
		})

		const allowConnectionManagementOnApiGatewayPolicy = new PolicyStatement({
			effect: Effect.ALLOW,
			resources: [
				`arn:aws:execute-api:${props.region}:${props.account}:${api.apiId}/${wsStage.stageName}/*`,
			],
			actions: ['execute-api:ManageConnections'],
		});

		const processLambda = createLambda(
			this,
			"ProcessHandler",
			"../lambdas/messageHandler",
			hotReloadBucket,
			{ TABLE_NAME: table.tableName, DB_ADDRESS: props.db_url }
		)

		processLambda.addToRolePolicy(allowConnectionManagementOnApiGatewayPolicy)


		let crossRegionalEventbusTargets: EventBus[] = []
		if (props.isProd) {
			const crossRegionEventRole = new Role(this, 'CrossRegionRole', {
				inlinePolicies: {},
				assumedBy: new ServicePrincipal('events.amazonaws.com'),
			});

			// Generate list of Event buses in other regions
			crossRegionalEventbusTargets = props.regionsToReplicate
				.map((regionCode) =>
					new EventBus(events.EventBus.fromEventBusArn(
						this,
						`WebsocketNotificationBus-${regionCode}`,
						`arn:aws:events:${regionCode}:${props.account}:event-bus/${eventBus.eventBusName}`,
					), {
						role: crossRegionEventRole,
					}));

		}
		new events.Rule(this, 'ProcessRequest', {
			eventBus,
			enabled: true,
			ruleName: 'ProcessNotificationReq',
			eventPattern: {
				detailType: ['NotificationReceived'],
				source: ['notifications'],
			},
			targets: [
				new LambdaFunction(processLambda),
				...crossRegionalEventbusTargets
			],
		});

		eventBus.grantPutEventsTo(processLambda)
		table.grantFullAccess(processLambda)


		new CfnOutput(this, 'bucketName', {
			value: wsStage.url,
			description: 'WebSocket API URL',
		});

		new CfnOutput(this, 'apiId', {
			value: wsStage.api.apiId,
			description: 'WebSocket API ID',
		});
	}
}
