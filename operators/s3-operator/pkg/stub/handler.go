package stub

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"

	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func getS3SvcSetup(region string) *s3.S3 {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	svc := s3.New(sess)
	return svc
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	objectStore := event.Object.(*v1alpha1.S3)
	ns := objectStore.GetNamespace
	svc := getS3SvcSetup(objectStore.S3Specs.Region)
	os.Setenv("AWS_REGION", objectStore.S3Specs.Region)

	metdataLabels := objectStore.ObjectMeta.GetLabels()
	if _, exists := metdataLabels["namespace"]; !exists {
		metdataLabels["namespace"] = ns()
	}
	if objectStore.Status.Deployed != true {
		logrus.Infof("Creating %v bucket in %v for namespace: %v", objectStore.S3Specs.BucketName, objectStore.S3Specs.Region, ns())
		err := CreateBucket(
			objectStore.S3Specs.BucketName,
			objectStore.S3Specs.Region,
			objectStore.S3Specs.SyncWith.BucketName,
			ns(),
			metdataLabels,
			svc,
		)
		if err != nil {
			logrus.Errorf("Something failed while creating the s3 bucket for namespace: ", ns)
		} else {
			objectStore.Status.Deployed = true
			err := sdk.Update(objectStore)
			if err != nil {
				return fmt.Errorf("failed to update s3 status: %v", err)
			}
		}
	}

	if event.Deleted {
		DeleteBucket(objectStore.S3Specs.BucketName, objectStore.S3Specs.Region, ns(), svc)
	}

	return nil
}
