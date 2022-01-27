GET http://{{ host }}:{{ port }}/healthz

HTTP/* 200

```OK```

# Next endpoint marked as deprecated
GET http://{{ host }}:{{ port }}/health/live

HTTP/* 200

```OK```
