package digitalocean

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"net/http"
	"os"
)

func Download(filePath string) *s3.GetObjectOutput {
	key := "DO006HT8YM4CZHXNLL6D"
	secret := "hU5t7MNPk1iU/yPnmS3XPUCcwtFIvA2njO01ASX4eAU"

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String("https://fra1.digitaloceanspaces.com"),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String("fra1"),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		panic(err)
	}
	s3Client := s3.New(newSession)

	input := s3.GetObjectInput{
		Bucket: aws.String("discord-ai"),
		Key:    aws.String(filePath),
	}
	getObject, err := s3Client.GetObject(&input)

	if err != nil {
		if requestFailure, ok := err.(awserr.RequestFailure); ok && requestFailure.StatusCode() == http.StatusNotFound {
			return nil
		}
		panic(err)
	}
	return getObject
}

func Upload(opusFile, filePath string, tags map[string]*string) {
	open, err := os.Open(opusFile)
	if err != nil {
		panic(err)
	}
	key := "DO006HT8YM4CZHXNLL6D"                           // Access key pair. You can create access key pairs using the control panel or API.
	secret := "hU5t7MNPk1iU/yPnmS3XPUCcwtFIvA2njO01ASX4eAU" // Secret access key defined through an environment variable.

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String("https://fra1.digitaloceanspaces.com"),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String("fra1"),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		panic(err)
	}
	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket:   aws.String("discord-ai"),
		Key:      aws.String(filePath),
		Body:     open,
		ACL:      aws.String("private"),
		Metadata: tags,
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		panic(err)
	}
}
