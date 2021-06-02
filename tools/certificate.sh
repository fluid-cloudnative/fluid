#!/bin/sh
set -e


usage() {
    cat <<EOF
Generate certificate suitable for use with an admission controller service.
This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with webhook
services. This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explanation and additional instructions.
The server key/cert k8s CA cert are stored in a k8s secret.
usage: ${0} [OPTIONS]
The following flags are required.
       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
       --certDir          Dir to store certificate
EOF
    exit 0
}


while [ $# -gt 0 ]; do
    case ${1} in
        --service)
            SERVICE="$2"
            shift
            ;;
        --namespace)
            NAMESPACE="$2"
            shift
            ;;
        --certDir)
            CERT_DIR="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done


if [ -z "${SERVICE}" ]; then
    echo "'--service' must be specified"
    exit 1
fi

if [ -z "${NAMESPACE}" ]; then
    echo "'--namespace' must be specified"
    exit 1
fi

if [ -z "${CERT_DIR}" ]; then
    echo "'--certDir' must be specified"
    exit 1
fi

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

echo "creating certs in dir ${CERT_DIR} "

cat <<EOF > ${CERT_DIR}/csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE}
DNS.2 = ${SERVICE}.${NAMESPACE}
DNS.3 = ${SERVICE}.${NAMESPACE}.svc
EOF


openssl genrsa -out ${CERT_DIR}/ca.key 2048
openssl req -x509 -new -nodes -key ${CERT_DIR}/ca.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -days 36500 -out ${CERT_DIR}/ca.crt


openssl genrsa -out ${CERT_DIR}/tls.key 2048
openssl req -new -key ${CERT_DIR}/tls.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -days 36500 -out ${CERT_DIR}/tls.csr -config ${CERT_DIR}/csr.conf


openssl x509 -req -in  ${CERT_DIR}/tls.csr -CA  ${CERT_DIR}/ca.crt -CAkey  ${CERT_DIR}/ca.key \
  -CAcreateserial -out  ${CERT_DIR}/tls.crt \
  -extensions v3_req -extfile  ${CERT_DIR}/csr.conf
