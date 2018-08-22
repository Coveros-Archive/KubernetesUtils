package stub

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func listBuckets(region string, svc *s3.S3) []string {
	allBuckets := []string{}
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		logrus.Errorf("Failed to list buckets", err)
	}
	for _, bucket := range result.Buckets {
		allBuckets = append(allBuckets, aws.StringValue(bucket.Name))
	}
	return allBuckets
}

func BucketExists(bucketName, region string, svc *s3.S3) bool {
	exists := false
	availBuckets := listBuckets(region, svc)
	if sliceContainsString(bucketName, availBuckets) {
		exists = true
	}
	return exists
}

func sliceContainsString(whichValue string, whichSlice []string) bool {
	exists := false
	for _, ele := range whichSlice {
		if whichValue == ele {
			exists = true
			break
		}
	}
	return exists
}

// Assumes empty the bucket and then delete it
// Perhaps this can be parameterized
func DeleteBucket(bucket, region, ns string, svc *s3.S3) {

	if BucketExists(bucket, region, svc) {
		iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
		})
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Deleting all objects ", ns, bucket)

		if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to delete objects %v", ns, bucket, err)
		}
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Deleted all objects ", ns, bucket)

		_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to delete bucket %v", ns, bucket, err)
		}

		err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Error while deleting bucket %v", ns, bucket, err)
		}
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Bucket Deleted ", ns, bucket)

	} else {
		logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Bucket does not exist while deleting %v", ns, bucket)
	}
}

func CreateBucket(bucketName, region, ns string, tags map[string]string, svc *s3.S3) error {

	bucket := bucketName
	t := []*s3.Tag{}
	for k, v := range tags {
		t = append(t, &s3.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	var err error
	// Create the S3 Bucket
	if !BucketExists(bucketName, region, svc) {
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Unable to create bucket %v", ns, bucket, err)
		} else {
			err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Error occured while bucket creation %v", ns, bucket, err)
			} else {
				addTagsToS3Bucket(bucket, t, svc)
				logrus.Infof("Namespace: %v | Bucket: %v | Msg: Bucket created successfully", ns, bucket)
			}
		}
	} else {
		logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Bucket already exists", ns, bucket)
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
