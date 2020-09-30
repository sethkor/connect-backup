package connect_backup

import (
	"os"

	"github.com/aws/aws-sdk-go/service/connect"
)

type ConnectBackup struct {
	ConnectInstanceId *string
}

func (cb ConnectBackup) backupFlows(svc *connect.Connect, theWriter Writer) {
	connectInstanceId := cb.ConnectInstanceId
	_ = svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {

		for _, v := range output.ContactFlowSummaryList {

			result, _ := svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    connectInstanceId,
				ContactFlowId: v.Id,
			})
			err := theWriter.write(*result.ContactFlow)

			if err != nil {
				os.Exit(1)
			}

		}
		return true
	})
}

func (cb ConnectBackup) backupUsers(svc *connect.Connect, theWriter Writer) {
	connectInstanceId := cb.ConnectInstanceId
	_ = svc.ListUsersPages(&connect.ListUsersInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUsersOutput, b bool) bool {

		for _, v := range output.UserSummaryList {

			result, _ := svc.DescribeUser(&connect.DescribeUserInput{
				InstanceId: connectInstanceId,
				UserId:     v.Id,
			})

			err := theWriter.write(*result.User)

			if err != nil {
				os.Exit(1)
			}

		}
		return true
	})
}

func (cb ConnectBackup) backupUserHierarchyGroups(svc *connect.Connect, theWriter Writer) {
	connectInstanceId := cb.ConnectInstanceId
	_ = svc.ListUserHierarchyGroupsPages(&connect.ListUserHierarchyGroupsInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListUserHierarchyGroupsOutput, b bool) bool {

		for _, v := range output.UserHierarchyGroupSummaryList {

			result, _ := svc.DescribeUserHierarchyGroup(&connect.DescribeUserHierarchyGroupInput{
				InstanceId:       connectInstanceId,
				HierarchyGroupId: v.Id,
			})

			err := theWriter.write(*result.HierarchyGroup)

			if err != nil {
				os.Exit(1)
			}

		}
		return true
	})
}

func (cb ConnectBackup) backupUserHierarchyStructure(svc *connect.Connect, theWriter Writer) {
	connectInstanceId := cb.ConnectInstanceId

	result, _ := svc.DescribeUserHierarchyStructure(&connect.DescribeUserHierarchyStructureInput{
		InstanceId: connectInstanceId,
	})

	err := theWriter.write(*result.HierarchyStructure)

	if err != nil {
		os.Exit(1)
	}
}

func (cb ConnectBackup) backupRoutingProfile(svc *connect.Connect, theWriter Writer) {
	connectInstanceId := cb.ConnectInstanceId
	_ = svc.ListRoutingProfilesPages(&connect.ListRoutingProfilesInput{
		InstanceId: connectInstanceId,
	}, func(output *connect.ListRoutingProfilesOutput, b bool) bool {

		for _, v := range output.RoutingProfileSummaryList {

			result, _ := svc.DescribeRoutingProfile(&connect.DescribeRoutingProfileInput{
				InstanceId:       connectInstanceId,
				RoutingProfileId: v.Id,
			})

			err := theWriter.write(*result.RoutingProfile)

			if err != nil {
				os.Exit(1)
			}

		}
		return true
	})
}

func (cb ConnectBackup) Backup(svc *connect.Connect, theWriter Writer) {

	cb.backupFlows(svc, theWriter)
	cb.backupUsers(svc, theWriter)
	cb.backupRoutingProfile(svc, theWriter)
	cb.backupUserHierarchyGroups(svc, theWriter)
	cb.backupUserHierarchyStructure(svc, theWriter)
}
