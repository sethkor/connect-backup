package connect_backup

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/connect"
)

type ConnectBackup struct {
	Svc             *connect.Connect
	TheWriter       Writer
	RawFlow         bool
	ConnectInstance connect.Instance
}

func (cb ConnectBackup) backupFlows() error {
	log.Println("Backing up Flows")
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			result, err := cb.Svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    cb.ConnectInstance.Id,
				ContactFlowId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe flow "+(*v).String(), ". ", err)
				continue
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

	log.Println("Backing up Flow " + name)
	foundFlow := false
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {
		for _, v := range output.ContactFlowSummaryList {

			if *v.Name != name {
				continue
			}
			foundFlow = true
			result, err := cb.Svc.DescribeContactFlow(&connect.DescribeContactFlowInput{
				InstanceId:    cb.ConnectInstance.Id,
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
	log.Println("Backing up Users")
	err := cb.Svc.ListUsersPages(&connect.ListUsersInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListUsersOutput, b bool) bool {
		for _, v := range output.UserSummaryList {

			result, err := cb.Svc.DescribeUser(&connect.DescribeUserInput{
				InstanceId: cb.ConnectInstance.Id,
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
	log.Println("Backing up user Hierarchy Groups")
	err := cb.Svc.ListUserHierarchyGroupsPages(&connect.ListUserHierarchyGroupsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListUserHierarchyGroupsOutput, b bool) bool {

		for _, v := range output.UserHierarchyGroupSummaryList {

			result, err := cb.Svc.DescribeUserHierarchyGroup(&connect.DescribeUserHierarchyGroupInput{
				InstanceId:       cb.ConnectInstance.Id,
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
	log.Println("Backing up Hierarchy Structures")

	result, err := cb.Svc.DescribeUserHierarchyStructure(&connect.DescribeUserHierarchyStructureInput{
		InstanceId: cb.ConnectInstance.Id,
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
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListRoutingProfilesOutput, b bool) bool {

		for _, v := range output.RoutingProfileSummaryList {

			result, err := cb.Svc.DescribeRoutingProfile(&connect.DescribeRoutingProfileInput{
				InstanceId:       cb.ConnectInstance.Id,
				RoutingProfileId: v.Id,
			})

			if err != nil {
				log.Println("Failed to describe user routing profile")
			}

			err = cb.TheWriter.write(*result.RoutingProfile)

			if err != nil {
				log.Println("Failed to write to the destination")
			}

			err = cb.backupRoutingProfileQueues(*result.RoutingProfile.RoutingProfileId)

			if err != nil {
				log.Println("Failed to backup Routing Profile Queues " + *result.RoutingProfile.RoutingProfileId)
			}

		}
		return true
	})

	return err
}

type backupRoutingProfileQueueSummary struct {
	routingProfile string
	//routingProfileQueueConfigSummary connect.RoutingProfileQueueConfigSummary
}

func (cb ConnectBackup) backupRoutingProfileQueues(routingProfileId string) error {
	log.Println("Backing up Routing Profile Queues")
	err := cb.Svc.ListRoutingProfileQueuesPages(&connect.ListRoutingProfileQueuesInput{
		InstanceId:       cb.ConnectInstance.Id,
		RoutingProfileId: aws.String(routingProfileId),
	}, func(output *connect.ListRoutingProfileQueuesOutput, b bool) bool {
		_ = cb.TheWriter.writeList(routingProfileId, output.RoutingProfileQueueConfigSummaryList)
		return true
	})

	return err
}

func (cb ConnectBackup) backupPrompts() error {
	log.Println("Backing up Prompts")
	result, err := cb.Svc.ListPrompts(&connect.ListPromptsInput{
		InstanceId: cb.ConnectInstance.Id,
	})

	_ = cb.TheWriter.writeList(string(Prompts), result.PromptSummaryList)
	return err

}

func (cb ConnectBackup) backupHours() error {
	log.Println("Backing up Hours")

	err := cb.Svc.ListHoursOfOperationsPages(&connect.ListHoursOfOperationsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListHoursOfOperationsOutput, b bool) bool {

		for _, v := range output.HoursOfOperationSummaryList {
			result, err := cb.Svc.DescribeHoursOfOperation(&connect.DescribeHoursOfOperationInput{
				InstanceId:         cb.ConnectInstance.Id,
				HoursOfOperationId: v.Id,
			})
			if err != nil {
				log.Println("Failed to describe Hours of Operation")
				log.Println(err)
				return true
			}

			err = cb.TheWriter.write(*result.HoursOfOperation)

			if err != nil {
				log.Println(err)
				log.Fatal("Failed to write to the destination")

			}
		}
		return true
	})
	return err
}

func (cb ConnectBackup) backupQuickConnects() error {
	log.Println("Backing up QuickConnects")

	var allOutputs []*connect.QuickConnectSummary
	err := cb.Svc.ListQuickConnectsPages(&connect.ListQuickConnectsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListQuickConnectsOutput, b bool) bool {

		allOutputs = append(allOutputs, output.QuickConnectSummaryList...)
		return true
	})

	_ = cb.TheWriter.write(allOutputs)
	return err
}

func (cb ConnectBackup) backupLambdas() error {
	log.Println("Backing up Lambdas")

	var allOutputs lambdaStrings
	err := cb.Svc.ListLambdaFunctionsPages(&connect.ListLambdaFunctionsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListLambdaFunctionsOutput, b bool) bool {

		allOutputs = append(allOutputs, output.LambdaFunctions...)
		return true
	})

	_ = cb.TheWriter.write(allOutputs)
	return err
}

func (cb ConnectBackup) backupLex() error {
	log.Println("Backing up Lex Bots")

	var allOutputs []*connect.LexBot
	err := cb.Svc.ListLexBotsPages(&connect.ListLexBotsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListLexBotsOutput, b bool) bool {

		allOutputs = append(allOutputs, output.LexBots...)
		return true
	})

	_ = cb.TheWriter.write(allOutputs)
	return err
}

func (cb ConnectBackup) backupQueues() error {
	log.Println("Backing up Queue")

	err := cb.Svc.ListQueuesPages(&connect.ListQueuesInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListQueuesOutput, b bool) bool {

		for _, v := range output.QueueSummaryList {
			if *v.QueueType != "AGENT" {
				result, err := cb.Svc.DescribeQueue(&connect.DescribeQueueInput{
					InstanceId: cb.ConnectInstance.Id,
					QueueId:    v.Id,
				})

				if err != nil {
					log.Println("Error Describing Queue "+*v.Id, " "+*v.Name)
					continue
				}

				err = cb.TheWriter.write(*result.Queue)
			}
		}
		return true
	})
	return err
}

func (cb ConnectBackup) backupInstance() error {
	log.Println("Backing up Instance")

	result, err := cb.Svc.DescribeInstance(&connect.DescribeInstanceInput{
		InstanceId: cb.ConnectInstance.Id,
	})

	if err != nil {
		log.Println("error Descrining instance " + *cb.ConnectInstance.Id)
		return err
	}

	err = cb.TheWriter.write(*result.Instance)

	return err
}

func (cb ConnectBackup) backupItems() {

	var err error

	err = cb.backupInstance()
	if err != nil {
		log.Print("Error backing up Prompts")
		log.Println(err)
	}
	err = cb.backupLambdas()
	if err != nil {
		log.Print("Error backing up Lambdas")
		log.Println(err)
	}
	err = cb.backupLex()
	if err != nil {
		log.Print("Error backing up Lex Bots")
		log.Println(err)
	}
	err = cb.backupPrompts()
	if err != nil {
		log.Print("Error backing up Prompts")
		log.Println(err)
	}
	err = cb.backupHours()
	if err != nil {
		log.Print("Error backing up Operating Hours")
		log.Println(err)
	}
	err = cb.backupQuickConnects()
	if err != nil {
		log.Print("Error backing up Quick Connects")
		log.Println(err)
	}
	err = cb.backupFlows()
	if err != nil {
		log.Print("Error backing up Flows")
		log.Println(err)
	}
	err = cb.backupUsers()
	if err != nil {
		log.Print("Error backing up Users")
		log.Println(err)
	}
	err = cb.backupRoutingProfile()
	if err != nil {
		log.Print("Error backing up Routing Profiles")
		log.Println(err)
	}
	err = cb.backupUserHierarchyGroups()
	if err != nil {
		log.Print("Error backing up Hierarchy Groups")
		log.Println(err)
	}
	err = cb.backupUserHierarchyStructure()
	if err != nil {
		log.Print("Error backing up Hierarchy Structure")
		log.Println(err)
	}
	err = cb.backupQueues()
	if err != nil {
		log.Print("Error backing up Queues")
		log.Println(err)
	}
}

func (cb ConnectBackup) Backup() error {

	var err error = nil

	var found bool = false
	if *cb.ConnectInstance.Id != "" {
		err = cb.Svc.ListInstancesPages(&connect.ListInstancesInput{}, func(output *connect.ListInstancesOutput, b bool) bool {

			for _, v := range output.InstanceSummaryList {
				if *cb.ConnectInstance.Id == *v.Id {
					log.Println("Backing up instance " + *v.InstanceAlias + ", " + *v.Id)
					err = cb.TheWriter.init(*cb.ConnectInstance.Id)
					cb.backupItems()
					found = true
				}
			}
			return true
		})

		if !found {
			log.Println("Instance ID passed was not found")
		}

	} else {
		err = cb.Svc.ListInstancesPages(&connect.ListInstancesInput{}, func(output *connect.ListInstancesOutput, b bool) bool {

			for _, v := range output.InstanceSummaryList {
				log.Println("Backing up instance " + *v.InstanceAlias + ", " + *v.Id)
				cb.ConnectInstance.Id = v.Id
				err = cb.TheWriter.init(*cb.ConnectInstance.Id)
				cb.backupItems()
			}
			return true
		})
	}
	return err
}

func (cb ConnectBackup) RenameFlows(prefix string, allFlows bool) error {

	//List all flows
	err := cb.Svc.ListContactFlowsPages(&connect.ListContactFlowsInput{
		InstanceId: cb.ConnectInstance.Id,
	}, func(output *connect.ListContactFlowsOutput, b bool) bool {

		for _, v := range output.ContactFlowSummaryList {
			if !allFlows {
				if !defaultFlows[*v.Name] {
					continue
				}
			}

			_, err := cb.Svc.UpdateContactFlowName(&connect.UpdateContactFlowNameInput{
				InstanceId:    cb.ConnectInstance.Id,
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
