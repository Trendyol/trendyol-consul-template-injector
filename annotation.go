package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"log"
	"strconv"
)

type annotationPreValidationType func(i string) error

type annotation struct {
	Value         string
	PreValidation annotationPreValidationType
}

func (a *annotation) shouldDoPreCheck() bool {
	return a.PreValidation != nil
}

type podAnnotations map[string]*annotation

func newPodAnnotations(p *corev1.Pod) (podAnnotations, error) {
	annotations := podAnnotations{
		ConsulTemplateInjectAnnotation: &annotation{Value: "false", PreValidation: func(i string) error {
			if _, err := strconv.ParseBool(i); err != nil {
				return fmt.Errorf("could not parse 'trendyol.com/consul-template-inject' annotation's value,"+
					" \"%s\" is not valid, please use 'true' or 'false'", i)
			}
			return nil
		}},
		ConsulTemplateConsulAddressAnnotation:         &annotation{Value: "consul-consul-server.default:8500"},
		ConsulTemplateFilePathAnnotation:              &annotation{Value: "/out/output.txt"},
		ConsulTemplateTemplateConfigMapNameAnnotation: &annotation{Value: "consul-template-cm"},
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
			if av.shouldDoPreCheck() {
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
