(() => { // TODO: refactor this file; it's currently difficult to read and maintain
  'use strict';

  const L10N_ATTR = 'data-l10n';
  const L10N_SELECTOR = '[' + L10N_ATTR + ']';

  const tokenizeRegex = /[^a-z0-9]+/g;
  const t = (s) => s.toLowerCase().replace(tokenizeRegex, '');
  const f = Object.freeze;

  /**
   * BCP 47 resolution - exact match first, then base-language fallback (e.g. 'zh-tw' tries 'zh-tw', then 'zh').
   *
   * @param {string} locale
   * @param {Object} map
   * @returns {string|undefined}
   */
  const resolve = (locale, map) => {
    if (Object.prototype.hasOwnProperty.call(map, locale)) {
      return map[locale];
    }

    const base = locale.split('-')[0];

    if (base !== locale && Object.prototype.hasOwnProperty.call(map, base)) {
      return map[base];
    }

    return undefined;
  };

  /** @type {Object<string, Object<string, string>>} */
  const translations = f({
    {{- range .Tokens }}
    [t({{.Key | quote}})]: f({
      {{- range .Translations}}
      {{ .LangCode | quote }}: {{ .Value | quote }},
      {{- end}}
    }),
    {{- end}}
  });

  /** @type {string[]} */
  let locales = (navigator.languages && navigator.languages.length
      ? Array.from(navigator.languages)
      : [navigator.language || 'en']
  ).map((l) => l.toLowerCase());

  /**
   * @param {string} token
   * @returns {string|undefined}
   */
  const translate = (token) => {
    const map = translations[t(token)];
    if (!map) {
      return undefined;
    }

    for (const locale of locales) {
      if (locale === 'en' || locale.startsWith('en-')) {
        return token;
      }

      const result = resolve(locale, map);
      if (result !== undefined) {
        return result;
      }
    }

    return token;
  };

  /**
   * On first call: promotes the implicit token (element's textContent) to an explicit value of the data-l10n
   * attribute. Subsequent setLocale() calls always read from the attribute - no extra DOM attributes needed.
   *
   * @param {Element} el
   */
  const localizeEl = (el) => {
    let token = el.getAttribute(L10N_ATTR).trim();

    if (!token) {
      token = el.textContent.trim();
      if (!token) {
        return;
      }

      el.setAttribute(L10N_ATTR, token); // promote once, read forever
    }

    const localized = translate(token);

    if (localized !== undefined) {
      el.textContent = localized;
    } else {
      console.debug('[l10n] Unknown token: "' + token + '" (locales: ' + locales.join(', ') + ')', el);
    }
  };

  /** @param {Document|Element} [root] */
  const localizeDocument = (root = document) => {
    root.querySelectorAll(L10N_SELECTOR).forEach(localizeEl);

    if (locales.length > 0 && !locales[0].startsWith('en')) {
      document.documentElement.setAttribute('lang', locales[0]);
    }
  };

  // observe the document for new elements with the data-l10n attribute and localize them
  new MutationObserver((mutations) => {
    for (const {addedNodes} of mutations) {
      for (const node of addedNodes) {
        if (node.nodeType !== 1) {
          continue;
        }

        if (node.hasAttribute(L10N_ATTR)) {
          localizeEl(node);
        }

        node.querySelectorAll(L10N_SELECTOR).forEach(localizeEl);
      }
    }
  }).observe(document.documentElement, {childList: true, subtree: true});

  Object.defineProperty(window, 'l10n', {
    value: Object.freeze({
      setLocale(locale) {
        locales = (Array.isArray(locale) ? locale : [locale]).map((l) => l.toLowerCase());
        localizeDocument();
      },
      translate,
      localizeDocument,
    }),
    writable: false,
    enumerable: false,
    configurable: false,
  });

  document.readyState === 'loading'
    ? document.addEventListener('DOMContentLoaded', () => localizeDocument())
    : localizeDocument();
})();
