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
	write(result connect.ContactFlow) error
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

func (fw *FileWriter) write(result connect.ContactFlow) error {
	//for each contact flow, write to a file with the name contact flow.

	ioutil.WriteFile(fw.Path+string(os.PathSeparator)+*result.Name, []byte(*result.Content), 0644)
	return nil
}

func (s3w *S3Writer) write(result connect.ContactFlow) error {
	if s3w.Destination.Scheme != "s3" {
		return errors.New("URL passes is not for S3")
	}

	svc := s3.New(s3w.Sess)

	_, err := svc.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Bucket: aws.String(s3w.Destination.Host),
		Body:   bytes.NewReader([]byte(result.String())),
		Key:    aws.String(s3w.Destination.Path + *result.Name),
	})

	if err != nil {
		return err
	}
	return nil
}

func (*StdoutWriter) write(result connect.ContactFlow) error {

	fmt.Println(result)
	return nil
}
