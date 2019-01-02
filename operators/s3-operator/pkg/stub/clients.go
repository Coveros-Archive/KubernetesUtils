package stub

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

func getS3Client(region string) *s3.S3 {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	s3Client := s3.New(sess)
	return s3Client
}

func getIamClient(region string) *iam.IAM {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	iamClient := iam.New(sess)
	return iamClient

}
