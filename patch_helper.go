package main

import (
	corev1 "k8s.io/api/core/v1"
	"log"
)

type Patches []patchOperation

func (p Patches) patchReport() {
	log.Printf("--------------APPLYING PATCHES ARE----------------------")
	for _, patch := range p {
		log.Printf("Operation: %s \n", patch.Op)
		log.Printf("Path: %s \n", patch.Path)
		log.Printf("Value: %s \n", patch.Value)
	}
	log.Printf("--------------------------------------------------------")
}

func (p *Patches) addVolumes(pod *corev1.Pod, volumes []corev1.Volume) {
	first := len(pod.Spec.Volumes) == 0
	path := "/spec/volumes"
	var value interface{}
	for _, v := range volumes {
		value = v
		tempPath := path
		if first {
			first = false
			value = []corev1.Volume{v}
		} else {
			tempPath = path + "/-"
		}

		*p = append(*p, patchOperation{
			Op:    "add",
			Path:  tempPath,
			Value: value,
		})
	}
}

func (p *Patches) addContainers(pod *corev1.Pod, containers []corev1.Container) {
	first := len(pod.Spec.Containers) == 0
	path := "/spec/containers"
	var value interface{}
	for _, c := range containers {
		value = c
		tempPath := path
		if first {
			first = false
			value = []corev1.Container{c}
		} else {
			tempPath = path + "/-"
		}

		*p = append(*p, patchOperation{
			Op:    "add",
			Path:  tempPath,
			Value: value,
		})
	}
}

func (p *Patches) addInitContainers(pod *corev1.Pod, containers []corev1.Container) {
	first := len(pod.Spec.InitContainers) == 0
	path := "/spec/initContainers"
	var value interface{}
	for _, c := range containers {
		value = c
		tempPath := path
		if first {
			first = false
			value = []corev1.Container{c}
		} else {
			tempPath = path + "/-"
		}

		*p = append(*p, patchOperation{
			Op:    "add",
			Path:  tempPath,
			Value: value,
		})
	}
}

func (p *Patches) addVolumeMounts(pod *corev1.Pod, vms []corev1.VolumeMount) {
	first := len(pod.Spec.Containers[0].VolumeMounts) == 0
	path := "/spec/containers/0/volumeMounts"
	var value interface{}
	for _, vm := range vms {
		value = vm
		tempPath := path
		if first {
			first = false
			value = []corev1.VolumeMount{vm}
		} else {
			tempPath = path + "/-"
		}

		*p = append(*p, patchOperation{
			Op:    "add",
			Path:  tempPath,
			Value: value,
		})
	}
}
