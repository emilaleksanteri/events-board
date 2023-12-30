#!/bin/bash

awslocal lambda create-function \
    --function-name test \
    --handler main \
    --runtime go1.x \
    --zip-file fileb://app/main.zip \
    --timeout 240 \
    --role arn:aws:iam::000000000000:role/lambda-ri
