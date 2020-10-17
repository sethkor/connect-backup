package connect_backup

import (
	"bytes"
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
}

type FileWriter struct {
	Path string
}

type S3Writer struct {
	Destination url.URL
	Sess        *session.Session
}

type StdoutWriter struct {
}

const (
	//flows                  = "flows"
	//routingProfiles        = "routing-profiles"
	//routingProfilesQueues  = "routing-profile-queues"
	//user                   = "users"
	//userHierarchyGroup     = "user-hierarchy-groups"
	//userHierarchyStructure = "user-hierarchy-structures"
	common        = "common"
	unknown       = "unknown"
	pathSeparator = string(os.PathSeparator)
	jsonExtn      = ".json"
)

func buildPrefix(result interface{}) (string, error) {

	var objectPrefix string

	switch result.(type) {
	case connect.ContactFlow:
		objectPrefix = string(Flows) + "s/" + *result.(connect.ContactFlow).Name + jsonExtn
	case connect.RoutingProfile:
		objectPrefix = string(RoutingProfiles) + "s/" + *result.(connect.RoutingProfile).Name + jsonExtn
	case backupRoutingProfileQueueSummary:
		objectPrefix = string(RoutingProfileQueues) + "s/" + result.(backupRoutingProfileQueueSummary).routingProfile + jsonExtn
	case connect.User:
		objectPrefix = string(Users) + "s/" + *result.(connect.User).Username + jsonExtn
	case connect.HierarchyGroup:
		objectPrefix = string(UserHierarchyGroups) + "s/" + *result.(connect.HierarchyGroup).Name + jsonExtn
	case connect.HierarchyStructure:
		objectPrefix = common + "/" + string(UserHierarchyStructure) + jsonExtn
	default:
		return "", errors.New("unexpected type passed to writer")
	}
	return objectPrefix, nil
}

func buildPrefixList(name string, result interface{}) (string, error) {

	var objectPrefix string

	switch result.(type) {
	case []*connect.RoutingProfileQueueConfigSummary:
		objectPrefix = string(RoutingProfileQueues) + "s/" + name + jsonExtn
	default:
		return "", errors.New("unexpected type passed to writer")
	}
	return objectPrefix, nil
}

func (fw *FileWriter) InitDirs() {
	//ensure the needed child dirs are present
	os.Mkdir(fw.Path+pathSeparator+string(Flows)+"s", 0744)
	os.Mkdir(fw.Path+pathSeparator+string(RoutingProfiles)+"s", 0744)
	os.Mkdir(fw.Path+pathSeparator+string(RoutingProfileQueues)+"s", 0744)
	os.Mkdir(fw.Path+pathSeparator+string(Users)+"s", 0744)
	os.Mkdir(fw.Path+pathSeparator+string(UserHierarchyGroups)+"s", 0744)
	os.Mkdir(fw.Path+pathSeparator+common, 0744)

}

//As some AWS connect elements are listed and don't have unique ids, we need to sometimes pass a name around

func (fw *FileWriter) writeRawFile(fileName string, result interface{}) error {

	json, err := jsonutil.BuildJSON(result)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(fw.Path+string(os.PathSeparator)+fileName, json, 0644)
}

func (fw *FileWriter) writeList(name string, result interface{}) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix, err := buildPrefixList(name, result)

	if err != nil {
		return err
	}

	return fw.writeRawFile(filePrefix, result)
}

func (fw *FileWriter) write(result interface{}) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix, err := buildPrefix(result)

	if err != nil {
		return err
	}

	return fw.writeRawFile(filePrefix, result)
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
		Key:    aws.String(s3w.Destination.Path + "/" + objectPrefix),
	})

	return err
}

func (s3w *S3Writer) write(result interface{}) error {
	objectPrefix, err := buildPrefix(result)

	if err != nil {
		return err
	}

	return s3w.writeRawObj(objectPrefix, result)
}

func (s3w *S3Writer) writeList(name string, result interface{}) error {
	objectPrefix, err := buildPrefixList(name, result)

	if err != nil {
		return err
	}

	return s3w.writeRawObj(objectPrefix, result)
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
