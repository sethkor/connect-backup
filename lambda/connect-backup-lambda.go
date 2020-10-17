package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/sethkor/connect-backup"
)

type Response struct {
	Answer string `json:"Response" yaml:"Response"`
}

type Request struct {
	ConnectInstanceId string `json:"ConnectInstanceId" yaml:"ConnectInstanceId"`
	S3DestURL         string `json:"S3DestURL" yaml:"S3DestURL"`
}

func HandleRequest(ctx context.Context, backupRequest Request) (Response, error) {
	log.Print(backupRequest)

	sess := connect_backup.GetAwsSession("", "")
	svc := connect.New(sess)

	instanceId := backupRequest.ConnectInstanceId
	if instanceId == "" {
		instanceId = os.Getenv("CONNECT_INSTANCE_ID")
		if instanceId == "" {
			log.Fatalln("No Connect instance ID passed in either the event or as an environment variable")
		}
		log.Println("The ConnectInstanceId in the event was blank, using the environment var CONNECT_INSTANCE_ID")
	}
	log.Println("Connect Instance: " + instanceId)

	s3destination := backupRequest.S3DestURL
	if s3destination == "" {
		s3destination = os.Getenv("S3_DEST_URL")
		if s3destination == "" {
			log.Fatalln("No S3DestURL passed in either the event or as an environment variable")
		}
		log.Println("The S3DestURL in the event was blank, using the environment var S3_DEST_URL")
	}
	s3Url, err := url.Parse(s3destination)
	log.Println("S3 URL: " + s3Url.String())

	if err != nil || s3Url.Scheme != "s3" {
		log.Println("There was an error parsing the S3 URL")
		return Response{Answer: "The S3 URL passed in the environment variable S3-DEST-URL was malformed"}, err
	}

	cb := connect_backup.ConnectBackup{ConnectInstanceId: &instanceId,
		Svc:       svc,
		TheWriter: &connect_backup.S3Writer{Destination: *s3Url, Sess: sess},
	}
	err = cb.Backup()

	if err != nil {
		log.Println("There was an error performing the backup")
		return Response{Answer: "There was an error performing the backup"}, err
	}

	return Response{Answer: "Processing Successful"}, nil

}

func main() {
	lambda.Start(HandleRequest)
}
