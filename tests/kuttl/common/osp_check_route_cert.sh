#!/bin/bash

ROUTE_NAME=$1

if [[ "$ROUTE_NAME" == "barbican" ]]; then
    EXPECTED_CERTIFICATE="-----BEGIN CERTIFICATE-----
MIIBfjCCASUCFB2bFw2MfRB0vIAZmNe81aRJIj8GMAoGCCqGSM49BAMCMEIxCzAJ
BgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1
bHQgQ29tcGFueSBMdGQwHhcNMjQwNDI5MDc0NzM4WhcNMjUwNDI5MDc0NzM4WjBC
MQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNE
ZWZhdWx0IENvbXBhbnkgTHRkMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExuPx
YX5SrJpymPX/j67XXiHiSte8S6wzbuJd0xM4HODaZbt0r0IWO8rRQUEeLnudFoF6
Ju4R3QUT8d+t3zDZqTAKBggqhkjOPQQDAgNHADBEAiBYsG23KRf/3o+R9+imWJa4
8d2zu9eMr6qkTg5Q4tj/sgIgLaox7QK2Ao4dY3nmlyg9ascsQNKV4bpdSXAX/drH
7DE=
-----END CERTIFICATE-----"

    EXPECTED_CA_CERTIFICATE="-----BEGIN CERTIFICATE-----
MIIB2DCCAX+gAwIBAgIUWARgHwuaus9w7uQ1opRlDXRPRvswCgYIKoZIzj0EAwIw
QjELMAkGA1UEBhMCWFgxFTATBgNVBAcMDERlZmF1bHQgQ2l0eTEcMBoGA1UECgwT
RGVmYXVsdCBDb21wYW55IEx0ZDAeFw0yNDA0MjkwNzQ2NDdaFw0yNTA0MjkwNzQ2
NDdaMEIxCzAJBgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNV
BAoME0RlZmF1bHQgQ29tcGFueSBMdGQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AATg695MQRE7erKhHfNL0MA+2bve1TruxjNm5kUJaCk8CbQWififn9WjHlPnND6x
ftxDyis5MyTvdjpYnzZTxSEwo1MwUTAdBgNVHQ4EFgQUVkjtNnURJ4c+kvZn7/fF
EJhMz7QwHwYDVR0jBBgwFoAUVkjtNnURJ4c+kvZn7/fFEJhMz7QwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiAuExPY3JFBuZ1IFv5Sf+Ai3YHwtoSb
PzKC6Dq0YRpNqgIgTTK9yfJpiJSHwJqa2dWnD8dJkf8jKH+6/VJrRk3mpkY=
-----END CERTIFICATE-----"
elif [[ "$ROUTE_NAME" == "placement" ]]; then
    EXPECTED_CERTIFICATE="CERT123"
    EXPECTED_CA_CERTIFICATE="CACERT123"
fi

TLS_DATA=$(oc get route ${ROUTE_NAME}-public -n $NAMESPACE -o jsonpath='{.spec.tls}')

# Extract certificates from the route
ACTUAL_CERTIFICATE=$(echo "$TLS_DATA" | jq -r '.certificate')
ACTUAL_CA_CERTIFICATE=$(echo "$TLS_DATA" | jq -r '.caCertificate')

TRIMMED_EXPECTED_CERTIFICATE=$(echo "$EXPECTED_CERTIFICATE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
TRIMMED_ACTUAL_CERTIFICATE=$(echo "$ACTUAL_CERTIFICATE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

TRIMMED_EXPECTED_CA_CERTIFICATE=$(echo "$EXPECTED_CA_CERTIFICATE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
TRIMMED_ACTUAL_CA_CERTIFICATE=$(echo "$ACTUAL_CA_CERTIFICATE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

if [[ "$TRIMMED_EXPECTED_CERTIFICATE" != "$TRIMMED_ACTUAL_CERTIFICATE" ]]; then
    echo "Certificate does not match for route $ROUTE_NAME in namespace $NAMESPACE."
    exit 1
fi

if [[ "$TRIMMED_EXPECTED_CA_CERTIFICATE" != "$TRIMMED_ACTUAL_CA_CERTIFICATE" ]]; then
    echo "CA Certificate does not match for route $ROUTE_NAME in namespace $NAMESPACE."
    exit 1
fi

echo "TLS data matches for route $ROUTE_NAME in namespace $NAMESPACE."
