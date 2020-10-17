package connect_backup

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/connect"
)

type ConnectBackup struct {
	ConnectInstanceId *string
	Svc               *connect.Connect
	TheWriter         Writer
}

func (cb ConnectBackup) backupFlows() error {
	log.Println("Backing up flows")
	connectInstanceId := cb.ConnectInstanceId
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			result, err := cb.Svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    connectInstanceId,
				ContactFlowId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe flow " + (*v).String())
				return true
			}

			err = cb.TheWriter.write(*result.ContactFlow)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) backupUsers() error {
	log.Println("Backing up users")
	connectInstanceId := cb.ConnectInstanceId
	err := cb.Svc.ListUsersPages(&connect.ListUsersInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUsersOutput, b bool) bool {
		for _, v := range output.UserSummaryList {

			result, err := cb.Svc.DescribeUser(&connect.DescribeUserInput{
				InstanceId: connectInstanceId,
				UserId:     v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user " + (*v).String())
				return true
			}
			err = cb.TheWriter.write(*result.User)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) backupUserHierarchyGroups() error {
	log.Println("Backing up user hierarchy groups")
	connectInstanceId := cb.ConnectInstanceId
	err := cb.Svc.ListUserHierarchyGroupsPages(&connect.ListUserHierarchyGroupsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUserHierarchyGroupsOutput, b bool) bool {

		for _, v := range output.UserHierarchyGroupSummaryList {

			result, err := cb.Svc.DescribeUserHierarchyGroup(&connect.DescribeUserHierarchyGroupInput{
				InstanceId:       connectInstanceId,
				HierarchyGroupId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user hierarchy group " + (*v).String())
				return true
			}
			err = cb.TheWriter.write(*result.HierarchyGroup)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})
	return err
}

func (cb ConnectBackup) backupUserHierarchyStructure() error {
	log.Println("Backing up hierarchy structures")
	connectInstanceId := cb.ConnectInstanceId

	result, err := cb.Svc.DescribeUserHierarchyStructure(&connect.DescribeUserHierarchyStructureInput{
		InstanceId: connectInstanceId,
	})

	if err != nil {
		log.Println("Failed to describe user hierarchy structure")
		return err
	}
	return cb.TheWriter.write(*result.HierarchyStructure)

}

func (cb ConnectBackup) backupRoutingProfile() error {
	log.Println("Backing up Routing Profiles")
	connectInstanceId := cb.ConnectInstanceId
	err := cb.Svc.ListRoutingProfilesPages(&connect.ListRoutingProfilesInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListRoutingProfilesOutput, b bool) bool {

		for _, v := range output.RoutingProfileSummaryList {

			result, err := cb.Svc.DescribeRoutingProfile(&connect.DescribeRoutingProfileInput{
				InstanceId:       connectInstanceId,
				RoutingProfileId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user routing profile")
			}

			err = cb.TheWriter.write(*result.RoutingProfile)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

			err = cb.backupRoutingProfileQueues(*result.RoutingProfile.RoutingProfileId)

		}
		return true
	})

	return err
}

type backupRoutingProfileQueueSummary struct {
	routingProfile                   string
	routingProfileQueueConfigSummary connect.RoutingProfileQueueConfigSummary
}

func (cb ConnectBackup) backupRoutingProfileQueues(routingProfileId string) error {
	log.Println("Backing up Routing Profile Queues")
	connectInstanceId := cb.ConnectInstanceId
	err := cb.Svc.ListRoutingProfileQueuesPages(&connect.ListRoutingProfileQueuesInput{
		InstanceId:       connectInstanceId,
		RoutingProfileId: aws.String(routingProfileId),
	}, func(output *connect.ListRoutingProfileQueuesOutput, b bool) bool {
		_ = cb.TheWriter.writeList(routingProfileId, output.RoutingProfileQueueConfigSummaryList)
		return true
	})

	return err
}

func (cb ConnectBackup) Backup() error {

	err := cb.backupFlows()
	if err != nil {
		return err
	}
	err = cb.backupUsers()
	if err != nil {
		return err
	}
	err = cb.backupRoutingProfile()
	if err != nil {
		return err
	}
	err = cb.backupUserHierarchyGroups()
	if err != nil {
		return err
	}
	return cb.backupUserHierarchyStructure()

}
