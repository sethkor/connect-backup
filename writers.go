package connect_backup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"

	"github.com/aws/aws-sdk-go/service/connect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Writer interface {
	write(result interface{}) error
	writeList(name string, result interface{}) error
	writeFlowString(name string, flow string) error
	init(instance string) error
}

type FileWriter struct {
	BasePath string
	//	path      string
	BaseWriter
}

type S3Writer struct {
	Destination url.URL
	//	path        string
	Sess *session.Session
	BaseWriter
}

type StdoutWriter struct {
	BaseWriter
}

type BaseWriter struct {
	path      string
	separator string
}

const (
	common = "common"
	//unknown       = "unknown"
	//pathSeparator = string(os.PathSeparator)
	jsonExtn = ".json"
)

func buildPrefix(separator string, result interface{}) (string, error) {

	var objectPrefix string

	switch result.(type) {
	case connect.ContactFlow:
		objectPrefix = string(Flows) + separator + *result.(connect.ContactFlow).Name + jsonExtn
	case connect.RoutingProfile:
		objectPrefix = string(RoutingProfiles) + separator + *result.(connect.RoutingProfile).Name + jsonExtn
	case backupRoutingProfileQueueSummary:
		objectPrefix = string(RoutingProfileQueues) + separator + result.(backupRoutingProfileQueueSummary).routingProfile + jsonExtn
	case connect.User:
		objectPrefix = string(Users) + separator + *result.(connect.User).Username + jsonExtn
	case connect.HierarchyGroup:
		objectPrefix = string(UserHierarchyGroups) + separator + *result.(connect.HierarchyGroup).Name + jsonExtn
	case connect.HoursOfOperation:
		objectPrefix = string(HoursOfOperation) + separator + *result.(connect.HoursOfOperation).Name + jsonExtn
	case []*connect.QuickConnectSummary:
		objectPrefix = string(QuickConnects) + separator + string(QuickConnects) + jsonExtn
	case connect.HierarchyStructure:
		objectPrefix = common + separator + string(UserHierarchyStructure) + jsonExtn
	case connect.Queue:
		objectPrefix = string(Queues) + separator + *result.(connect.Queue).Name + jsonExtn
	case connect.Instance:
		objectPrefix = common + separator + string(Instance) + jsonExtn
	case lambdaStrings:
		objectPrefix = common + separator + string(Lambdas) + jsonExtn
	default:
		return "", errors.New("unexpected type passed to writer")
	}
	return objectPrefix, nil
}

func buildPrefixList(name string, separator string, result interface{}) (string, error) {

	var objectPrefix string

	switch result.(type) {
	case []*connect.RoutingProfileQueueConfigSummary:
		objectPrefix = string(RoutingProfileQueues) + separator + name + jsonExtn
	case []*connect.PromptSummary:
		objectPrefix = string(Prompts) + separator + name + jsonExtn
	default:
		return "", errors.New("unexpected type passed to writer")
	}
	return objectPrefix, nil
}

func prettyJSON(flow string) (bytes.Buffer, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(flow), "", "  ")
	return prettyJSON, err
}

func (fw *FileWriter) init(instance string) error {
	//ensure the needed child dirs are present
	fw.separator = string(os.PathSeparator)
	fw.path = fw.BasePath + fw.separator + instance + fw.separator
	err := os.MkdirAll(fw.path+string(Flows), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(FlowsRaw), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(RoutingProfiles), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(RoutingProfileQueues), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(Users), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(UserHierarchyGroups), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(Prompts), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(HoursOfOperation), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(QuickConnects), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+string(Queues), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fw.path+common, 0744)
	return err
}

func (s3w *S3Writer) init(instance string) error {
	s3w.separator = "/"
	s3w.path = s3w.Destination.Path + s3w.separator + instance + s3w.separator

	return nil
}

func (fw *StdoutWriter) init(instance string) error {
	fw.path = string(os.PathSeparator)
	return nil
}

//As some AWS connect elements are listed and don't have unique ids, we need to sometimes pass a name around

func (fw *FileWriter) writeRawFile(fileName string, result interface{}) error {

	json, err := jsonutil.BuildJSON(result)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, json, 0644)
}

func (fw *FileWriter) writeList(name string, result interface{}) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix, err := buildPrefixList(name, fw.separator, result)

	if err != nil {
		return err
	}

	return fw.writeRawFile(fw.path+filePrefix, result)
}

func (fw *FileWriter) write(result interface{}) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix, err := buildPrefix(fw.separator, result)

	if err != nil {
		return err
	}

	return fw.writeRawFile(fw.path+filePrefix, result)
}

func (fw *FileWriter) writeFlowString(fileName string, flow string) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix := string(FlowsRaw) + fw.separator + fileName + jsonExtn

	prettyString, err := prettyJSON(flow)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fw.path+fw.separator+filePrefix, prettyString.Bytes(), 0644)
}

func (s3w *S3Writer) writeRawObj(objectPrefix string, result interface{}) error {
	if s3w.Destination.Scheme != "s3" {
		return errors.New("URL passes is not for S3")
	}

	svc := s3.New(s3w.Sess)

	json, err := jsonutil.BuildJSON(result)

	if err != nil {
		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Bucket: aws.String(s3w.Destination.Host),
		Body:   bytes.NewReader(json),
		Key:    aws.String(s3w.path + s3w.separator + objectPrefix),
	})

	return err
}

func (s3w *S3Writer) write(result interface{}) error {
	objectPrefix, err := buildPrefix(s3w.separator, result)

	if err != nil {
		return err
	}

	return s3w.writeRawObj(objectPrefix, result)
}

func (s3w *S3Writer) writeList(name string, result interface{}) error {
	objectPrefix, err := buildPrefixList(name, s3w.separator, result)

	if err != nil {
		return err
	}

	return s3w.writeRawObj(objectPrefix, result)
}

func (s3w *S3Writer) writeFlowString(fileName string, flow string) error {
	//for each contact flow, write to a file with the name contact flow.

	objectPrefix := string(FlowsRaw) + s3w.separator + fileName + jsonExtn

	if s3w.Destination.Scheme != "s3" {
		return errors.New("URL passes is not for S3")
	}

	prettyString, err := prettyJSON(flow)
	if err != nil {
		return err
	}

	svc := s3.New(s3w.Sess)

	_, err = svc.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Bucket: aws.String(s3w.Destination.Host),
		Body:   bytes.NewReader(prettyString.Bytes()),
		Key:    aws.String(s3w.path + "/" + objectPrefix),
	})

	return err
}

func (*StdoutWriter) write(result interface{}) error {

	json, err := jsonutil.BuildJSON(result)

	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}

func (*StdoutWriter) writeList(_ string, result interface{}) error {

	json, err := jsonutil.BuildJSON(result)

	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}

func (fw *StdoutWriter) writeFlowString(_ string, flow string) error {

	prettyString, err := prettyJSON(flow)
	if err == nil {
		fmt.Println(prettyString.String())
	}

	return err
}
