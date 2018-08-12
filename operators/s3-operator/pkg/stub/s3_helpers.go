package stub

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func listBuckets(region string) []string {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	allBuckets := []string{}
	svc := s3.New(sess)
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		exitErrorf("Failed to list buckets", err)
	}
	for _, bucket := range result.Buckets {
		allBuckets = append(allBuckets, aws.StringValue(bucket.Name))
	}
	return allBuckets
}

func BucketExists(bucketName, region string) bool {
	exists := false
	availBuckets := listBuckets(region)
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
func DeleteBucket(bucket, region string) {

	os.Setenv("AWS_REGION", region)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if BucketExists(bucket, region) {
		svc := s3.New(sess)
		iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
		})

		if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
			exitErrorf("Unable to delete objects from bucket %q, %v", bucket, err)
		}
		fmt.Printf("Deleted object(s) from bucket: %s\n", bucket)

		_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			exitErrorf("Unable to delete bucket %q, %v", bucket, err)
		}
		fmt.Printf("Waiting for bucket %q to be deleted...\n", bucket)

		err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			exitErrorf("Error occurred while waiting for bucket to be deleted, %v", bucket)
		}
		fmt.Printf("Bucket %q successfully deleted\n", bucket)

	} else {
		exitErrorf("ERROR!!! Deleting bucket", bucket, "in", region, ": Bucket does not exist")
	}

}

func CreateBucket(bucketName, region, synWith string, tags map[string]string) {

	os.Setenv("AWS_REGION", region)
	bucket := bucketName

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	// Create S3 service client
	svc := s3.New(sess)

	// Create the S3 Bucket
	if !BucketExists(bucketName, region) {
		logrus.Infof("Creating Bucket...")
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			exitErrorf("Unable to create bucket %q, %v", bucket, err)
		}
		fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

		err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			exitErrorf("Error occurred while waiting for bucket to be created, %v", bucket)
		}
		fmt.Printf("Bucket %q successfully created\n", bucket)
	} else {
		exitErrorf("ERROR!!! Creating bucket:", bucketName, "in", region, ": Bucket ALREADY exist")
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+" ", args...)
	os.Exit(1)
}

func SyncBucketWith(newBucket, oldBucket, region string) {
	// sync

}
