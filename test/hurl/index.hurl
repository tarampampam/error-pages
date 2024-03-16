# HTML content
GET http://{{ host }}:{{ port }}/

HTTP 404

[Asserts]
header "Content-Type" contains "text/html"
body contains "404"
body contains "Not Found"

# JSON content
GET http://{{ host }}:{{ port }}/
Content-Type: text/json

HTTP 404

[Asserts]
header "Content-Type" contains "application/json"
header "X-Robots-Tag" == "noindex"
jsonpath "$.error" == true
jsonpath "$.code" == "404"
jsonpath "$.message" == "Not Found"

# XML content
GET http://{{ host }}:{{ port }}/
Content-Type: application/xml

HTTP 404

[Asserts]
header "Content-Type" contains "application/xml"
header "X-Robots-Tag" == "noindex"
xpath "string(//error/code)" == "404"
xpath "string(//error/message)" == "Not Found"
