package connect_backup

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/aws/arn"

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
	Flows                  ConnectElement = "flow"
	FlowsRaw               ConnectElement = "flows-raw"
	RoutingProfiles        ConnectElement = "routing-profile"
	RoutingProfileQueues   ConnectElement = "routing-profile-queue"
	Users                  ConnectElement = "user"
	UserHierarchyGroups    ConnectElement = "user-hierarchy-group"
	UserHierarchyStructure ConnectElement = "user-hierarchy-structure"
	Prompts                ConnectElement = "prompt"
)

type sourceType int

const (
	fileSource sourceType = iota
	s3Source              = iota
)

type ConnectRestore struct {
	ConnectInstanceId *string
	Session           session.Session
	Source            string
	location          sourceType
	url               url.URL
	Element           ConnectElement
	NewName           string
	destinationArn    arn.ARN
	sourceArn         arn.ARN
}

func (cr ConnectRestore) Restore() error {

	switch cr.Element {
	case Flows:
		return cr.restoreFlow()
	case RoutingProfiles:
		return cr.restoreRoutingProfile()
	case Users:
		return cr.restoreUser()

	default:
		return errors.New("only restoration of contact flows is supported for now")
	}
}

func (cr ConnectRestore) restoreUser() error {
	var theUser connect.User

	cr.readSource(&theUser)

	connectSvc := connect.New(&cr.Session)

	var err error

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
			log.Fatal("Could not generate new temporary password: " + err.Error())
		}
		newProfile.Password = aws.String(res)
		newProfile.Tags = nil

		_, err = connectSvc.CreateUser(&newProfile)

		if err != nil {
			log.Fatal("Could not Create User: " + err.Error())

		}

	} else {
		//Update the existing flow in place, this requires several operations.

		//First update the profile name and description
		_, err = connectSvc.UpdateUserIdentityInfo(&connect.UpdateUserIdentityInfoInput{
			InstanceId:   cr.ConnectInstanceId,
			UserId:       theUser.Id,
			IdentityInfo: theUser.IdentityInfo,
		})

		if err != nil {
			log.Fatal("Could not Update User Identity Info: " + err.Error())

		}

		_, err = connectSvc.UpdateUserSecurityProfiles(&connect.UpdateUserSecurityProfilesInput{
			InstanceId:         cr.ConnectInstanceId,
			UserId:             theUser.Id,
			SecurityProfileIds: theUser.SecurityProfileIds,
		})

		if err != nil {
			log.Fatal("Could not Update User Security Profile: " + err.Error())
		}

		_, err = connectSvc.UpdateUserPhoneConfig(&connect.UpdateUserPhoneConfigInput{
			InstanceId:  cr.ConnectInstanceId,
			UserId:      theUser.Id,
			PhoneConfig: theUser.PhoneConfig,
		})

		if err != nil {
			log.Fatal("Could not Update User Phone config: " + err.Error())
		}

		_, err = connectSvc.UpdateUserRoutingProfile(&connect.UpdateUserRoutingProfileInput{
			InstanceId:       cr.ConnectInstanceId,
			UserId:           theUser.Id,
			RoutingProfileId: theUser.RoutingProfileId,
		})

		if err != nil {
			log.Fatal("Could not Update User Routing Profile: " + err.Error())
		}

		if theUser.HierarchyGroupId != nil {
			_, err = connectSvc.UpdateUserHierarchy(&connect.UpdateUserHierarchyInput{
				InstanceId:       cr.ConnectInstanceId,
				UserId:           theUser.Id,
				HierarchyGroupId: theUser.HierarchyGroupId,
			})
			if err != nil {
				log.Fatal("Could not Update User Hierarchy: " + err.Error())
			}

		}

	}
	return err
}

func (cr ConnectRestore) restoreRoutingProfile() error {

	var theProfile connect.RoutingProfile

	cr.readSource(&theProfile)

	connectSvc := connect.New(&cr.Session)
	var err error
	//if we have a new flow name then we are creating a new routing profile with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewName != "" {
		var newProfile connect.CreateRoutingProfileInput
		awsutil.Copy(&newProfile, &theProfile)
		newProfile.Name = aws.String(cr.NewName)
		newProfile.InstanceId = cr.ConnectInstanceId
		newProfile.Tags = nil

		result, err := connectSvc.CreateRoutingProfile(&newProfile)

		if err != nil {
			log.Fatal("Could not Create Routing Profile: " + err.Error())
		}

		cr.NewName = *result.RoutingProfileId

	} else {
		//Update the existing flow in place, this requires several operations.

		//First update the profile name and description
		_, err := connectSvc.UpdateRoutingProfileName(&connect.UpdateRoutingProfileNameInput{
			RoutingProfileId: theProfile.RoutingProfileId,
			InstanceId:       cr.ConnectInstanceId,
			Name:             theProfile.Name,
			Description:      theProfile.Description,
		})

		if err != nil {
			log.Fatal("Could not Update Routing Profile Name: " + err.Error())
		}

		//Then the concurrency
		_, err = connectSvc.UpdateRoutingProfileConcurrency(&connect.UpdateRoutingProfileConcurrencyInput{
			RoutingProfileId:   theProfile.RoutingProfileId,
			InstanceId:         cr.ConnectInstanceId,
			MediaConcurrencies: theProfile.MediaConcurrencies,
		})

		if err != nil {
			log.Fatal("Could not Update Routing Profile Concurrency: " + err.Error())
		}

		//Now the default outbound queue
		//First update the flow name
		_, err = connectSvc.UpdateRoutingProfileDefaultOutboundQueue(&connect.UpdateRoutingProfileDefaultOutboundQueueInput{
			RoutingProfileId:       theProfile.RoutingProfileId,
			InstanceId:             cr.ConnectInstanceId,
			DefaultOutboundQueueId: theProfile.DefaultOutboundQueueId,
		})

		if err != nil {
			log.Fatal("Could not Update Routing Profile Default outbound Queue: " + err.Error())
		}

		cr.NewName = *theProfile.RoutingProfileId
		if cr.location == fileSource {
			cr.Source = filepath.Dir(filepath.Dir(cr.Source)) + pathSeparator + string(RoutingProfileQueues) + "s/" + *theProfile.RoutingProfileId + jsonExtn
		} else {
			newPath := filepath.Dir(filepath.Dir(cr.url.Path)) + string(RoutingProfileQueues) + "s/" + *theProfile.RoutingProfileId + jsonExtn
			cr.Source = "s3://" + cr.url.Host + newPath
		}

	}

	err = cr.restoreRoutingProfileQueue(connectSvc)

	return err
}

func (cr ConnectRestore) restoreRoutingProfileQueue(connectSvc *connect.Connect) error {

	theProfileQueueConfig := make([]connect.RoutingProfileQueueConfigSummary, 0)

	cr.readSource(&theProfileQueueConfig)

	var err error = nil

	var queueConfigs []*connect.RoutingProfileQueueConfig

	for _, v := range theProfileQueueConfig {
		queueConfigs = append(queueConfigs, &connect.RoutingProfileQueueConfig{
			Priority: v.Priority,
			Delay:    v.Delay,
			QueueReference: &connect.RoutingProfileQueueReference{
				QueueId: v.QueueId,
				Channel: v.Channel,
			},
		})
	}

	//For routing profiel queues if there queue is not associated you must associate it.  If it's associated then you
	//must update it.  We just use brute force here and call both
	associateQueueProfile := connect.AssociateRoutingProfileQueuesInput{
		RoutingProfileId: aws.String(cr.NewName),
		InstanceId:       cr.ConnectInstanceId,
		QueueConfigs:     queueConfigs,
	}

	_, err = connectSvc.AssociateRoutingProfileQueues(&associateQueueProfile)

	updateQueueProfile := connect.UpdateRoutingProfileQueuesInput{
		RoutingProfileId: aws.String(cr.NewName),
		InstanceId:       cr.ConnectInstanceId,
		QueueConfigs:     queueConfigs,
	}
	//First update the flow name

	_, err = connectSvc.UpdateRoutingProfileQueues(&updateQueueProfile)
	return err
}

func (cr ConnectRestore) readSource(destination interface{}) {
	s3Location, _ := url.Parse(cr.Source)
	if s3Location.Scheme == "s3" {
		cr.location = s3Source
		cr.url = *s3Location
		var stream io.ReadCloser
		s3Svc := s3.New(&cr.Session)

		result, err := s3Svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(s3Location.Host),
			Key:    aws.String(s3Location.Path),
		})

		if err != nil {
			log.Fatal("Could not read source from S3: " + err.Error())
		}
		stream = result.Body
		err = jsonutil.UnmarshalJSON(destination, stream)
		if err != nil {
			log.Fatal("Could not unmarshal json source: " + err.Error())
		}
	} else {
		cr.location = fileSource
		//Assume it's a file, try opening it
		fileByte, err := ioutil.ReadFile(cr.Source)
		if err != nil {
			log.Fatal("Could not read source from file: " + err.Error())
		}
		err = json.Unmarshal(fileByte, destination)
		if err != nil {
			log.Fatal("Could not unmarshal json source: " + err.Error())
		}
	}

}

func (cr ConnectRestore) checkSourceConnectInstance(sourceArn string) bool {

	//check to see if the arn and the instance id passed on the command line are the same
	decodedSourceArn, err := arn.Parse(sourceArn)

	if err != nil {
		log.Fatal(err)
	}
	found := false
	//different := false
	//decodedDestinationArn := decodedSourceArn
	//decodedDestinationArn.Resource = "instance/" + *cr.ConnectInstanceId

	//We need the account id, use sts to obtain this and build the destination arn

	stsSvc := sts.New(&cr.Session)

	result, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}

	cr.destinationArn, err = arn.Parse(*result.Arn)

	if err != nil {
		log.Fatal(err)
	}

	cr.destinationArn.Resource = "instance/" + *cr.ConnectInstanceId

	//check the account id, want to make sure it's a match otherwise we can exit from here
	//there may be a better way to do this but it will do for now.
	if cr.destinationArn.AccountID != decodedSourceArn.AccountID {
		//source and destination account id are different
		return found
	}

	if !strings.HasPrefix(decodedSourceArn.Resource, cr.destinationArn.Resource+"/") {
		//source and destination instance id are different
		return found
	}

	//List all connect instances
	connectSvc := connect.New(&cr.Session)

	_ = connectSvc.ListInstancesPages(&connect.ListInstancesInput{}, func(output *connect.ListInstancesOutput, b bool) bool {

		//iterate through the instances

		for _, v := range output.InstanceSummaryList {

			decodeReturnedArn, _ := arn.Parse(*v.Arn)

			if strings.HasPrefix(decodeReturnedArn.Resource, cr.destinationArn.Resource) {
				found = true
				break
			}
		}
		return found
	})

	return found
}

func (cr ConnectRestore) restoreFlow() error {

	//is the location S3 or file?
	var theFlow connect.ContactFlow

	cr.readSource(&theFlow)

	var err error = nil
	cr.sourceArn, err = arn.Parse(*theFlow.Arn)
	if err != nil {
		log.Fatal(err)
	}

	connectSvc := connect.New(&cr.Session)

	//Check to see if the source is from the same connect account, instance or region.
	if !cr.checkSourceConnectInstance(*theFlow.Arn) {
		//the source is from a different connect account, instance or region to the destination.  The flow can only be
		//restored if there are no ARNS in flow for things like queues, announcements etc.
		log.Fatal("Restoration of flow is only possible to same connect account, instance and region at this time")

	}

	//if we have a new flow name then we are creating a new flow with the backup, rather than restoring over the top of
	//the old flow.
	if cr.NewName != "" {
		var newFlow connect.CreateContactFlowInput
		awsutil.Copy(&newFlow, &theFlow)
		newFlow.Name = aws.String(cr.NewName)
		newFlow.InstanceId = cr.ConnectInstanceId
		newFlow.Tags = nil

		_, err = connectSvc.CreateContactFlow(&newFlow)

	} else {

		_, err = connectSvc.UpdateContactFlowContent(&connect.UpdateContactFlowContentInput{
			ContactFlowId: theFlow.Id,
			Content:       theFlow.Content,
			InstanceId:    cr.ConnectInstanceId,
		})
	}

	if err != nil {
		log.Fatal("Could not restore Contact Flow: " + err.Error())
	}

	return err
}
