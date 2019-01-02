package stub

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Input struct {
	s3Svc                         *s3.S3
	bucketName, region, namespace string
	bucketTags                    map[string]string
	syncFromBucket                string
	syncFromRegion                string
}

func listBucketObjects(s *s3Input, whichBucket string) []string {
	// s3Clients are not multi-regional so therefore setting up a new client where the existing bucket exists inorder for list to work properly
	s3Client := getS3Client(s.syncFromRegion)
	files, _ := s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(whichBucket),
	})
	allKeys := []string{}
	for _, e := range files.Contents {
		allKeys = append(allKeys, *e.Key)
	}
	logrus.Infof("All keys from %v: %v", whichBucket, allKeys)
	return allKeys
}

func copyItemsFromOneBucketToAnother(s *s3Input) {
	syncFromBucket := s.syncFromBucket

	items := listBucketObjects(s, syncFromBucket)
	itemsLength := len(items)

	copyTo := s.bucketName
	var wg sync.WaitGroup
	wg.Add(itemsLength)
	for i := 0; i < itemsLength; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := s.s3Svc.CopyObject(&s3.CopyObjectInput{
				Bucket:     aws.String(copyTo),
				CopySource: aws.String(syncFromBucket + "/" + items[i]),
				Key:        aws.String(items[i]),
			})
			if err != nil {
				logrus.Errorf("Unable to copy item %v from bucket %q to bucket %q: %v", items[i], syncFromBucket, copyTo, err)
			}
		}(i)
	}
	wg.Wait()

}

func (s *s3Input) bucketExists(bucketToCheck string) bool {
	exists := true
	_, err := s.s3Svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &bucketToCheck})
	if awserr, ok := err.(awserr.Error); ok && awserr.Code() == s3.ErrCodeNoSuchBucket {
		exists = false
	}
	return exists
}

// Assumes empty the bucket and then delete it
// Perhaps this can be parameterized
func (s *s3Input) deleteBucket() {

	if s.bucketExists(s.bucketName) {

		// empty out the bucket first
		s.emptyBucket()

		_, err := s.s3Svc.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(s.bucketName)})
		errorCheck(err, func() {
			logrus.Errorf("namespace: %v | Bucket: %v | Msg: Unable to delete bucket %v", s.namespace, s.bucketName, err)
		})

		// wait until bucket does not exist
		err = s.waitBucketExistenceCheck("notExist", s.bucketName)
		if err == nil {
			logrus.Infof("namespace: %v | Bucket: %v | Msg: Bucket Deleted ", s.namespace, s.bucketName)
		}

	} else {
		logrus.Errorf("namespace: %v | Bucket: %v | Msg: Bucket does not exist while deleting %v", s.namespace, s.bucketName)
	}
}

func (s *s3Input) createBucketIfDoesNotExist() error {

	bucket := s.bucketName
	t := []*s3.Tag{}
	for k, v := range s.bucketTags {
		t = append(t, &s3.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	var err error
	// Create the S3 Bucket
	if !s.bucketExists(s.bucketName) {

		bucketInput := &s3.CreateBucketInput{
			Bucket: &bucket,
		}

		if s.region != "us-east-1" {
			bucketInput.SetCreateBucketConfiguration(&s3.CreateBucketConfiguration{LocationConstraint: &s.region})
		}

		_, err = s.s3Svc.CreateBucket(bucketInput)
		errorCheck(err, func() {
			logrus.Errorf("namespace: %v | Bucket: %v | Msg: Unable to create bucket %v", s.namespace, s.bucketName, err)
		})

		// wait until bucket exists
		err = s.waitBucketExistenceCheck("exist", s.bucketName)
		if err == nil {
			addTagsToS3Bucket(bucket, t, s.s3Svc)
			logrus.Infof("namespace: %v | Bucket: %v | Msg: Bucket created successfully", s.namespace, s.bucketName)

			if len(s.syncFromBucket) != 0 && s.bucketExists(s.syncFromBucket) {
				logrus.Infof("namespace: %v | Bucket: %v | Msg: Copying all items from %v", s.namespace, s.bucketName, s.syncFromBucket)
				copyItemsFromOneBucketToAnother(s)
			}
		}

	} else {
		logrus.Warnf("namespace: %v | Bucket: %v | Msg: Bucket already exists", s.namespace, s.bucketName)
	}
	return err
}

func addTagsToS3Bucket(whichBucket string, tags []*s3.Tag, svc *s3.S3) {
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
}

func (s s3Input) waitBucketExistenceCheck(whichCheck string, whichBucket string) error {
	var err error
	switch whichCheck {
	case "notExist":
		err = s.s3Svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: &whichBucket,
		})
		errorCheck(err, func() {
			logrus.Errorf("namespace: %v | Bucket: %v | Msg: Error while waiting for bucket to no longer exist %v", s.namespace, s.bucketName, err)
		})
		break
	case "exist":
		err = s.s3Svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &whichBucket})
		errorCheck(err, func() {
			logrus.Errorf("namespace: %v | Bucket: %v | Msg: Error while waiting for bucket to exist %v", s.namespace, whichBucket, err)
		})
		break
	}
	return err
}

func (s s3Input) emptyBucket() {
	iter := s3manager.NewDeleteListIterator(s.s3Svc, &s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
	})
	logrus.Infof("namespace: %v | Bucket: %v | Msg: Deleting all objects ", s.namespace, s.bucketName)

	err := s3manager.NewBatchDeleteWithClient(s.s3Svc).Delete(aws.BackgroundContext(), iter)
	errorCheck(err, func() {
		logrus.Errorf("namespace: %v | Bucket: %v | Msg: Unable to delete objects %v", s.namespace, s.bucketName, err)
	})
	logrus.Infof("namespace: %v | Bucket: %v | Msg: Deleted all objects ", s.namespace, s.bucketName)
}
