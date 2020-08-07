#! /bin/sh

DIR=$(dirname $0)
cd $DIR
GOPATH=$(go env GOPATH)
$GOPATH/bin/gen-crd-api-reference-docs \
  --config ./example-config.json \
  --template-dir ./template \
  --api-dir ../../api \
  --out-file ./api_doc.md
