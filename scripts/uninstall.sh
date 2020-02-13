#!/usr/bin/env bash

PROJECT=trendyol-consul-template-injector
NAMESPACE=admission
manifestsBasedir="manifests"

kubectl delete secrets $PROJECT-server-tls-secret --namespace $NAMESPACE
kubectl delete -f $manifestsBasedir/ --force --grace-period 0
kubectl delete configmaps busybox-consul-template-cm --force --grace-period 0 --namespace $NAMESPACE
kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io trendyol-consul-template-injector-webhook --namespace $NAMESPACE
kubectl delete pods busybox --namespace $NAMESPACE --force --grace-period 0
