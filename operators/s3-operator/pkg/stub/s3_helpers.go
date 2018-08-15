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

		if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
			logrus.Errorf("Unable to delete objects from bucket %q, %v", bucket, err)
		}
		logrus.Infof("Deleted object(s) from bucket: %s", bucket)

		_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Unable to delete %v bucket for namespace %v. Error --, %v", bucket, ns, err)
		}
		logrus.Infof("Waiting for %v bucket to be deleted...", bucket)

		err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			logrus.Errorf("Error occurred while waiting for bucket to be deleted, %v", bucket)
		}
		logrus.Infof("Bucket %v in region %v successfully deleted for namespace: %v", bucket, region, ns)

	} else {
		logrus.Errorf("ERROR!!! Deleting bucket", bucket, "in", region, ": Bucket does not exist")
	}

}

func CreateBucket(bucketName, region, synWith, ns string, tags map[string]string, svc *s3.S3) error {

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
			logrus.Errorf("Unable to create %v bucket. Error was returned --, %v", bucket, err)
		} else {
			logrus.Infof("Waiting for %v bucket to be created...", bucket)
			err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				logrus.Errorf("Error occurred while waiting for bucket to be created, %v", bucket)
			} else {
				addTagsToS3Bucket(bucket, t, svc)
				logrus.Infof("%v bucket successfully created for namespace: %v", bucket, ns)
			}
		}
	} else {
		msg := fmt.Sprint("ERROR!!! Creating bucket:", bucketName, "in", region, ": Bucket ALREADY exist")
		logrus.Errorf(msg)
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
