package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/sethkor/connect-backup"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app       = kingpin.New("connect-backup", "A tool to backup and restore your AWS Connect instance")
	pProfile  = app.Flag("profile", "AWS credentials/config file profile to use").String()
	pRegion   = app.Flag("region", "AWS region").String()
	pInstance = app.Flag("instance", "The AWS Connect instance id to backup").Required().String()

	pBackupCommand = app.Command("backup", "Backup your instance")
	pFile          = pBackupCommand.Flag("file", "Write output to file with the provided path").ExistingDir()
	pS3            = pBackupCommand.Flag("s3", "Write file to S3 destination with path as a url").URL()

	pRestoreCommand = app.Command("restore", "Restore a connect component")
	pType           = pRestoreCommand.Flag("type", "Type to restore.  must be one of flow,routing-profile,user,user-hierarchy-group,user-hierarchy-structure").Required().Enum(
		string(connect_backup.Flows),
		string(connect_backup.RoutingProfiles),
		string(connect_backup.Users),
		string(connect_backup.UserHierarchyGroups),
		string(connect_backup.UserHierarchyStructure))
	pCreate = pRestoreCommand.Flag("create", "Restore contact flow as a new created flow with new name instead of overwriting").String()
	pSource = pRestoreCommand.Arg("json", "Location of restoration json (s3 URL or file)").Required().String()

	pRenameFlowsCommand = app.Command("rename-flows", "Rename all demo call flows with a prefix.  Defaults to just the AWS Demo flows")
	pPrefix             = pRenameFlowsCommand.Flag("prefix", "Prefix to use").Default("~").String()
	pAllFlows           = pRenameFlowsCommand.Flag("all-flows", "Rename all flows").Default("false").Bool()
)

var (
	version = "dev-local-version"
	commit  = "none"
	date    = "unknown"
)

func main() {

	app.Version(version + " " + date + " " + commit)
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	var err error
	sess := connect_backup.GetAwsSession(*pProfile, *pRegion)

	switch command {
	case pBackupCommand.FullCommand():
		var theWriter connect_backup.Writer = &connect_backup.StdoutWriter{}
		if *pFile != "" {
			theWriter = &connect_backup.FileWriter{Path: *pFile}
			theWriter.(*connect_backup.FileWriter).InitDirs()
		} else if *pS3 != nil {
			theWriter = &connect_backup.S3Writer{Destination: *(*pS3), Sess: sess}
		}
		cb := connect_backup.ConnectBackup{
			ConnectInstanceId: pInstance,
			TheWriter:         theWriter,
			Svc:               connect.New(sess),
		}
		err = cb.Backup()

	case pRestoreCommand.FullCommand():

		cr := connect_backup.ConnectRestore{
			ConnectInstanceId: pInstance,
			Session:           *sess,
			Source:            *pSource,
			Element:           connect_backup.ConnectElement(*pType),
			NewName:           *pCreate,
		}
		err = cr.Restore()

	case pRenameFlowsCommand.FullCommand():
		cb := connect_backup.ConnectBackup{
			ConnectInstanceId: pInstance,
			Svc:               connect.New(sess),
		}

		err = cb.RenameFlows(*pPrefix, *pAllFlows)
	default:
		app.FatalUsage("")
	}

	if err != nil {
		log.Fatal(err)
	}

}
