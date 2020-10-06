package connect_backup

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func GetAwsSession(profile string, region string) *session.Session {
	var sess *session.Session
	if profile != "" {

		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Profile:           profile,
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
				MaxRetries:                    aws.Int(30),
			},
		}))

	} else {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
				MaxRetries:                    aws.Int(30),
			},
		}))
	} //else

	if region != "" {
		sess.Config.Region = aws.String(region)
	}
	return sess
}
