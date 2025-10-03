# 🔤 Localization

This directory contains the file [l10n.js](l10n.js) for localizing error pages. Once the error page is loaded,
this script runs and translates the page content to the user's locale.

> [!NOTE]
> In version `2.*`, the working logic was simpler: error pages loaded this script using
> [jsdelivr.com](https://www.jsdelivr.com/) as a CDN for
> [versioned content from the GitHub repository](https://www.jsdelivr.com/features#gh), and it translated
> tag content with the special HTML attribute `data-l10n`.

By default, the error page markup contains strings in English (`en` locale). To localize the error pages to
different locales, please follow these steps:

1. Find your locale name on [this page](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) (column `Set 1` or `ISO 639-1:2002`)
2. Fork this repository
3. Edit the file [l10n.js](l10n.js) in the `data` map (append new localized strings) using the locale name from step 1
4. Please add your locale to the [playground.html](playground.html) file to test the localization
5. Make a PR with your changes

## 👍 Translators

- 🇫🇷 French by [@jvin042](https://github.com/jvin042)
- 🇵🇹 Portuguese by [@fabtrompet](https://github.com/fabtrompet)
- 🇳🇱 Dutch by [@SchoNie](https://github.com/SchoNie)
- 🇩🇪 German by [@mschoeffmann](https://github.com/mschoeffmann)
- 🇪🇸 Spanish by [@Runig006](https://github.com/Runig006)
- 🇨🇳 Chinese by [@CDN18](https://github.com/CDN18)
- 🇮🇩 Indonesian by [@getwisp](https://github.com/getwisp)
- 🇵🇱 Polish by [@wielorzeczownik](https://github.com/wielorzeczownik)
- 🇰🇷 Korean by [@NavyStack](https://github.com/NavyStack)
- 🇭🇺 Hungarian by [@oszto90](https://github.com/oszto90)
- 🇳🇴 Norwegian by [@EliasTors](https://github.com/EliasTors)
- 🇷🇴 Romanian by [@pasarenicu](https://github.com/pasarenicu)
