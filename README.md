<h1 align="center">Welcome to trendyol-consul-template-injector 👋</h1>
<p>
  <img alt="Version" src="https://img.shields.io/badge/version-0.0.1-blue.svg?cacheSeconds=2592000" />
  <a href="https://github.com/Trendyol/trendyol-consul-template-injector" target="_blank">
    <img alt="Documentation" src="https://img.shields.io/badge/documentation-yes-brightgreen.svg" />
  </a>
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>
</p>

> This projects is an implementation of &#34;Admission Webhook Controllers&#34; concept in Kubernetes , checkout this page to get more detail about admission https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/   

> trendyol-consul-template-injector injects trendyol-consul-template image (https://hub.docker.com/r/trendyoltech/trendyol-consul-template) as a sidecar container into your application's pod.

> It is configured using these annotations:
* "trendyol.com/consul-template-inject" -> "true" for enabling sidecar injection
* "trendyol.com/consul-template-consul-addr" -> consul url
* "trendyol.com/consul-template-template-config-map-name" -> configmap name which contains configuration template file which is needed by consul-template for rendering configuration file for your application
* "trendyol.com/consul-template-output-file" -> path to configuration file which is rendered by the consul-template

> Example pod with annotations and config map are given in examples folder.

### 🏠 [Homepage](https://github.com/Trendyol/trendyol-consul-template-injector)

### ✨ [Demo](https://github.com/Trendyol/trendyol-consul-template-injector)

## Author

👤 **Trendyol**

* Website: https://github.com/Trendyol/trendyol-consul-template-injector
* Github: [@Trendyol](https://github.com/Trendyol)

## Show your support

Give a ⭐️ if this project helped you!

***
_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_