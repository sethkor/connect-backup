package connect_backup

import (
	"log"

	"github.com/aws/aws-sdk-go/service/connect"
)

type ConnectBackup struct {
	ConnectInstanceId *string
}

func (cb ConnectBackup) backupFlows(svc *connect.Connect, theWriter Writer) error {
	log.Println("Backing up flows")
	connectInstanceId := cb.ConnectInstanceId
	err := svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			result, err := svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    connectInstanceId,
				ContactFlowId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe flow " + (*v).String())
				return true
			}
			err = theWriter.write(*result.ContactFlow)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) backupUsers(svc *connect.Connect, theWriter Writer) error {
	log.Println("Backing up users")
	connectInstanceId := cb.ConnectInstanceId
	err := svc.ListUsersPages(&connect.ListUsersInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUsersOutput, b bool) bool {
		for _, v := range output.UserSummaryList {

			result, err := svc.DescribeUser(&connect.DescribeUserInput{
				InstanceId: connectInstanceId,
				UserId:     v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user " + (*v).String())
				return true
			}
			err = theWriter.write(*result.User)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) backupUserHierarchyGroups(svc *connect.Connect, theWriter Writer) error {
	log.Println("Backing up user hierarchy groups")
	connectInstanceId := cb.ConnectInstanceId
	err := svc.ListUserHierarchyGroupsPages(&connect.ListUserHierarchyGroupsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUserHierarchyGroupsOutput, b bool) bool {

		for _, v := range output.UserHierarchyGroupSummaryList {

			result, err := svc.DescribeUserHierarchyGroup(&connect.DescribeUserHierarchyGroupInput{
				InstanceId:       connectInstanceId,
				HierarchyGroupId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user hierarchy group " + (*v).String())
				return true
			}
			err = theWriter.write(*result.HierarchyGroup)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})
	return err
}

func (cb ConnectBackup) backupUserHierarchyStructure(svc *connect.Connect, theWriter Writer) error {
	log.Println("Backing up hierarchy structures")
	connectInstanceId := cb.ConnectInstanceId

	result, err := svc.DescribeUserHierarchyStructure(&connect.DescribeUserHierarchyStructureInput{
		InstanceId: connectInstanceId,
	})

	if err != nil {
		log.Println("Failed to describe user hierarchy structure")
		return err
	}
	return theWriter.write(*result.HierarchyStructure)

}

func (cb ConnectBackup) backupRoutingProfile(svc *connect.Connect, theWriter Writer) error {
	log.Println("Backing up Routing Profiles")
	connectInstanceId := cb.ConnectInstanceId
	err := svc.ListRoutingProfilesPages(&connect.ListRoutingProfilesInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListRoutingProfilesOutput, b bool) bool {

		for _, v := range output.RoutingProfileSummaryList {

			result, _ := svc.DescribeRoutingProfile(&connect.DescribeRoutingProfileInput{
				InstanceId:       connectInstanceId,
				RoutingProfileId: v.Id,
			})

			err := theWriter.write(*result.RoutingProfile)

			if err != nil {
				log.Fatal("Failed to write to the destination")
			}

		}
		return true
	})

	return err
}

func (cb ConnectBackup) Backup(svc *connect.Connect, theWriter Writer) error {

	err := cb.backupFlows(svc, theWriter)
	if err != nil {
		return err
	}
	err = cb.backupUsers(svc, theWriter)
	if err != nil {
		return err
	}
	err = cb.backupRoutingProfile(svc, theWriter)
	if err != nil {
		return err
	}
	err = cb.backupUserHierarchyGroups(svc, theWriter)
	if err != nil {
		return err
	}
	return cb.backupUserHierarchyStructure(svc, theWriter)

}
