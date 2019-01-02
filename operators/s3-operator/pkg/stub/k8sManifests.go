package stub

import (
	"github.com/agill17/s3-operator/pkg/apis/amritgill/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func secretObjSpec(crObj *v1alpha1.S3) *v1.Secret {
	o := true
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crObj.S3Specs.NewUser.SecretName,
			Namespace: crObj.GetNamespace(),
			Labels:    crObj.ObjectMeta.GetLabels(),
			OwnerReferences: []metav1.OwnerReference{{
				Name:               crObj.GetName(),
				APIVersion:         crObj.APIVersion,
				Kind:               crObj.Kind,
				UID:                crObj.GetUID(),
				BlockOwnerDeletion: &o,
				Controller:         &o,
			}},
		},
		Type: v1.SecretType("Opaque"),
		Data: map[string][]byte{
			accessKeyForNewSecret: []byte(encodeDecode(crObj.Status.AccessKey, "decode")),
			secretKeyForNewSecret: []byte(encodeDecode(crObj.Status.SecretKey, "decode")),
		},
	}
}

func getSecret(objectStore *v1alpha1.S3, secretName string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: objectStore.GetNamespace(),
		},
	}
}
