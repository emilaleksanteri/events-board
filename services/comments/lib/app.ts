import * as cdk from 'aws-cdk-lib';
import { CommentsStack } from './comments-stack';

const app = new cdk.App();
new CommentsStack(app, 'CommentsStack');

app.synth()
