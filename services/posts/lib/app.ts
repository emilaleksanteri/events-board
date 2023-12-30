import * as cdk from 'aws-cdk-lib';
import { PostsStack } from './posts-stack';

const app = new cdk.App();
new PostsStack(app, 'PostsStack');

app.synth()
