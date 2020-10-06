package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/sethkor/connect-backup"
)

type Response struct {
	Answer string `json:"Response" yaml:"Response"`
}

func HandleRequest(ctx context.Context, connectEvent events.ConnectEvent) (Response, error) {

	sess := connect_backup.GetAwsSession("", "")
	svc := connect.New(sess)

	instanceId := os.Getenv("CONNECT_INSTANCE_ID")
	log.Println("Connect Instance: " + instanceId)
	s3destination, err := url.Parse(os.Getenv("S3_DEST_URL"))
	log.Println("S3 URL: " + s3destination.String())
	if err != nil {
		log.Println("There was an error parsing the S3 URL")
		return Response{Answer: "The S3 URL passed in the environment variable S3-DEST-URL was malformed"}, err
	}

	cb := connect_backup.ConnectBackup{ConnectInstanceId: &instanceId}
	err = cb.Backup(svc, &connect_backup.S3Writer{Destination: *s3destination, Sess: sess})

	if err != nil {
		log.Println("There was an error performing the backup")
		return Response{Answer: "There was an error performing the backup"}, err
	}

	return Response{Answer: "Processing Successful"}, nil

}

func main() {
	lambda.Start(HandleRequest)
}
