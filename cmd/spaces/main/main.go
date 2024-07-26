// Step 1: Import the all necessary libraries and SDK commands.
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
)

func main() {
	// Step 2: Define the parameters for the session you want to create.
	key := "DO006HT8YM4CZHXNLL6D"                           // Access key pair. You can create access key pairs using the control panel or API.
	secret := "hU5t7MNPk1iU/yPnmS3XPUCcwtFIvA2njO01ASX4eAU" // Secret access key defined through an environment variable.

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""), // Specifies your credentials.
		Endpoint:         aws.String("https://fra1.digitaloceanspaces.com"), // Find your endpoint in the control panel, under Settings. Prepend "https://".
		S3ForcePathStyle: aws.Bool(false),                                   // // Configures to use subdomain/virtual calling format. Depending on your version, alternatively use o.UsePathStyle = false
		Region:           aws.String("fra1"),                                // Must be "us-east-1" when creating new Spaces. Otherwise, use the region in your endpoint, such as "nyc3".
	}

	// Step 3: The new session validates your request and directs it to your Space's specified endpoint using the AWS SDK.

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		panic(err)
	}
	s3Client := s3.New(newSession)

	open, err := os.Open("test.opus")
	if err != nil {
		panic(err)
	}

	// Step 4: Define the parameters of the object you want to upload.
	object := s3.PutObjectInput{
		Bucket: aws.String("discord-ai"),              // The path to the directory you want to upload the object to, starting with your Space name.
		Key:    aws.String("youtube/cache/test.opus"), // Object key, referenced whenever you want to access this file later.
		Body:   open,                                  // The object's contents.
		ACL:    aws.String("private"),                 // Defines Access-control List (ACL) permissions, such as private or public.
		Metadata: map[string]*string{ // Required. Defines metadata tags.
			"x-amz-meta-my-key": aws.String("your-value"),
		},
	}

	// Step 5: Run the PutObject function with your parameters, catching for errors.
	_, err = s3Client.PutObject(&object)
	if err != nil {
		panic(err)
	}

	input := s3.GetObjectInput{
		Bucket: aws.String("discord-ai"), // The path to the directory you want to upload the object to, starting with your Space name.
		Key:    aws.String("youtube/cache/test.opus"),
	}
	getObject, err := s3Client.GetObject(&input)
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(getObject.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}
