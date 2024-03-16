# Hurl

Hurl is a command line tool that runs **HTTP requests** defined in a simple **plain text format**.

## How to use

It can perform requests, capture values and evaluate queries on headers and body response. Hurl is very versatile: it can be used for both fetching data and testing HTTP sessions.

```hurl
# Get home:
GET https://example.net

HTTP 200
[Captures]
csrf_token: xpath "string(//meta[@name='_csrf_token']/@content)"

# Do login!
POST https://example.net/login?user=toto&password=1234
X-CSRF-TOKEN: {{csrf_token}}

HTTP 302
```

### Links:

- [Official website](https://hurl.dev/)
- [GitHub](https://github.com/Orange-OpenSource/hurl)
