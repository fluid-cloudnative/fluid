## How to Generate CRD Api Reference Docs
Currently, we are using a doc generator named *gen-crd-api-reference-docs* to generate our API Reference Docs. You can find its project [here](https://github.com/ahmetb/gen-crd-api-reference-docs) on Github.

However, there might be some compatibility issue with *Kubebuilder* which we used for generating CRD template codes. Check [here](https://github.com/ahmetb/gen-crd-api-reference-docs/issues/15) for more information about the compatibility issue.

To generate CRD Api reference docs, some tricks need to be done to bypass the issue mentioned above. This guide will demonstrate what should be done to generate CRD API Reference docs.

### Prerequisite
Make sure you have the binary version of the doc generator in your environment:

Download releases directly: https://github.com/ahmetb/gen-crd-api-reference-docs/releases

And you shall find the following files(or directory) after decompression
```shell script
├── example-config.json
├── gen-crd-api-reference-docs
├── LICENSE
└── template
    ├── members.tpl
    ├── pkg.tpl
    └── type.tpl
```
Put `gen-crd-api-reference-docs` under your `$GOPATH/bin`. Copies of all the other files except `LICENSE` can be found in Fluid Project(under `tools/api-doc-gen`).

### Specify API Package to be generated for

Assume path of the your CRD API Package is `api/v1alpha1`.

Add a `doc.go` under the path `api/v1alpha1` like:
```go
// <Some comments about v1alpha1>
// +groupName=<your-group-name>
package v1alpha1
```

### Specify CRDs which need to be exported 
Take type `Dataset` in `api/v1alpha1/dataset_types.go` as an example, it's a root crd object.

To export the CRD, add a new comment line with `+genclient`:
```
  // +kubebuilder:object:root=true
  // +kubebuilder:subresource:status
+ // +genclient 

  // Dataset is the Schema for the datasets API
  type Dataset struct {
```

### Run the provided script to generate api docs
```shell script
# tools/api-doc-gen/generate_api_doc.sh
#! /bin/sh

DIR=$(dirname $0)
cd $DIR
GOPATH=$(go env GOPATH)
$GOPATH/bin/gen-crd-api-reference-docs \
  --config ./example-config.json \
  --template-dir ./template \
  --api-dir ../../api \
  --out-file ./api_doc.md
```

Some explanation for arguments above:
- config: which config should be used
- template-dir: templates(*.tpl) under the directory will be used to generate docs  
- api-dir: where to find all the CRDs (e.g. in our situation, `api/v1alpha1`)
- out-file: path and filename of the generated doc