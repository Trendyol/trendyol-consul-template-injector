package main

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestPatchHelperVolumeAdd(t *testing.T) {
	p := Patches{}
	p.addVolumes(&corev1.Pod{}, []corev1.Volume{{Name: "Test"}})

	assert.Equal(t, 1, len(p))
}
