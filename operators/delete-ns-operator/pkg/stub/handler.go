package stub

import (
	"context"
	"fmt"
	"time"

	"github.com/agill17/delete-ns-operator/pkg/apis/amritgill/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.DeleteNs:
		nsListObj := getNsListObj()
		err := sdk.List("", nsListObj)
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %v", err)
		}
		namespaces := getNamespaces(nsListObj.Items, o.Spec.Excludes)

		fmt.Println("---------------------------------------------------------------- BEGIN SCAN")
		fmt.Printf("Default Excludes: %v \n", o.Spec.Excludes)
		fmt.Printf("Final List of Namespaces after default exclusion: %v\n", namespaces)
		for name, timeCreated := range namespaces {
			timeDiff := int(time.Now().Sub(timeCreated).Hours())
			if timeDiff >= o.Spec.OlderThan {
				deleteNs(name)
			}
		}
		fmt.Printf("Namespaces older then %vhr will be deleted \n", o.Spec.OlderThan)
		fmt.Println("------------------------------------------------------------------ END SCAN")
		fmt.Printf("-\n")
		fmt.Printf("-\n")

	}
	return nil

}

func deleteNs(namespace string) {
	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	fmt.Printf("Deleting ns: %v\n", namespace)
	sdk.Delete(ns)
}

func getNsListObj() *v1.NamespaceList {
	nsPointer := &v1.NamespaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
	}
	return nsPointer
}

// returns map[namespaceName: CreationTime] after filtering out defaults
func getNamespaces(ns []v1.Namespace, excludes []string) map[string]time.Time {
	nsMetadata := make(map[string]time.Time)
	for _, v := range ns {
		if !sliceContainsString(v.Name, excludes) {
			nsMetadata[v.Name] = v.CreationTimestamp.Time
		}
	}
	return nsMetadata
}

// well.... linear not cool... but need to fix this.
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
