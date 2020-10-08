package connect_backup

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws/awsutil"

	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"

	"github.com/aws/aws-sdk-go/service/connect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type connectElement int

const (
	Flow connectElement = 1
)

type ConnectRestore struct {
	ConnectInstanceId *string
	Session           session.Session
	Source            string
	Element           connectElement
	NewFlowName       string
}

func (cr ConnectRestore) Restore() error {

	switch cr.Element {
	case Flow:
		return cr.restoreFlow()
	default:
		return errors.New("only restoration of contact flows is supported for now")
	}
}

func (cr ConnectRestore) restoreFlow() error {

	//is the location S3 or file?
	var theFlow connect.ContactFlow

	s3Location, err := url.Parse(cr.Source)
	if s3Location.Scheme == "s3" {
		var stream io.ReadCloser
		s3Svc := s3.New(&cr.Session)

		result, err := s3Svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(s3Location.Host),
			Key:    aws.String(s3Location.Path),
		})

		if err != nil {
			return err
		}
		stream = result.Body
		err = jsonutil.UnmarshalJSON(&theFlow, stream)

	} else {
		//Assume it's a file, try opening it
		byte, err := ioutil.ReadFile(cr.Source)
		if err != nil {
			return err
		}
		err = json.Unmarshal(byte, &theFlow)
	}

	if err != nil {
		return err
	}

	connectSvc := connect.New(&cr.Session)

	//if we have a new flow name then we are creating a new flow with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewFlowName != "" {
		var newFlow connect.CreateContactFlowInput
		awsutil.Copy(&newFlow, &theFlow)
		newFlow.Name = aws.String(cr.NewFlowName)
		newFlow.InstanceId = cr.ConnectInstanceId

		newFlow.Tags["restored-by"] = aws.String("https://github.com/sethkor/connect-backup")
		newFlow.Tags["restored-date"] = aws.String(time.Now().UTC().String())

		_, err = connectSvc.CreateContactFlow(&newFlow)

	} else {

		_, err = connectSvc.UpdateContactFlowContent(&connect.UpdateContactFlowContentInput{
			ContactFlowId: theFlow.Id,
			Content:       theFlow.Content,
			InstanceId:    cr.ConnectInstanceId,
		})
	}

	if err != nil {
		return err
	}

	return err
}
