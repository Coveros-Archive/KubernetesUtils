package stub

import (
	"context"
	"fmt"

	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

const (
	defaultPvc           = "minio-pvc"
	defaultVolName       = "data"
	defaultDataDir       = "/storage"
	defaultAccessKey     = "SOME_ACCESS_KEY"
	defaultSecretKey     = "SOME_SECRET_KEY"
	defaultImage         = "minio/minio"
	defaultContainerName = "minio"
	defaultLivenessPath  = "/minio/health/live"
	defaultReadinessPath = "/minio/health/ready"
	defaultLivenessPort  = 9000
	defaultContainerPort = 9000
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	s3 := event.Object.(*v1alpha1.S3)
	ns := s3.GetNamespace

	if s3.Status.Deployed != true {
		fmt.Println(s3.S3Specs.BucketName, "bucket does not exists for ", ns())
		CreateBucket(
			s3.S3Specs.BucketName,
			s3.S3Specs.Region,
			s3.S3Specs.SyncWith.BucketName,
			ns(),
			s3.S3Specs.Labels,
		)
		s3.Status.Deployed = true
		err := sdk.Update(s3)
		if err != nil {
			return fmt.Errorf("failed to update s3 status: %v", err)
		}
	}

	if event.Deleted {
		DeleteBucket(s3.S3Specs.BucketName, s3.S3Specs.Region, ns())
	}

	return nil
}

// func createMinioDeployment(cr *v1alpha1.S3) *appsv1.Deployment {
// 	labels := cr.MinioSpecs.Labels
// 	dataDir := cr.MinioSpecs.DataDir
// 	if len(dataDir) == 0 {
// 		dataDir = defaultDataDir
// 	}
// 	ns := cr.Namespace
// 	name := cr.Name
// 	image := defaultImage
// 	containerName := defaultContainerName
// 	s3_secret := cr.MinioSpecs.SecretKey
// 	s3_access := cr.MinioSpecs.AccessKey
// 	if len(s3_access) == 0 {
// 		s3_access = defaultAccessKey
// 	}
// 	if len(s3_secret) == 0 {
// 		s3_secret = defaultSecretKey
// 	}
// 	dep := &appsv1.Deployment{
// 		TypeMeta: metav1.TypeMeta{
// 			APIVersion: "apps/v1",
// 			Kind:       "Deployment",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name,
// 			Namespace: ns,
// 		},
// 		Spec: appsv1.DeploymentSpec{
// 			Selector: &metav1.LabelSelector{
// 				MatchLabels: labels,
// 			},
// 			Template: v1.PodTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Labels: labels,
// 				},
// 				Spec: v1.PodSpec{
// 					Containers: []v1.Container{{
// 						Image: image,
// 						Name:  containerName,
// 						Args:  []string{"server", defaultDataDir},
// 						Ports: []v1.ContainerPort{{
// 							ContainerPort: defaultContainerPort,
// 							Name:          defaultContainerName,
// 						}},
// 						LivenessProbe: &v1.Probe{
// 							Handler: v1.Handler{
// 								HTTPGet: &v1.HTTPGetAction{
// 									Path: defaultLivenessPath,
// 									Port: intstr.FromInt(defaultLivenessPort),
// 								},
// 							},
// 							InitialDelaySeconds: 120,
// 							PeriodSeconds:       20,
// 						},
// 						ReadinessProbe: &v1.Probe{
// 							Handler: v1.Handler{
// 								HTTPGet: &v1.HTTPGetAction{
// 									Path: defaultReadinessPath,
// 									Port: intstr.FromInt(defaultLivenessPort),
// 								},
// 							},
// 							InitialDelaySeconds: 120,
// 							PeriodSeconds:       20,
// 						},
// 						Resources: &v1.ResourceRequirements{
// 							Requests: &v1.ResourceList{},
// 						},
// 					}},
// 				},
// 			},
// 		},
// 	}

// 	return dep
// }
