GET http://{{ host }}:{{ port }}/metrics

HTTP/* 200

[Asserts]
header "Content-Type" contains "text/plain"
body contains "http_request_duration_millisecond"
body contains "http_request_total_count"
