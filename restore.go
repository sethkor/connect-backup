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

	"github.com/sethvargo/go-password/password"
)

type ConnectElement string

const (
	Flow                   ConnectElement = "flow"
	RoutingProfile         ConnectElement = "routing-profile"
	User                   ConnectElement = "user"
	UserHierarchyGroup     ConnectElement = "user-hierarchy-group"
	UserHierarchyStructure ConnectElement = "user-hierarchy-structure"
)

type ConnectRestore struct {
	ConnectInstanceId *string
	Session           session.Session
	Source            string
	Element           ConnectElement
	NewName           string
}

func (cr ConnectRestore) Restore() error {

	switch cr.Element {
	case Flow:
		return cr.restoreFlow()
	case RoutingProfile:
		return cr.restoreRoutingProfile()
	case User:
		return cr.restoreUser()

	default:
		return errors.New("only restoration of contact flows is supported for now")
	}
}

func (cr ConnectRestore) restoreUser() error {
	var theUser connect.User

	err := cr.readSource(&theUser)

	if err != nil {
		return err
	}

	connectSvc := connect.New(&cr.Session)

	//if we have a new user name then we are creating a new flow with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewName != "" {
		var newProfile connect.CreateUserInput
		awsutil.Copy(&newProfile, &theUser)
		newProfile.Username = aws.String(cr.NewName)
		newProfile.InstanceId = cr.ConnectInstanceId
		newProfile.DirectoryUserId = nil

		res, err := password.Generate(64, 10, 10, false, false)
		if err != nil {
			return err
		}
		newProfile.Password = aws.String(res)

		newProfile.Tags["restored-by"] = aws.String("https://github.com/sethkor/connect-backup")
		newProfile.Tags["restored-date"] = aws.String(time.Now().UTC().String())

		_, err = connectSvc.CreateUser(&newProfile)

	} else {
		//Update the existing flow in place, this requires several operations.

		//First update the profile name and description
		_, err = connectSvc.UpdateUserIdentityInfo(&connect.UpdateUserIdentityInfoInput{
			InstanceId:   cr.ConnectInstanceId,
			UserId:       theUser.Id,
			IdentityInfo: theUser.IdentityInfo,
		})

		if err != nil {
			return err
		}

		_, err = connectSvc.UpdateUserSecurityProfiles(&connect.UpdateUserSecurityProfilesInput{
			InstanceId:         cr.ConnectInstanceId,
			UserId:             theUser.Id,
			SecurityProfileIds: theUser.SecurityProfileIds,
		})

		if err != nil {
			return err
		}

		_, err = connectSvc.UpdateUserPhoneConfig(&connect.UpdateUserPhoneConfigInput{
			InstanceId:  cr.ConnectInstanceId,
			UserId:      theUser.Id,
			PhoneConfig: theUser.PhoneConfig,
		})

	}
	return err
}

func (cr ConnectRestore) restoreRoutingProfile() error {

	var theProfile connect.RoutingProfile

	err := cr.readSource(&theProfile)

	if err != nil {
		return err
	}

	connectSvc := connect.New(&cr.Session)

	//if we have a new flow name then we are creating a new routing profile with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewName != "" {
		var newProfile connect.CreateRoutingProfileInput
		awsutil.Copy(&newProfile, &theProfile)
		newProfile.Name = aws.String(cr.NewName)
		newProfile.InstanceId = cr.ConnectInstanceId

		newProfile.Tags["restored-by"] = aws.String("https://github.com/sethkor/connect-backup")
		newProfile.Tags["restored-date"] = aws.String(time.Now().UTC().String())

		_, err = connectSvc.CreateRoutingProfile(&newProfile)

	} else {
		//Update the existing flow in place, this requires several operations.

		//First update the profile name and description
		_, err = connectSvc.UpdateRoutingProfileName(&connect.UpdateRoutingProfileNameInput{
			RoutingProfileId: theProfile.RoutingProfileId,
			InstanceId:       cr.ConnectInstanceId,
			Name:             theProfile.Name,
			Description:      theProfile.Description,
		})

		if err != nil {
			return err
		}

		//Then the concurrency
		_, err = connectSvc.UpdateRoutingProfileConcurrency(&connect.UpdateRoutingProfileConcurrencyInput{
			RoutingProfileId:   theProfile.RoutingProfileId,
			InstanceId:         cr.ConnectInstanceId,
			MediaConcurrencies: theProfile.MediaConcurrencies,
		})

		if err != nil {
			return err
		}

		//Now the default outbound queue
		//First update the flow name
		_, err = connectSvc.UpdateRoutingProfileDefaultOutboundQueue(&connect.UpdateRoutingProfileDefaultOutboundQueueInput{
			RoutingProfileId:       theProfile.RoutingProfileId,
			InstanceId:             cr.ConnectInstanceId,
			DefaultOutboundQueueId: theProfile.DefaultOutboundQueueId,
		})

	}

	return err
}

func (cr ConnectRestore) readSource(destination interface{}) error {
	s3Location, _ := url.Parse(cr.Source)
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
		err = jsonutil.UnmarshalJSON(destination, stream)

	} else {
		//Assume it's a file, try opening it
		byte, err := ioutil.ReadFile(cr.Source)
		if err != nil {
			return err
		}
		err = json.Unmarshal(byte, destination)
	}
	return nil
}

func (cr ConnectRestore) restoreFlow() error {

	//is the location S3 or file?
	var theFlow connect.ContactFlow

	err := cr.readSource(&theFlow)

	if err != nil {
		return err
	}

	connectSvc := connect.New(&cr.Session)

	//if we have a new flow name then we are creating a new flow with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewName != "" {
		var newFlow connect.CreateContactFlowInput
		awsutil.Copy(&newFlow, &theFlow)
		newFlow.Name = aws.String(cr.NewName)
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
