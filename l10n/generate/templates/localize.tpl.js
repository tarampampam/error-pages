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
  const locales = Object.freeze(new Map([
    {{- range .Tokens }}
    [t({{.Key | quote}}), new Map([
      {{- range .Translations}}
      [{{ .LangCode | quote }}, {{ .Value | quote }}],
      {{- end}}
    ])],
    {{- end}}
  ]));

  /**
   * Set of supported language codes for quick lookup (each code is always in lowercase).
   *
   * @type {Set<string>}
   */
  const supported = Object.freeze(new Set([
    {{- range $i, $code := .SupportedLangCodes -}}
      {{- if $i }}, {{ end }}{{ $code | quote }}
    {{- end -}}
  ]));

  /**
   * Resolves a BCP 47 language tag to a supported language code, falling back to the base language if necessary.
   * Returns null if no match is found.
   *
   * @example
   * ```js
   * const supported = new Set(['en', 'fr', 'zh', 'zh-TW']);
   *
   * resolveLang('fr'); // returns 'fr'
   * resolveLang('fr-CA'); // returns 'fr'
   * resolveLang('zh-TW'); // returns 'zh-TW'
   * resolveLang('zh-HK'); // returns 'zh'
   * resolveLang('es'); // returns null
   * ```
   *
   * @param {string} lang - BCP 47 language tag
   * @returns {string|null}
   */
  const resolveLang = (lang) => {
    if (!lang || typeof lang !== 'string') {
      return null; // invalid input
    }

    lang = lang.trim().toLowerCase();

    if (supported.has(lang)) {
      return lang; // exact match
    }

    const baseParts = lang.split('-');
    if (baseParts.length < 2) {
      return null; // no base language to fall back to
    }

    const base = baseParts[0]; // base language (e.g., 'zh' from 'zh-TW')
    if (supported.has(base)) {
      return base; // base language match
    }

    return null; // no match found
  };

  /**
   * Translates the given raw text to the provided language. Returns null if no translation is available.
   *
   * @param {string} text - the raw text to translate (usually in English)
   * @param {string|null} language - the language code to translate to (will be normalized)
   * @returns {string|null}
   */
  const translateText = (text, language) => {
    if (!language) {
      return null; // no language to translate to
    }

    const lang = resolveLang(language);
    if (!lang) {
      return null; // unsupported language
    }

    // get all translations for the given text (tokenized)
    const translations = locales.get(t(text));
    if (!translations || !translations.size) {
      return null; // no translations for this text
    }

    const translated = translations.get(lang);
    if (translated) {
      return translated; // we found it!
    }

    return null; // no translation available for this language
  }

  // attribute name and selector for elements that can be localized
  const L10N_ATTR = 'data-l10n';
  const L10N_SELECTOR = '[' + L10N_ATTR + ']';

  /**
   * Localizes the given element by translating its text content to the provided language, if possible. The original
   * text is stored in the data-l10n attribute (if not already stored) to allow re-localization when the language
   * changes.
   *
   * In case the element has no translation for the given language, it will be restored to its original text and the
   * data-l10n attribute will be cleared.
   *
   * @param {Element} el
   * @param {string|null} language
   * @returns {boolean}
   */
  const localizeElement = (el, language) => {
    if (!el || el.nodeType !== 1) {
      return false; // skip non-element nodes
    }

    if (!el.hasAttribute(L10N_ATTR)) {
      return false; // node has no data-l10n attribute
    }

    const fromAttribute = el.getAttribute(L10N_ATTR) || null; // '' means "not yet promoted" - treat same as absent

    // on first localization the original text lives in textContent; afterward it is promoted to the data-l10n
    // attribute so subsequent calls can always read the original from there, regardless of the current textContent
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

      return true;
    } else {
      if (fromAttribute) {
        el.textContent = fromAttribute; // restore original when no translation available
        el.setAttribute(L10N_ATTR, ''); // remove the attribute value after restoring the original text
      }

      if (language) {
        // leave breadcrumbs for debugging purposes
        console.debug('[l10n] Unable to localize element', el, 'to language', language);
      }
    }

    return false;
  }

  /**
   * The state of the localization system.
   *
   * @type { {translateTo: string|null} }
   */
  const state = {
    // determine the initial language to translate to based on the user's browser settings and the supported languages
    translateTo: (() => {
      for (const lang of (navigator.languages || []).map((l) => l.toLowerCase())) {
        const resolved = resolveLang(lang);
        if (resolved) {
          return resolved;
        }
      }

      return null;
    })()
  }

  /**
   * Localizes the entire document by localizing all elements with the data-l10n attribute to the provided language.
   *
   * In case when provided language is null, unsupported, or the default one (English), previously localized elements
   * will be restored to their original text.
   *
   * @param {string|null} language
   */
  const localizeDocument = (language) => {
    document.querySelectorAll(L10N_SELECTOR).forEach((el) => localizeElement(el, language));
    document.documentElement.setAttribute('lang', language || 'en');
  }

  // start observing the document for new elements with the data-l10n attribute (required for localizing dynamically
  // added content)
  new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (node.nodeType !== 1) {
          continue; // skip non-element nodes
        }

        if (node.hasAttribute(L10N_ATTR)) {
          localizeElement(node, state.translateTo);
        }

        node.querySelectorAll(L10N_SELECTOR).forEach((el) => localizeElement(el, state.translateTo));
      }
    }
  }).observe(document.documentElement, {childList: true, subtree: true});

  Object.defineProperty(window, 'l10n', {
    value: Object.freeze({
      /** @param {string|string[]} locale */
      setLocale(locale) {
        if (!locale) {
          return; // no locale provided, do nothing
        }

        // supporting both string and array is required for backwards compatibility with the old API, but we will
        // only consider the first locale in the array if an array is provided
        const loc = (Array.isArray(locale) ? locale[0] ?? '' : locale).trim().toLowerCase();

        // resolve the language to a supported one
        const lang = resolveLang(loc);
        if (!lang) {
          return; // unsupported language, do nothing
        }

        state.translateTo = lang; // mutate the state with the new language to translate to
        localizeDocument(lang); // force re-localization of the entire document
      },
      /** @param {string} text */
      translate: (text) => translateText(text, state.translateTo),
      /**
       * Localizes the entire document to the provided language.
       *
       * If the provided language is null, unsupported, or the default one (English), all elements will be restored
       * to their original text.
       *
       * @param {string|null} language
       */
      localizeDocument: (language) => {
        const newLang = resolveLang(language);
        if (newLang) {
          state.translateTo = newLang; // mutate the state with the new language to translate to
        }

        localizeDocument(newLang || state.translateTo) // force re-localization of the entire document
      },
    }),
    writable: false,
    enumerable: false,
    configurable: false,
  });

  document.readyState === 'loading'
    ? document.addEventListener('DOMContentLoaded', () => localizeDocument(state.translateTo))
    : localizeDocument(state.translateTo);
})();
