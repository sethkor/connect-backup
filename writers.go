package connect_backup

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/service/connect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Writer interface {
	write(result interface{}) error
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
	flows                  = "flows"
	routingProfiles        = "routing-profiles"
	user                   = "users"
	userHierarchyGroup     = "user-hierarchy-groups"
	userHierarchyStructure = "user-hierarchy-structures"
	common                 = "common"
	unknown                = "unknown"
	pathSeparator          = string(os.PathSeparator)
	jsonExtn               = ".json"
)

func getElement(result interface{}) (string, string, error) {

	var objectPrefix, element string

	switch result.(type) {
	case connect.ContactFlow:
		objectPrefix = flows + "/" + *result.(connect.ContactFlow).Name + jsonExtn
		element = result.(connect.ContactFlow).String()
	case connect.RoutingProfile:
		objectPrefix = routingProfiles + "/" + *result.(connect.RoutingProfile).Name + jsonExtn
		element = result.(connect.RoutingProfile).String()
	case connect.User:
		objectPrefix = user + "/" + *result.(connect.User).Username + jsonExtn
		element = result.(connect.User).String()
	case connect.HierarchyGroup:
		objectPrefix = userHierarchyGroup + "/" + *result.(connect.HierarchyGroup).Name + jsonExtn
		element = result.(connect.HierarchyGroup).String()
	case connect.HierarchyStructure:
		objectPrefix = common + "/" + userHierarchyStructure + jsonExtn
		element = result.(connect.HierarchyStructure).String()
	default:
		return "", "", errors.New("unexpected type passed to writer")
	}
	return objectPrefix, element, nil
}

func (fw *FileWriter) InitDirs() {
	//ensure the needed child dirs are present
	os.Mkdir(fw.Path+pathSeparator+flows, 0744)
	os.Mkdir(fw.Path+pathSeparator+routingProfiles, 0744)
	os.Mkdir(fw.Path+pathSeparator+user, 0744)
	os.Mkdir(fw.Path+pathSeparator+userHierarchyGroup, 0744)
	os.Mkdir(fw.Path+pathSeparator+common, 0744)
}

func (fw *FileWriter) write(result interface{}) error {
	//for each contact flow, write to a file with the name contact flow.

	filePrefix, element, err := getElement(result)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(fw.Path+string(os.PathSeparator)+filePrefix, []byte(element), 0644)
}

func (s3w *S3Writer) write(result interface{}) error {
	if s3w.Destination.Scheme != "s3" {
		return errors.New("URL passes is not for S3")
	}

	svc := s3.New(s3w.Sess)

	objectPrefix, element, err := getElement(result)

	if err != nil {
		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Bucket: aws.String(s3w.Destination.Host),
		Body:   bytes.NewReader([]byte(element)),
		Key:    aws.String(s3w.Destination.Path + "/" + objectPrefix),
	})

	return err
}

func (*StdoutWriter) write(result interface{}) error {

	_, element, err := getElement(result)

	if err != nil {
		return err
	}
	fmt.Println(element)
	return nil
}
