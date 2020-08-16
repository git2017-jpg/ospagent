package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"sigs.k8s.io/yaml"

	//"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

type DynamicResource struct {
	*kubernetes.KubeClient
	*schema.GroupVersionResource
}

func NewDynamicResource(kubeClient *kubernetes.KubeClient, gv *schema.GroupVersionResource) *DynamicResource {
	return &DynamicResource{
		KubeClient:           kubeClient,
		GroupVersionResource: gv,
	}
}

type DynamicDeleteResourceParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DynamicDeleteParams struct {
	Resources []DynamicDeleteResourceParams `json:"resources"`
}

func (d *DynamicResource) Delete(deleteParams interface{}) *utils.Response {
	params := &DynamicDeleteParams{}
	json.Unmarshal(deleteParams.([]byte), params)
	if len(params.Resources) > 0 {
		//go func() {
		deletePolicy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		for _, r := range params.Resources {
			if err := d.DynamicClient.Resource(*d.GroupVersionResource).Namespace(r.Namespace).Delete(r.Name, deleteOptions); err != nil {
				klog.Errorf("delete group %v namespace %s name %s error: %v", d.GroupVersionResource, r.Namespace, r.Name, err)
				return &utils.Response{Code: code.DeleteError, Msg: fmt.Sprintf("Delete %s error: %s", r.Name, err.Error())}
			}
		}
		//}()
	}
	return &utils.Response{Code: code.Success}
}

type DynamicUpdateParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	YamlStr   string `json:"yaml"`
}

func (d *DynamicResource) Update(updateParams interface{}) *utils.Response {
	params := &DynamicUpdateParams{}
	json.Unmarshal(updateParams.([]byte), params)
	klog.Info(params.YamlStr)
	mapObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(params.YamlStr), &mapObj); err != nil {
		klog.Error("Parse yaml error: ", err)
		return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("Parse yaml error: %s", err.Error())}
	}
	obj := &unstructured.Unstructured{Object: mapObj}
	_, updateErr := d.DynamicClient.Resource(*d.GroupVersionResource).Namespace(params.Namespace).Update(obj, metav1.UpdateOptions{})
	if updateErr != nil {
		klog.Error("Update error: ", updateErr)
		return &utils.Response{Code: code.UpdateError, Msg: updateErr.Error()}
	}
	return &utils.Response{Code: code.Success}
}
