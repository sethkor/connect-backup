package connect_backup

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
)

func backupFlows(svc *connect.Connect, instanceId string, theWriter Writer) {
	connectInstanceId := aws.String(instanceId)
	_ = svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: connectInstanceId,
	}, func(page *connect.ListContactFlowsOutput, lastPage bool) bool {

		for _, v := range page.ContactFlowSummaryList {

			result, _ := svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    connectInstanceId,
				ContactFlowId: v.Id,
			})
			theWriter.write(*result.ContactFlow)

		}

		return true
	})
}

func Backup(svc *connect.Connect, instanceId string, theWriter Writer) {

	backupFlows(svc, instanceId, theWriter)
}
