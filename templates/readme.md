# Templates

Creating templates is a very simple operation, even for those who know nothing at all about [Go Template](https://pkg.go.dev/text/template). All you should know is:

- The template should be one page. Without additional `css` or `js` files (but you can load them from the CDN or another GitHub repositories using [jsdelivr.com](https://www.jsdelivr.com/), for example)
- Don't forget to include `<meta name="robots" content="noindex, nofollow" />` tag in the header
- You can use a special "placeholders" (wrapped in `{{` and `}}`) for the rendering error code, message and others (see details below)

## Supported signatures

### Error page & request data

| Signature                            | Description                                                   | Example                                      |
|--------------------------------------|---------------------------------------------------------------|----------------------------------------------|
| `{{ code }}`                         | Error page code                                               | `404`                                        |
| `{{ message }}`                      | Error code message                                            | `Not found`                                  |
| `{{ description }}`                  | Error code description                                        | `The server can not find the requested page` |
| `{{ original_uri }}`                 | `X-Original-URI` header value                                 | `/foo1/bar2`                                 |
| `{{ namespace }}`                    | `X-Namespace` header value                                    | `foo`                                        |
| `{{ ingress_name }}`                 | `X-Ingress-Name` header value                                 | `bar`                                        |
| `{{ service_name }}`                 | `X-Service-Name` header value                                 | `baz`                                        |
| `{{ service_port }}`                 | `X-Service-Port` header value                                 | `8080`                                       |
| `{{ request_id }}`                   | `X-Request-ID` header value                                   | `12AB34CD56EF78`                             |
| `{{ forwarded_for }}`                | `X-Forwarded-For` header value                                | `203.0.113.195, 70.41.3.18`                  |
| `{{ host }}`                         | `Host` header value                                           | `example.com`                                |
| `{{ now.Unix }}`                     | Current timestamp (e.g. in Unix format)                       | `1643621927`                                 |
| `{{ hostname }}`                     | OS hostname                                                   | `ab12cd34ef56`                               |
| `{{ version }}`                      | Application version                                           | `2.5.0`                                      |
| `{{ if show_details }}...{{ end }}`  | Logical operator (server started with "show details" option?) |                                              |
| `{{ if hide_details }}...{{ end }}`  | Same as above, but inverted                                   |                                              |
| `{{ if l10n_enabled }}...{{ end }}`  | Logical operator (l10n is enabled?)                           |                                              |
| `{{ if l10n_disabled }}...{{ end }}` | Same as above, but inverted                                   |                                              |

### Modifiers

| Signature                          | Description                    | Example                             |
|------------------------------------|--------------------------------|-------------------------------------|
| <code>{{ ... &#124; json }}</code> | Convert value into json-string | <code>{{ code &#124; json }}</code> |
| <code>{{ ... &#124; int }}</code>  | Convert value into integer     | <code>{{ code &#124; int }}</code>  |
