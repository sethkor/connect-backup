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
	RawFlow           bool
}

func (cb ConnectBackup) backupFlows() error {
	log.Println("Backing up flows")
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			result, err := cb.Svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    cb.ConnectInstanceId,
				ContactFlowId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe flow " + (*v).String())
				return true
			}

			err = cb.TheWriter.write(*result.ContactFlow)

			if err != nil {
				log.Fatal("Failed to write flow object to the destination")
			}

			if cb.RawFlow {
				err = cb.TheWriter.writeFlowString(*result.ContactFlow.Name, *result.ContactFlow.Content)

				if err != nil {
					log.Fatal("Failed to write flow string to the destination")
				}
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) BackupFlowByName(name string) error {

	log.Println("Backing/exporting flow " + name)
	foundFlow := false
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			if *v.Name != name {
				continue
			}
			foundFlow = true
			result, err := cb.Svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    cb.ConnectInstanceId,
				ContactFlowId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe flow " + (*v).String())
				return true
			}

			err = cb.TheWriter.write(*result.ContactFlow)

			if err != nil {
				log.Fatal("Failed to write flow object to the destination")
			}

			if cb.RawFlow {
				err = cb.TheWriter.writeFlowString(*result.ContactFlow.Name, *result.ContactFlow.Content)

				if err != nil {
					log.Fatal("Failed to write flow string to the destination")
				}
			}

		}
		return true
	})
	if !foundFlow {
		log.Println("Did not find a contact flow named " + name)
	}
	return err
}

func (cb ConnectBackup) backupUsers() error {
	log.Println("Backing up users")
	err := cb.Svc.ListUsersPages(&connect.ListUsersInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListUsersOutput, b bool) bool {
		for _, v := range output.UserSummaryList {

			result, err := cb.Svc.DescribeUser(&connect.DescribeUserInput{
				InstanceId: cb.ConnectInstanceId,
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
	err := cb.Svc.ListUserHierarchyGroupsPages(&connect.ListUserHierarchyGroupsInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListUserHierarchyGroupsOutput, b bool) bool {

		for _, v := range output.UserHierarchyGroupSummaryList {

			result, err := cb.Svc.DescribeUserHierarchyGroup(&connect.DescribeUserHierarchyGroupInput{
				InstanceId:       cb.ConnectInstanceId,
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

	result, err := cb.Svc.DescribeUserHierarchyStructure(&connect.DescribeUserHierarchyStructureInput{
		InstanceId: cb.ConnectInstanceId,
	})

	if err != nil {
		log.Println("Failed to describe user hierarchy structure")
		return err
	}
	return cb.TheWriter.write(*result.HierarchyStructure)

}

func (cb ConnectBackup) backupRoutingProfile() error {
	log.Println("Backing up Routing Profiles")
	err := cb.Svc.ListRoutingProfilesPages(&connect.ListRoutingProfilesInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListRoutingProfilesOutput, b bool) bool {

		for _, v := range output.RoutingProfileSummaryList {

			result, err := cb.Svc.DescribeRoutingProfile(&connect.DescribeRoutingProfileInput{
				InstanceId:       cb.ConnectInstanceId,
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
	err := cb.Svc.ListRoutingProfileQueuesPages(&connect.ListRoutingProfileQueuesInput{
		InstanceId:       cb.ConnectInstanceId,
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

func (cb ConnectBackup) RenameFlows(prefix string, allFlows bool) error {

	//List all flows
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {

		for _, v := range output.ContactFlowSummaryList {
			if !allFlows {
				if !defaultFlows[*v.Name] {
					continue
				}
			}

			_, err := cb.Svc.UpdateContactFlowName(&connect.UpdateContactFlowNameInput{
				InstanceId:    cb.ConnectInstanceId,
				Name:          aws.String(prefix + *v.Name),
				ContactFlowId: v.Id,
			})

			if err == nil {
				log.Println("Renamed from " + *v.Name + " to " + prefix + *v.Name)
			} else {
				log.Print("Failed to update name for flow " + *v.Name + ". ID: " + *v.Id)
			}
		}

		return true
	})

	return err
}
