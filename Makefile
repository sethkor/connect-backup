include env.mk
.PHONY: lambda

mac:
	env GOOS=darwin GOARCH=amd64 go build -o connect-backup cmd/main.go

linux:
	env GOOS=linux GOARCH=amd64 go build -o connect-backup.linux cmd/main.go
	zip -qu9 lexbelt.linux.zip ./lexbelt.linux

clean:
	rm connect-backup.linux connect-backup.linux.gz connect-backup connect-backup-lambda connect-backup-lambda.linux connect-backup-lambda.linux.gz

publish-test:
	goreleaser --snapshot --skip-publish --rm-dist

publish:
	goreleaser --rm-dist --skip-validate

lambda:
	env GOOS=linux GOARCH=amd64 go build -o connect-backup-lambda lambda/connect-backup-lambda.go
	zip -qu9 connect-backup-lambda.zip ./connect-backup-lambda

sam-package:
	sam package --profile $(BUILD-PROFILE) --region $(REGION) --template-file lambda/template.yaml --s3-bucket $(SAM-BUCKET) \
	--output-template-file packaged.yaml

sam-deploy: sam-package
	sam deploy --profile $(DEPLOY-PROFILE) --region $(REGION) --template-file packaged.yaml --stack-name $(STACK-NAME) --capabilities CAPABILITY_IAM \
	--parameter-overrides ParameterKey=connectInstanceId,ParameterValue=$(CONNECT-INSTANCE-ID) \
	ParameterKey=bucketName,ParameterValue=$(BACKUP-BUCKET)

sam-remove:
	aws --profile $(DEPLOY-PROFILE) --region $(REGION) cloudformation delete-stack  --stack-name $(STACK-NAME)
	aws --profile $(DEPLOY-PROFILE) --region $(REGION) cloudformation wait stack-delete-complete --stack-name $(STACK-NAME)

