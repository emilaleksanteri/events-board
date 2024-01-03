import * as cdk from 'aws-cdk-lib';
import { PubSub } from './pubsub';
const app = new cdk.App();
new PubSub(app, 'PubSub');
app.synth()
