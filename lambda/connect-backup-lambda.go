package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"strconv"

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
	FlowsRaw          *bool  `json:"FlowsRaw" yaml:"FlowsRaw"`
}

func HandleRequest(ctx context.Context, backupRequest Request) (Response, error) {
	log.Print(backupRequest)

	sess := connect_backup.GetAwsSession("", "")
	svc := connect.New(sess)

	instanceId := backupRequest.ConnectInstanceId
	if instanceId == "" {
		instanceId = os.Getenv("CONNECT_INSTANCE_ID")
		if instanceId == "" {
			log.Fatalln("No Connect instance ID passed in either the event or as an environment variable CONNECT_INSTANCE_ID")
		}
		log.Println("The ConnectInstanceId in the event was blank, using the environment var CONNECT_INSTANCE_ID")
	}
	log.Println("Connect Instance: " + instanceId)

	s3destination := backupRequest.S3DestURL
	if s3destination == "" {
		s3destination = os.Getenv("S3_DEST_URL")
		if s3destination == "" {
			log.Fatalln("No S3DestURL passed in either the event or as an environment variable S3_DEST_URL")
		}
		log.Println("The S3DestURL in the event was blank, using the environment var S3_DEST_URL")
	}

	s3Url, err := url.Parse(s3destination)
	log.Println("S3 URL: " + s3Url.String())

	if err != nil || s3Url.Scheme != "s3" {
		log.Println("There was an error parsing the S3 URL")
		return Response{Answer: "The S3 URL passed in the environment variable S3_DEST_URL was malformed"}, err
	}

	var flowsRaw = false
	if backupRequest.FlowsRaw != nil {
		flowsRaw = *backupRequest.FlowsRaw
	} else {
		log.Println("The FlowsRaw in the event was blank, using the environment var FLOWS_RAW")

		flowsRawEnvString := os.Getenv("FLOWS_RAW")
		if flowsRawEnvString == "" {
			log.Println("No FlowsRaw passed in either the event or as an environment variable FLOWS_RAW, I am setting this to false and continuing")
		} else {
			var err error
			flowsRaw, err = strconv.ParseBool(flowsRawEnvString)
			if err != nil {
				log.Println("The FLOWS_RAW env variable was not a bool that can be parsed, I am setting this to false and continuing")
				flowsRaw = false
			}
		}
	}
	log.Println("FlowsRaw : " + strconv.FormatBool(flowsRaw))

	connectSvc := connect.New(sess)
	result, err := connectSvc.DescribeInstance(&connect.DescribeInstanceInput{
		InstanceId: &instanceId,
	})

	cb := connect_backup.ConnectBackup{ConnectInstance: *result.Instance,
		Svc:       svc,
		TheWriter: &connect_backup.S3Writer{Destination: *s3Url, Sess: sess},
		RawFlow:   flowsRaw,
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
