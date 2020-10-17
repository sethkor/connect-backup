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
    backup your instance

  restore [<flags>] <json>
    Restore a connect component
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

If you want to undeploy you can run:
```
make sam-remove
```
but remember to make sure your bucket is empty first (including all object versions) otherwise you won't be able to drop
the stack.

## What is included in the backup
- [X] Published Call Flows (The AWS API restricts this to published flows only)
- [X] Routing Profiles including Routing Profile Queues
- [X] User Data (except Passwords)
- [X] User Hierarchy Groups
- [X] User Hierarchy 

connect-backup use a directory/prefix (see what I did there?) structure so everything is neat and tidy.  If the structure
is not there it will create it on the fly:
```
your-connect-backup-workspace
   ├──flows
   ├──routing-profiles
   ├──users
   └──user-hierarchy-groups

````
### What about Queues?
Currently, there is no API call to describe or create queues.  When the API becomes available, I'll add it.

## Restoration
You can restore AWS Connect elements you have previously backed up:

- [X] Published Call Flows (The AWS API restricts this to published flows only)
- [X] Routing Profiles (except Queues)
- [X] User Data (except Passwords)
- [ ] User Hierarchy Groups
- [ ] User Hierarchy 

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

## FAQ
#### Can I take a backup json and restore it manually via the AWS Connect Console?
No.  The export/import function on the console supports a completley different format to the AWs API leveraged by `connect-backup`

#### How about restoring a call flow export taken manually from the AWS Connect Console?
No.  Like the question above, the formats are very different.

#### Can I back-up and restore saved flows?
No.  Only published flows can be operated on.  This is a limitation of the AWS API.

#### I've found a bug, what do I fo?
Report it and I'll take a look.

#### Do you accept feature requests?
Yes.  Let me know what you would like to see and I'll consider adding it to the backlog.

#### Will you provide any other handy tools for AWS Connect.
Yes, now that we finally have an AWS API I'll add usefull things over time.  You may also want to take a look at my tools
for provisioning AWS Lex chat bots via yaml/json called [Lexbelt](https://github.com/sethkor/lexbelt)



