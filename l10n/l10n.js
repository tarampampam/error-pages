const l10n = new function () {
    const defaultLocale = 'en'; // default locale

    this.data = { // all keys should be in english by default
        'Bad Request': {ru: 'Некорректный Запрос'},
        'Unauthorized': {ru: 'Не Авторизован'},
        'Forbidden': {ru: 'Запрещено'},
        'Not Found': {ru: 'Не Найдено'},
        'Method Not Allowed': {ru: 'Метод Не Поддерживается'},
        'Proxy Authentication Required': {ru: 'Необходима Аутентификация Прокси'},
        'Request Timeout': {ru: 'Истекло Время Ожидания'},
        'Conflict': {ru: 'Конфликт'},
        'Gone': {ru: 'Удалён'},
        'Length Required': {ru: 'Необходима Длина'},
        'Precondition Failed': {ru: 'Условие Ложно'},
        'Payload Too Large': {ru: 'Полезная Нагрузка Слишком Велика'},
        'Requested Range Not Satisfiable': {ru: 'Диапазон Не Достижим'},
        'I\'m a teapot': {ru: 'Я Чайник'},
        'Too Many Requests': {ru: 'Слишком Много Запросов'},
        'Internal Server Error': {ru: 'Внутренняя Ошибка Сервера'},
        'Bad Gateway': {ru: 'Ошибка Шлюза'},
        'Service Unavailable': {ru: 'Сервис Недоступен'},
        'Gateway Timeout': {ru: 'Шлюз Не Отвечает'},
        'HTTP Version Not Supported': {ru: 'Версия HTTP Не Поддерживается'},

        'Host': {ru: 'Хост'},
        'Original URI': {ru: 'Исходный URI'},
        'Forwarded for': {ru: 'Перенаправлен'},
        'Namespace': {ru: 'Пространство имён'},
        'Ingress name': {ru: 'Имя Ingress'},
        'Service name': {ru: 'Имя сервиса'},
        'Service port': {ru: 'Порт сервиса'},
        'Request ID': {ru: 'ID запроса'},
        'Timestamp': {ru: 'Временная метка'},

        'client-side error': {ru: 'ошибка на стороне клиента'},
        'server-side error': {ru: 'ошибка на стороне сервера'},

        'Your Client': {ru: 'Ваш браузер'},
        'Network': {ru: 'Сеть'},
        'Web Server': {ru: 'Web сервер'},
        'What happened?': {ru: 'Что произошло?'},
        'What can i do?': {ru: 'Что можно сделать?'},
        'Please try again in a few minutes': {ru: 'Пожалуйста, попробуйте ещё раз чуть позже'},
        'Working': {ru: 'Работает'},
        'Unknown': {ru: 'Неизвестно'},
        'Please try to change the request method, headers, payload, or URL': {ru: 'Пожалуйста, попробуйте изменить метод запроса, заголовки, его содержимое или URL'},
        'Please check your authorization data': {ru: 'Пожалуйста, проверьте данные авторизации'},
        'Please double-check the URL and try again': {ru: 'Пожалуйста, дважды проверьте URL и попробуйте снова'},
    };

    let activeLocale = defaultLocale;

    // detect browser locale (take only 2 first symbols)
    const match = /[a-z]{2}/s.exec((window.navigator.languages[0] || window.navigator.language).trim().toLowerCase());
    if (typeof match[0] === 'string') {
        activeLocale = match[0].toLowerCase();
    }

    /**
     * @param {string} locale
     */
    this.setLocale = function (locale) {
        activeLocale = locale
    }

    /**
     * @param {string} token
     * @return {string}
     */
    const normalizeToken = function (token) {
        return token.trim().toLowerCase();
    };

    /**
     * @param {string} token
     * @param {string|undefined?} def
     */
    this.translate = function (token, def) {
        if (activeLocale === defaultLocale && this.data.hasOwnProperty(token)) {
            return token;
        }

        if (this.data.hasOwnProperty(token) && this.data[token].hasOwnProperty(activeLocale)) {
            return this.data[token][activeLocale];
        }

        const cacheKey = '__ck';
        for (const key in this.data) { // slowest way (fallback)
            if (!this.data[key].hasOwnProperty(cacheKey)) { // boost the cache
                this.data[key][cacheKey] = normalizeToken(key);
            }

            if (this.data[key][cacheKey] === normalizeToken(token) && this.data[key].hasOwnProperty(activeLocale)) {
                return this.data[key][activeLocale];
            }
        }

        return def;
    };

    /**
     * Localize all elements with HTML attributes `data-l10n`.
     */
    this.localize = function () {
        Array.prototype.forEach.call(document.querySelectorAll('[data-l10n]'), ($el) => {
            const localized = this.translate($el.getAttribute('data-l10n') || $el.innerText.trim(), undefined);

            if (localized !== undefined) {
                $el.innerText = localized;
            }
        });
    };
};
