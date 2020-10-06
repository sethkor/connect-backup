#The contents of this file is used only for building and deploying a lambda via SAM.

#Some versioning vars, probably best not to fiddle with these
VERSION=$(shell git describe --abbrev=0 --exact-match --tags)
BRANCH=$(shell git branch | grep \* | cut -d ' ' -f2)
DATE=$(shell date)
COMMIT=$(shell git rev-parse HEAD)

#The name for the CFN stack to be deployed
STACK-NAME=connect-backup-lambda
#The aws profile to build your SAM package
BUILD-PROFILE=sethkor
#The aws profile to deploy the sam package.  This can be different to the BUILD-PROFILE if you want
DEPLOY-PROFILE=sethkor
EXE-NAME=connect-backup-lambda-test
LAMBDA-NAME=ConnectBackup
REGION=ap-southeast-2
#This is your existing AWS sam deployment bucket.  It is not created as part of this project and I have assumed it
#already exists for you.  For more info, see the AWS SAM documentation
SAM-BUCKET=versent-sethkor-sam
#This is the bucket that will be created for your backups
BACKUP-BUCKET=sethkor-versent-connect-backup-test
#This is your AWS Connect instance id.  You can get this from the console
CONNECT-INSTANCE-ID=946d3719-32f0-4ccc-a3de-713f52f6db7f
