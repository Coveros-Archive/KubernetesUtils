package stub

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Input struct {
	S3Svc                         *s3.S3
	BucketName, Region, Namespace string
	BucketTags                    map[string]string
}

func (s *S3Input) BucketExists() bool {
	exists := true
	_, err := s.S3Svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &s.BucketName})
	if awserr, ok := err.(awserr.Error); ok && awserr.Code() == s3.ErrCodeNoSuchBucket {
		exists = false
	}
	return exists
}

// Assumes empty the bucket and then delete it
// Perhaps this can be parameterized
func (s *S3Input) DeleteBucket() {

	if s.BucketExists() {
		iter := s3manager.NewDeleteListIterator(s.S3Svc, &s3.ListObjectsInput{
			Bucket: aws.String(s.BucketName),
		})
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Deleting all objects ", s.Namespace, s.BucketName)

		if err := s3manager.NewBatchDeleteWithClient(s.S3Svc).Delete(aws.BackgroundContext(), iter); err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to delete objects %v", s.Namespace, s.BucketName, err)
		}
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Deleted all objects ", s.Namespace, s.BucketName)

		_, err := s.S3Svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(s.BucketName),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to delete bucket %v", s.Namespace, s.BucketName, err)
		}

		err = s.S3Svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: aws.String(s.BucketName),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Error while deleting bucket %v", s.Namespace, s.BucketName, err)
		}
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Bucket Deleted ", s.Namespace, s.BucketName)

	} else {
		logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Bucket does not exist while deleting %v", s.Namespace, s.BucketName)
	}
}

func (s *S3Input) CreateBucketIfDoesNotExist() error {

	bucket := s.BucketName
	t := []*s3.Tag{}
	for k, v := range s.BucketTags {
		t = append(t, &s3.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	var err error
	// Create the S3 Bucket
	if !s.BucketExists() {
		_, err = s.S3Svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to create bucket %v", s.Namespace, s.BucketName, err)
		} else {
			err = s.S3Svc.WaitUntilBucketExists(&s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Error occured while bucket creation %v", s.Namespace, s.BucketName, err)
			} else {
				addTagsToS3Bucket(bucket, t, s.S3Svc)
				logrus.Infof("Namespace: %v | Bucket: %v | Msg: Bucket created successfully", s.Namespace, s.BucketName)
			}
		}
	} else {
		logrus.Warnf("Namespace: %v | Bucket: %v | Msg: Bucket already exists", s.Namespace, s.BucketName)
	}
	return err
}

func SyncBucketWith(newBucket, oldBucket, region string, svc *s3.S3) {
	// sync
	// cross account? or same account ? or both ( ideal? )
}

func addTagsToS3Bucket(whichBucket string, tags []*s3.Tag, svc *s3.S3) error {
	tagInput := &s3.PutBucketTaggingInput{
		Bucket: &whichBucket,
		Tagging: &s3.Tagging{
			TagSet: tags,
		},
	}

	_, err := svc.PutBucketTagging(tagInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return err
}
