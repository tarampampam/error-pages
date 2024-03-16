GET http://{{ host }}:{{ port }}/version

HTTP 200

[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.version" exists
jsonpath "$.version" isString
