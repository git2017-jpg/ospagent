package resource

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeYaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
	"strings"

	//"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

type DynamicResource struct {
	*kubernetes.KubeClient
	*schema.GroupVersionResource
	decUnstructured runtime.Serializer
	restMapper      *restmapper.DeferredDiscoveryRESTMapper
	context         context.Context
}

func NewDynamicResource(kubeClient *kubernetes.KubeClient, gv *schema.GroupVersionResource) *DynamicResource {
	return &DynamicResource{
		KubeClient:           kubeClient,
		GroupVersionResource: gv,
		decUnstructured:      runtimeYaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		restMapper:           restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(kubeClient.DiscoveryClient)),
		context:              context.Background(),
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
	Kind      string `json:"kind"`
}

func (d *DynamicResource) UpdateYaml(updateParams interface{}) *utils.Response {
	params := &DynamicUpdateParams{}
	json.Unmarshal(updateParams.([]byte), params)
	mapObj := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(params.YamlStr), &mapObj); err != nil {
		klog.Error("Parse yaml error: ", err)
		return &utils.Response{Code: code.ParamsError, Msg: fmt.Sprintf("Parse yaml error: %s", err.Error())}
	}
	obj := &unstructured.Unstructured{Object: mapObj}
	var updateErr error
	if params.Namespace != "" {
		_, updateErr = d.DynamicClient.Resource(*d.GroupVersionResource).Namespace(params.Namespace).Update(obj, metav1.UpdateOptions{})
	} else {
		_, updateErr = d.DynamicClient.Resource(*d.GroupVersionResource).Update(obj, metav1.UpdateOptions{})
	}
	if updateErr != nil {
		klog.Error("Update error: ", updateErr)
		return &utils.Response{Code: code.UpdateError, Msg: updateErr.Error()}
	}
	return &utils.Response{Code: code.Success}
}

type ApplyParams struct {
	YamlStr string `json:"yaml"`
}

func (d *DynamicResource) ApplyYaml(applyParams interface{}) *utils.Response {
	params := &ApplyParams{}
	json.Unmarshal(applyParams.([]byte), params)
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader([]byte(params.YamlStr))))
	var res []string
	applyErr := false
	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return &utils.Response{Code: code.ParamsError, Msg: "read yaml error: " + err.Error()}
		}
		obj, dr, err := d.buildDynamicResourceClient(buf)
		if err != nil {
			applyErr = true
			res = append(res, err.Error())
			continue
		}

		// Create or Update
		_, err = dr.Patch(obj.GetName(), types.ApplyPatchType, buf, metav1.PatchOptions{
			FieldManager: "ospagent",
		})
		if err != nil {
			applyErr = true
			res = append(res, obj.GetKind()+"/"+obj.GetName()+" error : "+err.Error())
		} else {
			res = append(res, obj.GetKind()+"/"+obj.GetName()+" applied successful.")
		}
	}
	if applyErr {
		return &utils.Response{Code: code.ApplyError, Msg: strings.Join(res, "\n")}
	}
	return &utils.Response{Code: code.Success, Msg: strings.Join(res, "\n")}
}

func (d *DynamicResource) buildDynamicResourceClient(data []byte) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
	// Decode YAML manifest into unstructured.Unstructured
	obj = &unstructured.Unstructured{}
	_, gvk, err := d.decUnstructured.Decode(data, nil, obj)
	if err != nil {
		return obj, dr, errors.Wrap(err, "Decode yaml failed. ")
	}

	// Find GVR
	mapping, err := d.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return obj, dr, errors.Wrap(err, "Mapping kind with version failed")
	}

	// Obtain REST interface for the GVR
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = d.DynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = d.DynamicClient.Resource(mapping.Resource)
	}
	return obj, dr, nil
}
