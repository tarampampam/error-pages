(() => {
  'use strict';

  /**
   * Tokenizes a string by lowercasing it and removing all non-alphanumeric characters.
   *
   * Tokenization is used to flexibly match raw text with localization tokens, allowing for variations in
   * whitespace and punctuation.
   *
   * @param {string} s
   * @returns {string}
   */
  const t = (s) => s.toLowerCase().replace(/[^a-z0-9]+/g, '');

  /**
   * Two-level map of localization tokens and their translations.
   * The first level maps the token (usually raw text in English) to a map of language codes (always in lowercase)
   * and their translations.
   *
   * @type {Map<string, Map<string, string>>}
   */
  const locales = new Map([
    {{- range .Tokens }}
    [t({{.Key | quote}}), new Map([
      {{- range .Translations}}
      [{{ .LangCode | quote }}, {{ .Value | quote }}],
      {{- end}}
    ])],
    {{- end}}
  ]);

  /**
   * Set of supported language codes for quick lookup (each code is always in lowercase).
   *
   * @type {Set<string>}
   */
  const supported = new Set([
    {{- range $i, $code := .SupportedLangCodes -}}
      {{- if $i }}, {{ end }}{{ $code | quote }}
    {{- end -}}
  ]);

  /**
   * Resolves a BCP 47 language tag using the provided lookup, trying an exact match first, then the base language
   * (e.g., 'zh-TW' → 'zh'). Returns the resolved code, or null if neither matches.
   *
   * @param {string} lang - lowercase BCP 47 language tag
   * @param {function(string): boolean} has - returns true if the code is available
   * @returns {string|null}
   */
  const resolveLang = (lang, has) => {
    if (has(lang)) { return lang; }
    const base = lang.split('-')[0];
    return (base !== lang && has(base)) ? base : null;
  };

  /**
   * Translates the given raw text to the provided language, if possible. If not, this will return null.
   *
   * @param {string} text - the raw text to translate (usually in English)
   * @param {string} language - the language code to translate to (will be normalized)
   * @returns {string|null}
   */
  const translateText = (text, language) => {
    if (!language) {
      return null; // no language to translate to
    }

    const lang = language.trim().toLowerCase();

    const translations = locales.get(t(text));
    if (!translations || !translations.size) {
      return null; // no translations for this text
    }

    const resolved = resolveLang(lang, (l) => translations.has(l));
    return (resolved && translations.get(resolved)) || null;
  }

  // attribute name and selector for elements that can be localized
  const L10N_ATTR = 'data-l10n';
  const L10N_SELECTOR = '[' + L10N_ATTR + ']';

  /**
   * Localizes the given element by translating its text content to the provided language, if possible. The original
   * text is stored in the data-l10n attribute (if not already stored) to allow for re-localization when the language
   * changes.
   *
   * @param {Element} el
   * @param {string} language
   * @returns {boolean}
   */
  const localizeElement = (el, language) => {
    if (!el || el.nodeType !== 1) {
      return false; // process only element nodes
    }

    if (!el.hasAttribute(L10N_ATTR)) {
      return false; // note has no data-l10n attribute
    }

    const fromAttribute = el.getAttribute(L10N_ATTR) || null;

    // the original raw text may be stored in the data attribute (if we already localized this element before), or
    // read from the element's textContent (it means this is the first time we localize this element)
    const elementText = fromAttribute ?? el.textContent ?? null;
    if (!elementText) {
      return false; // no text to translate
    }

    // try to translate the text, if we have a translation for the current language
    const localized = translateText(elementText, language);
    if (localized) {
      if (!fromAttribute) {
        el.setAttribute(L10N_ATTR, elementText); // promote once, read forever
      }

      el.textContent = localized;
    } else {
      if (fromAttribute) {
        el.textContent = fromAttribute; // restore original when no translation available (e.g., switching back to English)
      } else if (language) {
        console.debug('[l10n] Unable to localize element', el, 'to language', language);
      }

      return false; // no translation available
    }

    return true;
  }

  /**
   * Localizes the entire document by localizing all elements with the data-l10n attribute to the provided language.
   *
   * @param {string|null} language
   */
  const localizeDocument = (language) => {
    if (!language) {
      return; // no language provided, do not localize
    }

    document.querySelectorAll(L10N_SELECTOR).forEach((el) => localizeElement(el, language));
    document.documentElement.setAttribute('lang', language);
  }

  /**
   * Determines the best language to translate to based on the user's browser settings and supported languages.
   * If no supported language is found - this will be null (the default page language (English) will be used).
   *
   * In other words - here we decide which language to use for the page.
   *
   * @type {string|null}
   */
  let translateTo = (() => {
    for (const lang of (navigator.languages || []).map((l) => l.toLowerCase())) {
      if (supported.has(lang)) { // quick exact match
        return lang;
      }

      // since lang is BCP 47 language tag, we can try to match the base language (e.g., 'en' from 'en-US')
      const base = lang.split('-')[0];
      if (base !== lang && supported.has(base)) {
        return base;
      }
    }

    return null;
  })();

  // start observing the document for new elements with the data-l10n attribute (required for localizing dynamically
  // added content)
  new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (node.nodeType !== 1) {
          continue; // process only element nodes
        }

        if (node.hasAttribute(L10N_ATTR)) {
          localizeElement(node, translateTo);
        }

        node.querySelectorAll(L10N_SELECTOR).forEach((el) => localizeElement(el, translateTo));
      }
    }
  }).observe(document.documentElement, {childList: true, subtree: true});

  Object.defineProperty(window, 'l10n', {
    value: Object.freeze({
      setLocale(locale) {
        locale = (Array.isArray(locale) ? locale[0] ?? '' : locale).trim().toLowerCase();

        // BCP 47 base language fallback (e.g., 'zh-TW' → 'zh')
        if (locale && !supported.has(locale)) {
          const base = locale.split('-')[0];
          if (base !== locale && supported.has(base)) {
            locale = base;
          }
        }

        translateTo = locale || null; // overwrite the auto-detected language with the user-provided one
        localizeDocument(translateTo); // force re-localization of the entire document
      },
      translate: (text) => translateText(text, translateTo),
      localizeDocument,
    }),
    writable: false,
    enumerable: false,
    configurable: false,
  });

  document.readyState === 'loading'
    ? document.addEventListener('DOMContentLoaded', () => localizeDocument(translateTo))
    : localizeDocument(translateTo);
})();
