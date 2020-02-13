package main

import (
	"fmt"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strconv"
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

type AnnotationPreValidationType func(i string) error

type Annotation struct {
	Value         string
	PreValidation AnnotationPreValidationType
}

func (a *Annotation) ShouldDoPreCheck() bool {
	return a.PreValidation != nil
}

type PodAnnotations map[string]*Annotation

func New(p *corev1.Pod) (PodAnnotations, error) {
	annotations := PodAnnotations{
		ConsulTemplateInjectAnnotation: &Annotation{Value: "false", PreValidation: func(i string) error {
			if _, err := strconv.ParseBool(i); err != nil {
				return fmt.Errorf("could not parse 'trendyol.com/consul-template-inject' annotation's value,"+
					" \"%s\" is not valid, please use 'true' or 'false'", i)
			}
			return nil
		}},
		ConsulTemplateConsulAddressAnnotation:         &Annotation{Value: "consul-consul-server.default:8500"},
		ConsulTemplateFilePathAnnotation:              &Annotation{Value: "/out/output.txt"},
		ConsulTemplateTemplateConfigMapNameAnnotation: &Annotation{Value: "consul-template-cm"},
	}

	incomingPodAnnotations := p.ObjectMeta.Annotations
	log.Println("-------------INCOMING POD ANNOTATIONS------------------")
	for key, val := range incomingPodAnnotations {
		log.Printf("Key: %s \n", key)
		log.Printf("Value: %s \n", val)
	}
	log.Println("-------------------------------------------------------")

	for k, v := range incomingPodAnnotations {
		if av, exists := annotations[k]; exists {
			if av.ShouldDoPreCheck() {
				err := av.PreValidation(v)
				if err != nil {
					return nil, err
				}
			}
			av.Value = v
		}
	}

	log.Println("-------------PREPARED POD ANNOTATIONS------------------")
	for key, val := range annotations {
		log.Printf("Key: %s \n", key)
		log.Printf("Value: %s \n", val)
	}
	log.Println("-------------------------------------------------------")

	return annotations, nil
}

func generatePodPatches(req *v1beta1.AdmissionRequest) (Patches, error) {
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, fmt.Errorf("expect resource to be %s", podResource)
	}

	rawObj := req.Object.Raw
	pod := new(corev1.Pod)

	if _, _, err := universalDeserializer.Decode(rawObj, nil, pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	podAnnotations, err := New(pod)

	if err != nil {
		return nil, err
	}

	patches := Patches{}

	templatePath := "/conf/init.tmpl"
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
				{Name: "CONSUL_TEMPLATE_TEMPLATE_PATH", Value: templatePath},
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
				{Name: "CONSUL_TEMPLATE_TEMPLATE_PATH", Value: templatePath},
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
