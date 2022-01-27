# The common request
GET http://{{ host }}:{{ port }}/502.html
Content-Type: application/json;charset=UTF-8
X-Original-URI: foo
X-Namespace: bar
X-Ingress-Name: baz
X-Service-Name: aaa
X-Service-Port: bbb
X-Request-ID: ccc

HTTP/* 200

[Asserts]
header "Content-Type" contains "application/json"
header "X-Robots-Tag" == "noindex"
jsonpath "$.error" == true
jsonpath "$.code" == "502"
jsonpath "$.message" == "Bad Gateway"
jsonpath "$.description" == "The server received an invalid response from the upstream server"
jsonpath "$.details.original_uri" == "foo"
jsonpath "$.details.namespace" == "bar"
jsonpath "$.details.ingress_name" == "baz"
jsonpath "$.details.service_name" == "aaa"
jsonpath "$.details.service_port" == "bbb"
jsonpath "$.details.request_id" == "ccc"
jsonpath "$.details.timestamp" isInteger

# X-Format in the action
GET http://{{ host }}:{{ port }}/502.html
X-Format: text/json

HTTP/* 200

[Asserts]
header "Content-Type" contains "application/json"
header "X-Robots-Tag" == "noindex"
jsonpath "$.error" == true
jsonpath "$.code" == "502"
jsonpath "$.message" == "Bad Gateway"
