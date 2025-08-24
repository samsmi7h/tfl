LAMBDA_NAME=tfl
LAMBDA_BINARY=bootstrap # MUST be named this for Lambda
LAMBDA_ZIP=tflLambda.zip
LAMBDA_ROLE=LambdaBasicExecutionRole

lambda/role/create:
	aws iam create-role \
	  --role-name ${LAMBDA_ROLE} \
	  --assume-role-policy-document file://aws/lambda-policy.json
	aws iam attach-role-policy \
	  --role-name ${LAMBDA_ROLE} \
	  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

lambda/role/get/arn:
	@aws iam get-role --role-name ${LAMBDA_ROLE} \
		--query Role.Arn --output text

lambda/build:
	GOOS=linux GOARCH=amd64 go build -o ${LAMBDA_BINARY} cmd/lambda/main.go
	zip ${LAMBDA_ZIP} ${LAMBDA_BINARY}

lambda/create:
	LAMBDA_ROLE_ARN="$$(make -s lambda/role/get/arn)"; \
 	aws lambda create-function \
 	  --function-name ${LAMBDA_NAME} \
 	  --runtime provided.al2 \
 	  --handler ${LAMBDA_ZIP} \
 	  --zip-file fileb://${LAMBDA_ZIP} \
 	  --role "$$LAMBDA_ROLE_ARN"

lambda/update:
	aws lambda update-function-code \
 	  --function-name ${LAMBDA_NAME} \
 	  --zip-file fileb://${LAMBDA_ZIP}


lambda/config:
	if [ -z "$$TFL_APP_KEY" ]; then echo "TFL_APP_KEY is not set"; exit 1; fi
	aws lambda update-function-configuration \
  		--function-name ${LAMBDA_NAME} \
  		--environment "Variables={TFL_APP_KEY=$$TFL_APP_KEY}"

lambda/public-url:
	aws lambda create-function-url-config \
	  --function-name ${LAMBDA_NAME} \
	  --auth-type NONE
	aws lambda add-permission \
	  --function-name ${LAMBDA_NAME} \
	  --action lambda:InvokeFunctionUrl \
	  --principal "*" \
	  --function-url-auth-type NONE \
	  --statement-id FunctionURLAllowPublic

