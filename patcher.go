package main

import (
	"fmt"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strings"
)

const (
	ConsulTemplateInjectAnnotation                = "trendyol.com/consul-template-inject"
	ConsulTemplateConsulAddressAnnotation         = "trendyol.com/consul-template-consul-addr"
	ConsulTemplateFilePathAnnotation              = "trendyol.com/consul-template-output-file"
	ConsulTemplateTemplateConfigMapNameAnnotation = "trendyol.com/consul-template-template-config-map-name"
)

var (
	podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods",}
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func generatePodPatches(req *v1beta1.AdmissionRequest) (patches, error) {
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, fmt.Errorf("expect resource to be %s", podResource)
	}

	rawObj := req.Object.Raw
	pod := new(corev1.Pod)

	if _, _, err := universalDeserializer.Decode(rawObj, nil, pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	podAnnotations, err := newPodAnnotations(pod)

	if err != nil {
		return nil, err
	}

	patches := patches{}

	templateFile := "/conf/init.tmpl"
	outputFile := podAnnotations[ConsulTemplateFilePathAnnotation].Value
	templateConfigMapName := podAnnotations[ConsulTemplateTemplateConfigMapNameAnnotation].Value
	consulAddr := podAnnotations[ConsulTemplateConsulAddressAnnotation].Value

	lastSlashIndex := strings.LastIndex(outputFile, "/")
	outputFolderPath := outputFile[:lastSlashIndex]

	patches.addVolumes(pod, []corev1.Volume{
		{
			Name: "consul-template-input",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: templateConfigMapName,
					},
				},
			},
		},
		{
			Name: "consul-template-output",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		},
	}, )

	patches.addVolumeMounts(pod, []corev1.VolumeMount{
		{
			Name:      "consul-template-output",
			MountPath: outputFolderPath,
		},
	}, )

	patches.addInitContainers(pod, []corev1.Container{
		{
			Name:  "trendyol-consul-template-init",
			Image: "trendyoltech/trendyol-consul-template:latest",
			Env: []corev1.EnvVar{
				{Name: "CONSUL_TEMPLATE_PROCESS_FLAGS", Value: "-once"},
				{Name: "CONSUL_ADDR", Value: consulAddr},
				{Name: "CONSUL_TEMPLATE_TEMPLATE_PATH", Value: templateFile},
				{Name: "CONSUL_TEMPLATE_OUTPUT_PATH", Value: outputFile},
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "consul-template-input", MountPath: "/conf"},
				{Name: "consul-template-output", MountPath: outputFolderPath},
			},
			ImagePullPolicy: "IfNotPresent",
		},
	}, )

	patches.addContainers(pod, []corev1.Container{
		{
			Name:  "trendyol-consul-template-sidecar",
			Image: "trendyoltech/trendyol-consul-template:latest",
			Env: []corev1.EnvVar{
				{Name: "CONSUL_ADDR", Value: consulAddr},
				{Name: "CONSUL_TEMPLATE_TEMPLATE_PATH", Value: templateFile},
				{Name: "CONSUL_TEMPLATE_OUTPUT_PATH", Value: outputFile},
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "consul-template-input", MountPath: "/conf"},
				{Name: "consul-template-output", MountPath: outputFolderPath},
			},
			ImagePullPolicy: "IfNotPresent",
		},
	}, )

	return patches, nil
}
