# 🔤 Localization

[![jsDelivr hits](https://img.shields.io/jsdelivr/gh/hm/tarampampam/error-pages)](https://www.jsdelivr.com/package/gh/tarampampam/error-pages)

This directory contains file [l10n.js](l10n.js) for the error pages localization. The working logic is very simple - pages load this script using [jsdelivr.com](https://www.jsdelivr.com/) as a CDN for [versioned content from the GitHub repository](https://www.jsdelivr.com/features#gh), and it translates tag content with the special HTML attribute `data-l10n`.

By default, pages markup contains strings in English (`en` locale). If you want to localize the error pages on the different locales, you should:

- Find your locale name on [this page](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) (column `639-1`)
- Make a fork of this repository
- Edit file [l10n.js](l10n.js) in `data` section (append new localized strings) using locale name from the step 1
- Make a PR with your changes

## 👍 Translators

- 🇫🇷 French by [@jvin042](https://github.com/jvin042)
