package stub

import (
	"context"
	"fmt"
	"os"

	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	s3Svc := getS3SvcSetup(objectStore.S3Specs.Region)
	bucket := objectStore.S3Specs.BucketName
	region := objectStore.S3Specs.Region
	syncWith := objectStore.S3Specs.SyncWith.BucketName
	metdataLabels := objectStore.ObjectMeta.GetLabels()
	if _, exists := metdataLabels["namespace"]; !exists {
		metdataLabels["namespace"] = ns()
	}

	os.Setenv("AWS_REGION", region)

	if objectStore.Status.Deployed != true {
		logrus.Infof("Namespace: %v | Bucket: %v | Msg: Creating Bucket ", ns(), bucket)
		err := CreateBucket(
			bucket, region,
			syncWith, ns(),
			metdataLabels, s3Svc,
		)
		if err != nil {
			logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Error while creating bucket ", ns(), bucket, err)
		} else {
			objectStore.Status.Deployed = true
			err := sdk.Update(objectStore)
			if err != nil {
				logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Failed to update bucket status ", ns(), bucket, err)
			}
			assumedURL := fmt.Sprintf("%v.%v.amazonaws.com", bucket, region)
			externalSvc := createExternalService("s3", ns(), assumedURL, metdataLabels)
			err = sdk.Create(externalSvc)
			if err != nil {
				logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Failed to create externalName service %v", ns(), bucket, err)
			} else {
				logrus.Infof("Namespace: %v | Bucket: %v | Msg: Created externalName service ", ns(), bucket)
			}
		}
	}

	if event.Deleted {
		DeleteBucket(bucket, region, ns(), s3Svc)
	}

	return nil
}

func createExternalService(name, ns, endpoint string, labels map[string]string) *v1.Service {
	s := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: endpoint,
		},
	}
	return s
}
