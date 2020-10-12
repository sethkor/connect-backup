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

	pBackupCommand = app.Command("backup", "backup your instance")
	pFile          = pBackupCommand.Flag("file", "write output to file with the provided path").ExistingDir()
	pS3            = pBackupCommand.Flag("s3", "write file to S3 destination with path as a url").URL()

	pRestoreCommand = app.Command("restore", "Restore a connect component")
	pType           = pRestoreCommand.Flag("type", "type to restore.  must be one of flow,routing-profile,user,user-hierarchy-group,user-hierarchy-structure").Required().Enum(
		string(connect_backup.Flow),
		string(connect_backup.RoutingProfile),
		string(connect_backup.User),
		string(connect_backup.UserHierarchyGroup),
		string(connect_backup.UserHierarchyStructure))
	//pFlow = pRestoreCommand.Command("flow", "Restore a flow") //pRestoreCommand.Flag("flow", "Restore a contact flow").Default("false").Bool()
	//pUser                   = pRestoreCommand.Flag("user", "Restore a user").Default("false").Bool()
	//pUserHierarchyGroup     = pRestoreCommand.Flag("user-hierarchy-group", "Restore a user hierarchy group").Default("false").Bool()
	//pUserHierarchyStructure = pRestoreCommand.Flag("user-hierarchy-structure", "Restore the user hierarchy structure").Default("false").Bool()
	//pRoutingProfile         = pRestoreCommand.Flag("routing-profile", "Restore a routing profile").Default("false").Bool()

	pCreate = pRestoreCommand.Flag("create", "Restore contact flow as a new created flow with new name instead of overwriting").String()
	pSource = pRestoreCommand.Arg("json", "Location of restoration json (s3 URL or file)").Required().String()
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
	default:
		app.FatalUsage("")
	}

	if err != nil {
		log.Fatal(err)
	}

}
