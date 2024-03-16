# disabled until https://github.com/Orange-OpenSource/hurl/issues/2540 is not fixed

#GET http://{{ host }}:{{ port }}/metrics
#
#HTTP 200
#
#[Asserts]
#header "Content-Type" contains "text/plain"
#body contains "http_requests_duration_millisecond"
#body contains "http_requests_total_count"
