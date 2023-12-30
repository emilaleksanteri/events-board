#!/bin/bash

awslocal lambda invoke \
    --function-name test test.lambda.log
