package stub

import (
	"context"
	"fmt"
	"time"

	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

// What this is doing;
/*
	1. Operator get deployed in a central namespace
	2. Cr gets created in multiple namespaces
	3. Operator checks CR status ( timeUpdated and deployed )
	4. If timeUpdated == nil && deployed = false, then make aws calls
	5. Create IAM User if it does not exist
	6. Base64 encode the AccessKeyId in CR annotation for future use
	7. Create S3 bucket if it does not exist
	8. Update each CR Status ( timeUpdated -- time.format and deployed -- boolean )
	9. Create Kube Secret using IAM creds ( base64 encoded ) -- Outside of main condition check so this is auto-created if deleted by operator-sdk
	10. Check CR how often to verify existence of external resource and update CR statuses at each interval

	This way if bucket / IAM user gets deleted, the operator is able to auto-create the services again.
	However the check will be not dont every RESYNC_PERIOD, this will apply globally.
	Instead we fine-grain control how often to check for each CR.

	TODO:
	if IAM User is re-created, need to update aws-creds secrets (done)
	copy contents from an existing bucket to new bucket (done)
*/

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	objectStore := event.Object.(*v1alpha1.S3)
	ns := objectStore.GetNamespace()
	existenceCheck, region, iamUsername, k8sSecretName, syncFromRegion := getDefaults(objectStore)
	s3Client, iamClient := getS3Client(region), getIamClient(region)
	bucket := objectStore.S3Specs.BucketName
	syncFromBucket := objectStore.S3Specs.SyncWithBucket.BucketName
	accessPolicy := objectStore.S3Specs.NewUser.Policy
	metdataLabels := objectStore.ObjectMeta.GetLabels()
	if _, exists := metdataLabels["namespace"]; !exists {
		metdataLabels["namespace"] = ns
	}
	var err error

	//iamUser inputs
	iamUser := iamUserInput{
		iamClient:       iamClient,
		username:        iamUsername,
		accessPolicyArn: accessPolicy,
		namespace:       ns,
		objectStore:     objectStore,
	}

	// s3 inputs
	s3Input := s3Input{
		s3Svc:          s3Client,
		bucketName:     bucket,
		bucketTags:     metdataLabels,
		region:         region,
		namespace:      ns,
		syncFromBucket: syncFromBucket,
		syncFromRegion: syncFromRegion,
	}
	timeNullCheckVal := *new(time.Time)
	timeDiffSinceLastCheck := int(time.Now().Sub(objectStore.Status.TimeUpdated).Minutes())

	// check initial cr status
	if objectStore.Status.TimeUpdated == timeNullCheckVal && iamUser.iamUserExists() && objectStore.Status.AccessKey == "" {
		logrus.Errorf("IAM User - %v - already exists at initial run.. Cannot continue until user is deleted for NS: %v", iamUsername, ns)
	} else if (objectStore.Status.TimeUpdated == timeNullCheckVal) || (objectStore.Status.Deployed && timeDiffSinceLastCheck >= existenceCheck) {

		// create IAM user if it does not exists and get accessKey
		iamUser.createUserIfDoesNotExists()

		// Create Bucket if it does not exist
		err = s3Input.createBucketIfDoesNotExist()
		if err == nil {
			objectStore.Status.Deployed = true
			objectStore.Status.TimeUpdated = time.Now()
			err = sdk.Update(objectStore)
			errorCheck(err, func() {
				logrus.Errorf("Namespace: %v | Bucket: %v | Msg: Failed to update CR status %v", ns, bucket, err)
			})
		}

	}

	deployIAMSecret(objectStore, metdataLabels, k8sSecretName)
	if event.Deleted {
		cleanUp(s3Input, objectStore, iamUser)
	}

	return nil

}

func cleanUp(s3Input s3Input, cr *v1alpha1.S3, iamUser iamUserInput) {
	s3Input.deleteBucket()
	if cr.Status.AccessKey != "" {
		decodedAccessKey := encodeDecode(cr.Status.AccessKey, "decode")
		iamUser.deleteIamUser(decodedAccessKey)
	} else {
		logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: CR does not have accessKeyId", cr.GetNamespace(), cr.S3Specs.NewUser.Name)
		logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: Please delete this user by hand!", cr.GetNamespace(), cr.S3Specs.NewUser.Name)
	}
}

func deployIAMSecret(objectStore *v1alpha1.S3, metdataLabels map[string]string, secretName string) {
	// create secret -- this will be picked after resync period for each CR deployed in a namespace

	defer updateSecretIfNeeded(objectStore, secretName)

	err := sdk.Create(secretObjSpec(objectStore))

	if err != nil && !apierrors.IsAlreadyExists(err) && !apierrors.IsForbidden(err) {
		logrus.Errorf("Failed to create Aws Secret : %v", err)
	}

}

func updateSecretIfNeeded(objectStore *v1alpha1.S3, secretName string) {

	secretObj := getSecret(objectStore, secretName)
	err := sdk.Get(secretObj)

	if !apierrors.IsNotFound(err) && !apierrors.IsForbidden(err) {
		deocdedAccessKeyFromSecret := fmt.Sprintf("%s", secretObj.Data[accessKeyForNewSecret])
		decodedAccessKeyFromCR := encodeDecode(objectStore.Status.AccessKey, "decode")

		if deocdedAccessKeyFromSecret != decodedAccessKeyFromCR {
			logrus.Warnf("Updating secret \"%v\" with new creds for namespace: %v", objectStore.S3Specs.NewUser.SecretName, objectStore.GetNamespace())
			err = sdk.Update(secretObjSpec(objectStore))
			errorCheck(err, func() { logrus.Errorf("ERROR updating secrets in namespace %v: %v", objectStore.GetNamespace(), err) })
		}
	}

}
