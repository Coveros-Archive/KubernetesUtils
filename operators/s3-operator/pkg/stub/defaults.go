package stub

import "github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"

const (
	accessKeyForNewSecret         = "ACCESS_KEY"
	secretKeyForNewSecret         = "SECRET_KEY"
	defaultRegion                 = "us-east-1"
	defaultExistenceCheckInterval = 1
	defaultSecretName             = "s3-creds"
)

// returns existenceCheck, region, iamUsername, k8sSecretName
func getDefaults(cr *v1alpha1.S3) (int, string, string, string, string) {
	existenceCheck := defaultExistenceCheckInterval
	region := defaultRegion
	iamUsername := cr.GetNamespace()
	k8sSecretName := defaultSecretName
	syncFromRegion := defaultRegion

	if cr.S3Specs.SyncWithBucket.Region != "" {
		syncFromRegion = cr.S3Specs.SyncWithBucket.Region
	}

	if cr.S3Specs.ExistenceCheckAfterMins != 0 {
		existenceCheck = cr.S3Specs.ExistenceCheckAfterMins
	}
	if cr.S3Specs.Region != "" {
		region = cr.S3Specs.Region
	}

	if cr.S3Specs.NewUser.Name != "" {
		iamUsername = cr.S3Specs.NewUser.Name
	}

	if cr.S3Specs.NewUser.SecretName != "" {
		k8sSecretName = cr.S3Specs.NewUser.SecretName
	}

	return existenceCheck, region, iamUsername, k8sSecretName, syncFromRegion

}
