# 📝 Templating and Localization

Templates are parsed **once at startup**. The same template engine is used for all output formats. When a custom
HTML template is set via `--html-template`, the `--template-name` and `--rotation-mode` flags are ignored.

### Template data

All templates receive a data object with the following fields:

| Field                        | Type     | Description                                                            |
|------------------------------|----------|------------------------------------------------------------------------|
| `.StatusCode`                | `uint16` | HTTP status code (e.g. `404`)                                          |
| `.Message`                   | `string` | Short status text (e.g. `Not Found`)                                   |
| `.Description`               | `string` | Longer description (e.g. `The server can not find the requested page`) |
| `.OriginalURI`               | `string` | Request URI that caused the error *                                    |
| `.Namespace`                 | `string` | Kubernetes namespace of the backend service *                          |
| `.IngressName`               | `string` | Name of the Ingress resource *                                         |
| `.ServiceName`               | `string` | Name of the backend service *                                          |
| `.ServicePort`               | `string` | Port of the backend service *                                          |
| `.RequestID`                 | `string` | Unique request ID *                                                    |
| `.ForwardedFor`              | `string` | Original client IP(s) from `X-Forwarded-For` *                         |
| `.Host`                      | `string` | Request `Host` header *                                                |
| `.Config.ShowRequestDetails` | `bool`   | Whether `--show-details` is enabled                                    |
| `.Config.L10nDisabled`       | `bool`   | Whether `--disable-l10n` is set                                        |

> `*` - Requires `--show-details`

In addition to the fields above, templates also have access to a set of built-in functions (see below), which are
pipeline-friendly (needle before haystack): `{{ .Message | default "Unknown" | upper }}`.

| Function                     | Description                                      | Example                                                        |
|------------------------------|--------------------------------------------------|----------------------------------------------------------------|
| `now`                        | Current time (`time.Time`)                       | `{{ now.Format "2006-01-02" }}` / `{{ now.Unix }}`             |
| `hostname`                   | Server hostname                                  | `{{ hostname }}`                                               |
| `version`                    | Application version string                       | `{{ version }}`                                                |
| `env "KEY"`                  | Env var value (sensitive keys masked with `***`) | `{{ env "STAGE" }}`                                            |
| `toJson` / `toJSON`          | JSON-encode a value                              | `{{ .Message \| toJson }}`                                     |
| `toInt` / `int`              | Convert to integer                               | `{{ .StatusCode \| int }}`                                     |
| `toString` / `str`           | Convert to string                                | `{{ .StatusCode \| str }}`                                     |
| `escape`                     | HTML-escape                                      | `{{ .OriginalURI \| escape }}`                                 |
| `urlEncode`                  | URL-encode                                       | `{{ .OriginalURI \| urlEncode }}`                              |
| `trim`                       | Strip leading/trailing whitespace                | `{{ .Message \| trim }}`                                       |
| `trimPrefix`                 | Remove prefix                                    | `{{ .Message \| trimPrefix "Error: " }}`                       |
| `trimSuffix` / `trimPostfix` | Remove suffix                                    | `{{ .Message \| trimSuffix "!" }}`                             |
| `trimAll`                    | Strip specific characters                        | `{{ ".test." \| trimAll "." }}`                                |
| `lower` / `upper`            | Change case                                      | `{{ .Message \| upper }}`                                      |
| `replace`                    | Replace all occurrences                          | `{{ .Message \| replace " " "_" }}`                            |
| `contains`                   | Substring check                                  | `{{ .Message \| contains "test" }}...{{ end }}`                |
| `hasPrefix` / `hasSuffix`    | Prefix/suffix check                              | `{{ .Message \| hasPrefix "test" }}...{{ end }}`               |
| `split`                      | Split string by separator                        | `{{ split ";" "a;b;c" }}`                                      |
| `join`                       | Join slice with separator                        | `{{ split ";" "a;b;c" \| join ", " }}`                         |
| `fields`                     | Split string by whitespace                       | `{{ fields "foo bar baz" \| join "-" }}`                       |
| `substr`                     | Substring by rune index and length               | `{{ "Hello, World!" \| substr 7 5 }}`                          |
| `truncate`                   | Truncate with `...` appended                     | `{{ .Description \| truncate 80 }}`                            |
| `repeat`                     | Repeat string N times                            | `{{ "Ha" \| repeat 3 }}`                                       |
| `quote` / `squote`           | Wrap in double/single quotes                     | `{{ .Message \| quote }}`                                      |
| `count`                      | Count substring occurrences                      | `{{ "test" \| count "t" }}`                                    |
| `default`                    | Fallback value for empty input                   | `{{ .OriginalURI \| default "N/A" }}`                          |
| `coalesce`                   | First non-empty value from a list                | `{{ coalesce .Message .Description "error" }}`                 |
| `ternary`                    | Inline conditional                               | `{{ .Config.ShowRequestDetails \| ternary "shown" "hidden" }}` |
| `isEmpty` / `isNotEmpty`     | Emptiness check                                  | `{{ if isNotEmpty .Description }}...{{ end }}`                 |
| `l10nScript`                 | Inline the localization JS script                | `<script>{{ l10nScript }}</script>`                            |

> [!NOTE]
> `env` masks values whose key (split by `_`) contains `PASSWORD`, `SECRET`, `KEY`, `TOKEN`, `PASS`, `PWD`,
> or `CRED` (case-insensitive). Those calls return a string of `*` characters instead of the actual value.

### Localization

HTML pages support automatic **client-side localization** in 15+ languages, templates that want l10n must:

1. Add `data-l10n` attributes to elements whose text content should be translated
2. Include `<script>{{ l10nScript }}</script>` in the HTML

The browser detects the visitor's preferred language via `navigator.languages` and translates all `[data-l10n]`
elements in-place - no server round-trip required. Localization can be disabled globally with `--disable-l10n`
(or via env `DISABLE_L10N=true`).
