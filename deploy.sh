#!/usr/bin/env bash

go build -o main
zip -r main.zip main php*
aws lambda update-function-code --function-name php-syntax-checker --zip fileb://main.zip