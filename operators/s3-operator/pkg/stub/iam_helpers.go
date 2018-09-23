package stub

import (
	"encoding/base64"

	"github.com/sirupsen/logrus"

	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

type IamUserInput struct {
	IAMClient                            *iam.IAM
	Username, AccessPolicyArn, Namespace string
	AccessKeysOutput                     *iam.CreateAccessKeyOutput
	ObjectStore                          *v1alpha1.S3
}

func (i *IamUserInput) IamUserExists() bool {
	exists := true

	_, err := i.IAMClient.GetUser(&iam.GetUserInput{
		UserName: aws.String(i.Username),
	})
	if awserr, ok := err.(awserr.Error); ok && awserr.Code() == iam.ErrCodeNoSuchEntityException {
		exists = false
	}
	logrus.Warnf("Namespace: %v | IAM User: %v | Msg: User exists: %v", i.Namespace, i.Username, exists)

	return exists
}

func (i *IamUserInput) CreateUserIfDoesNotExists() error {
	var err error

	if !i.IamUserExists() {
		_, err = i.IAMClient.CreateUser(&iam.CreateUserInput{
			UserName: aws.String(i.Username),
		})
		if i.AccessPolicyArn != "" {
			if _, err := i.IAMClient.AttachUserPolicy(&iam.AttachUserPolicyInput{PolicyArn: aws.String(i.AccessPolicyArn), UserName: aws.String(i.Username)}); err != nil {
				logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: ERROR while attaching accessPolicy; %v", i.Namespace, i.Username, err)
			}
		} else {
			logrus.Infof("Namespace: %v | PolicyArn: %v | Msg: Access policy does not exist. Skipping attachment", i.Namespace, i.AccessPolicyArn)
		}
		i.AccessKeysOutput, err = i.IAMClient.CreateAccessKey(&iam.CreateAccessKeyInput{
			UserName: aws.String(i.Username),
		})
		i.ObjectStore.Status.AccessKey = base64.StdEncoding.EncodeToString([]byte(*i.AccessKeysOutput.AccessKey.AccessKeyId))
		i.ObjectStore.Status.SecretKey = base64.StdEncoding.EncodeToString([]byte(*i.AccessKeysOutput.AccessKey.SecretAccessKey))
		err = sdk.Update(i.ObjectStore)
		if err != nil {
			logrus.Errorf("ERROR While updating objectStore for Namespace: %v", i.Namespace)
		}
	}

	return err
}

func (i *IamUserInput) DeleteIamUser(accessKeyFromCr string) {
	logrus.Infof("Namespace: %v | IAM Username: %v | Msg: Deleting user", i.Namespace, i.Username)

	if _, err := i.IAMClient.DetachUserPolicy(&iam.DetachUserPolicyInput{PolicyArn: aws.String(i.AccessPolicyArn), UserName: aws.String(i.Username)}); err != nil {
		logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: ERROR while detaching accessPolicy; %v", i.Namespace, i.Username, err)
	}

	if _, err := i.IAMClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{AccessKeyId: aws.String(accessKeyFromCr), UserName: aws.String(i.Username)}); err != nil {
		logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: ERROR while deleting accessKey; %v", i.Namespace, i.Username, err)
	}

	if _, err := i.IAMClient.DeleteUser(&iam.DeleteUserInput{UserName: aws.String(i.Username)}); err != nil {
		logrus.Errorf("Namespace: %v | IAM Username: %v | Msg: ERROR while deleting IAM, %v", i.Namespace, i.Username, err)
	}

	logrus.Infof("Namespace: %v | IAM Username: %v | Msg: Deleted user", i.Namespace, i.Username)

}
