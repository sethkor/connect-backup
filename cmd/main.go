package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/sethkor/connect-backup"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app       = kingpin.New("connect-backup", "A tool to backup and restore your AWS Connect instance")
	pProfile  = app.Flag("profile", "AWS credentials/config file profile to use").String()
	pRegion   = app.Flag("region", "AWS region").String()
	pInstance = app.Flag("instance", "The AWS Connect instance id to backup").Required().String()

	pBackupCommand = app.Command("backup", "backup your instance")
	pFile          = pBackupCommand.Flag("file", "write output to file with the provided path").ExistingDir()
	pS3            = pBackupCommand.Flag("s3", "write file to S3 destination with path as a url").URL()
)

var (
	version = "dev-local-version"
	commit  = "none"
	date    = "unknown"
)

func getAwsSession() *session.Session {
	var sess *session.Session
	if *pProfile != "" {

		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Profile:           *pProfile,
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
				MaxRetries:                    aws.Int(30),
			},
		}))

	} else {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
				MaxRetries:                    aws.Int(30),
			},
		}))
	} //else

	if *pRegion != "" {
		sess.Config.Region = aws.String(*pRegion)
	}
	return sess
}

func main() {

	app.Version(version + " " + date + " " + commit)
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	sess := getAwsSession()
	svc := connect.New(sess)

	switch command {
	case pBackupCommand.FullCommand():
		var theWriter connect_backup.Writer = &connect_backup.StdoutWriter{}
		if *pFile != "" {
			theWriter = &connect_backup.FileWriter{Path: *pFile}
			theWriter.(*connect_backup.FileWriter).InitDirs()
		} else if *pS3 != nil {
			theWriter = &connect_backup.S3Writer{Destination: *(*pS3), Sess: sess}
		}
		connect_backup.Backup(svc, *pInstance, theWriter)
	}

}
