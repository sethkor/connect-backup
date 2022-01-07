# connect-backup
A tool to back-up and restore AWS Connect.  You can back-up to file, S3 or just to stdout.

```
usage: connect-backup --instance=INSTANCE [<flags>] <command> [<args> ...]

A tool to backup and restore your AWS Connect instance

Flags:
  -h, --help               Show context-sensitive help (also try --help-long and --help-man).
      --profile=PROFILE    AWS credentials/config file profile to use
      --region=REGION      AWS region
      --instance=INSTANCE  The AWS Connect instance id to backup
  -v, --version            Show application version.

Commands:
  help [<command>...]
    Show help.

  backup [<flags>]
    Backup your instance

  restore --type=TYPE [<flags>] <json>
    Restore a connect component

  rename-flows [<flags>]
    Rename all call flows with a suffix
```

## Getting connect-backup
Easiest way to install if you're on a Mac or Linux (amd64 or arm64)  is to use [Homebrew](https://brew.sh/)

Type:

```
brew tap sethkor/tap
brew install connect-backup
```

For other platforms take a look at the releases in Github.  I build binaries for:

|OS            | Architecture                           |
|:------------ |:-------------------------------------- |
|Mac (Darwin)  | amd64 (aka x86_64)                     |
|Linux         | amd64, arm64, 386 (32 bit) |
|Windows       | amd64, 386 (32 bit)                   |

Let me know if you would like a particular os/arch binary regularly built.

## Lambda
If you'd rather set up a lambda to periodically trigger a backup, clone the repo as it contains all the lambda bits and
a template to use with [AWS SAM](https://aws.amazon.com/serverless/sam/) to deploy it.  You will need to update the env.mk 
file with the values fo your environment.

Then just simply:
```
make lambda
make sam-deploy
```

You can either set environment variables for the lambda or trigger the lambda with an event json that contains the connect instance id and S3 bucket URL like this:
```
{
  "ConnectInstanceId": "your-AWS-connect-instance-id",
  "S3DestURL": "s3://your-backup-bucket/whatever-prefix-you want-like-the-instance-id"
  "FlowsRaw": true
}
```

`FlowsRaw`, which is  boolean and doesn't need quotes, follows the same logic as `--flows-raw` on the command line (see below) where the contact flow is also written 
to it's own file in S3 with pretty print json.  If the value is omitted it is treated as false.

`ConnectInstanceId` is only required if you wish to backup a specific connect instance.  Omitting this will backup all 
instances (IAM policy permitting).

The sam template in `lambda/template.yaml` contains a single sample `AWS::Events::Rule` with an Input that constructs 
this JSON.  You can add additional `AWS::Events::Rule` to back up other connect instances (or the same one to different 
destinations).  If you are using the same backup bucket to backup multiple connect instances, try adding the connect
instance id as a prefix in the `S3DestURL` value of the json.

You can also specify the connect instance is and s3 destination URL as environment vars and leave the event blank.  This
provides some backward compatibility to early generations of this lambda that relied soley on environment vars.

If you want to undeploy you can run:
```
make sam-remove
```
but remember to make sure your bucket is empty first (including all object versions) otherwise you won't be able to drop
the stack.

## What is included in the backup
- [X] Published Call Flows (The AWS API restricts this to published flows only)
- [X] Raw Call flows as json objects without AWS Connect provisioning metadata
- [X] Routing Profiles including Routing Profile Queues
- [X] User Data (except Passwords)
- [X] User Hierarchy Groups
- [X] User Hierarchy
- [X] Prompt Data (But not any wav files)
- [X] Hours of Operation
- [X] Quick Connects
- [X] Queues (except the default AGENT queue)

For contact flows, the actual flow is a json object encapsulated within the connect json flow object.  If you wish to export also just
the flow as a json object, pass the `--flows-raw` flag and it will write the contact flow itself as a seperate json in 
the `flows-raw` directory of prefix.  This seperate raw flow is for informational purposes only and is not involved in restoration.

connect-backup use a directory/prefix (see what I did there?) structure so everything is neat and tidy.  If the structure
is not there it will create it on the fly:
```
your-connect-backup-workspace
   └──your-connect-instance-id
       ├──common
       ├──flows
       ├──flows-raw
       ├──hours-of-operation
       ├──prompts
       ├──quick-connects
       ├──routing-profile-queues
       ├──routing-profiles
       ├──user-hierarchy-groups
       └──users
````

If you wish to only backup or export a single contact flow, pass `--flow-name` to the backup comand.

The default behaviour is to backup every connect instance found unless you specify an instance with `--instance`

## What about Queues?
Currently, there is no API call to describe or create queues.  When the API becomes available, I'll add it.

## Restoration
You can restore AWS Connect elements you have previously backed up:

- [X] Published Call Flows (The AWS API restricts this to published flows only)
- [X] Routing Profiles including Routing Profile Queues
- [X] User Data (except Passwords)
- [ ] User Hierarchy Groups
- [X] User Hierarchy 

The `--create` flg will allow you to create a new element, rather than overwriting the existing one.

If you choose to restore with a new call flow name via `--create` you can only do this once for the new name.  If you wish
to overwrite this new flow with another restore then omit `--create` like a normal overwrite restoration.

When restoring Users, in order for the restoration to be reflected in the AWS Connect Console, you must refresh the 
User Management screen.  This is due to the console using the listing on this screen as a cache to the underlying data.

You can use the restore function for a user to update the users first/last name by editing the json backup file.  You can't
do this via the AWS Connect Console at all.

If you use the `--create` flag when restoring a user a new user will be created with the user id passed with the `--create`
flag.  The password will be set to a very random long string (64chars, Caps and Upper case, Symbols and Numbers included)
Which won't be returned.  You will have to instruct the user to go through the password reset process to reset it.  If the
user already exists the user will not be recreated or updated.

Further enhancements for restoration, including restoration between instances is WIP.

## Renaming all the contact flows
AWS Connect won't let you delete any contact flows. Ever.  Also every new instance you create comes with a bunch of example 
contact flows you can never delete either.  This leads to your contact flow workspace jumbling up the contact flows you
create and work with every day with the examples making things annoyingly hard to find.  Now you can use `--rename-flows`
which will add a prefix to the default set of AWS demo contact flows that are created when your AWS Connect Instance is first created
which can help you with sorting and put all the example flows at the bottom of  your contact flow list.  If you
wish to rename **all** contact flows pass the `--all-flows` flag. The default prefix is `~` (you can supply a different one 
to use) which will mean it will sort the renamed flows at the bottom.  You can run this when you first create a connect
instance or any time after.  As the Name is really only metadata, renaming flows won't impact any references or live call flows.

## IAM Policy
You will need the following IAM access ata minimum.  The Lambda example deploys this policy for you.  The resource scope
is lest as `*` to cover the use case of backing up all connect instances, but the scope can be limited to a particular instance only (as per the comments below).

```
            - Effect: Allow
              Action:
                - s3:PutObject
                - s3:PutObjectACL
              Resource: !GetAtt s3Bucket.Arn
            - Effect: Allow
              Action:
                - ds:DescribeDirectories
              Resource: "*"
            - Effect: Allow
              Action:
                - connect:ListInstances
              Resource: "*"
            - Effect: Allow
              Action:
                - connect:ListContactFlow
                - connect:ListRoutingProfiles
                - connect:ListUserHierarchyGroups
                - connect:ListUsers
                - connect:ListPrompts
                - connect:ListHoursOfOperations
                - connect:ListQueues
                - connect:DescribeUserHierarchyStructure
                - connect:DescribeInstance
                - connect:DescribeQueue
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}"
            - Effect: Allow
              Action:
                - connect:DescribeContactFlow
                - connect:ListContactFlows
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/contact-flow/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/contact-flow/*"
            - Effect: Allow
              Action:
                - connect:DescribeUser
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/agent/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/agent/*"
            - Effect: Allow
              Action:
                - connect:DescribeRoutingProfile
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/routing-profile/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/routing-profile/*"
            - Effect: Allow
              Action:
                - connect:DescribeUserHierarchyGroup
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/agent-group/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/agent-group/*"
            - Effect: Allow
              Action:
                - connect:ListQuickConnects
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/transfer-destination/*"
#              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/transfer-destination/*"
            - Effect: Allow
              Action:
                - connect:DescribeHoursOfOperation
              Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/*/operating-hours/*"
 #             Resource: !Sub "arn:aws:connect:${AWS::Region}:${AWS::AccountId}:instance/${connectInstanceId}/operating-hours/*"
 ```
## FAQ
#### Can I take a backup json and restore it manually via the AWS Connect Console?
No.  The export/import function on the console supports a completley different format to the AWs API leveraged by `connect-backup`

#### How about restoring a call flow export taken manually from the AWS Connect Console?
No.  Like the question above, the formats are very different.

#### Can I restore to a different connect instance as the source?
No, not yet anyway.  AWS Connect objects have a lot of ARN's that need to be manipulated plus some intelligence to
correlate a few things.  This is a WIP but requires _A LOT_ of work.

#### Can I back-up and restore saved flows?
No.  Only published flows can be operated on.  This is a limitation of the AWS API.

#### Why can't I restore routing profile queues?
The AWS API appears to have a bug with the UpdateRoutingProfileQueue API currently

#### Why can't I restore a user hierarchy group to be empty?
The AWS API doesn't accept an empty or nil value for this currently

### What is the Raw Flow?
Contact flows are json objects stored within another json object.  This means they are escaped and can't be parsed or 
read easily.  The ecapsulating json object also has other attributes (name, description etc) that are needed for restoration.
A Raw flow takes this json object within the json object, unescapes it and pretty prints it so you can have a better visual
representation of your contact flow as a json object.

#### I've found a bug, what do I fo?
Report it and I'll take a look.

#### Do you accept feature requests?
Yes.  Let me know what you would like to see and I'll consider adding it to the backlog.

#### Will you provide any other handy tools for AWS Connect.
Yes, now that we finally have an AWS API I'll add usefull things over time.  You may also want to take a look at my tools
for provisioning AWS Lex chat bots via yaml/json called [Lexbelt](https://github.com/sethkor/lexbelt)



