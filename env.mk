VERSION=$(shell git describe --abbrev=0 --exact-match --tags)
BRANCH=$(shell git branch | grep \* | cut -d ' ' -f2)
DATE=$(shell date)
COMMIT=$(shell git rev-parse HEAD)
STACK-NAME=connect-backup-lambda
BUILD-PROFILE=sethkor
DEPLOY-PROFILE=sethkor
EXE-NAME=connect-backup-lambda-test
LAMBDA-NAME=ConnectBackup
REGION=ap-southeast-2
#This is your existing AWS sam deployment bucket.  It is not created as part of this project and I have assumed it
#already exists for you.  For more info, see the AWS SAM documentation
SAM-BUCKET=your-sam-deployment-bucket
#This is the bucket that will be create for your backups
BACKUP-BUCKET=sethkor-connect-backup-test
#This is your AWS Connect instance id.  You can get this from the console
CONNECT-INSTANCE-ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
