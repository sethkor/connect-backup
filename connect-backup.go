package connect_backup

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
)

func backupFlows(svc *connect.Connect, instanceId string, theWriter Writer) {
	connectInstanceId := aws.String(instanceId)
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

func backupUsers(svc *connect.Connect, instanceId string, theWriter Writer) {
	connectInstanceId := aws.String(instanceId)
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

func backupUserHierarchyGroups(svc *connect.Connect, instanceId string, theWriter Writer) {
	connectInstanceId := aws.String(instanceId)
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

func backupRoutingProfile(svc *connect.Connect, instanceId string, theWriter Writer) {
	connectInstanceId := aws.String(instanceId)
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

func Backup(svc *connect.Connect, instanceId string, theWriter Writer) {

	backupFlows(svc, instanceId, theWriter)
	backupUsers(svc, instanceId, theWriter)
	backupRoutingProfile(svc, instanceId, theWriter)
	backupUserHierarchyGroups(svc, instanceId, theWriter)

}
