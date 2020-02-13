package main

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestShouldThrowErrorWhenAnnotationValueIsNotParseable(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"trendyol.com/consul-template-inject": "hello",
			},
		},
	}

	_, err := newPodAnnotations(pod)

	assert.EqualError(t, err, "")
}

func TestPreCheckFunctionalityOfPodAnnotations(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"trendyol.com/consul-template-inject":      "true",
				"trendyol.com/consul-template-consul-addr": "testadress",
			},
		},
	}

	podAnnotations, err := newPodAnnotations(pod)

	assert.NoError(t, err)
	assert.Equal(t, podAnnotations["trendyol.com/consul-template-inject"].Value, "true")
	assert.Equal(t, podAnnotations["trendyol.com/consul-template-consul-addr"].Value, "testadress")
}
