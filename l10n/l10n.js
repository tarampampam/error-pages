Object.defineProperty(window, 'l10n', {
    value: new function () {
        // language codes list: <https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes>
        const data = { // all keys should be in english (it is default/main locale)
            'Error': {ru: 'Ошибка', uk: 'Помилка'},
            'Good luck': {ru: 'Удачи', uk: 'Успіхів'},
            'UH OH': {ru: 'Ох', uk: 'Ох'},
            'Request details': {ru: 'Детали запроса', uk: 'Деталі запиту'},
            'Double-check the URL': {ru: 'Дважды проверьте URL', uk: 'Двічі перевіряйте URL-адресу'},
            'Alternatively, go back': {ru: 'Или можете вернуться назад', uk: 'Або ви можете повернутися'},
            'Here\'s what might have happened': {ru: 'Из-за чего это могло случиться', uk: 'Що це може статися'},
            'You may have mistyped the URL': {
                ru: 'Вы могли ошибиться в URL',
                uk: 'Ви можете зробити помилку в URL-адресі'
            },
            'The site was moved': {ru: 'Сайт был перемещён', uk: 'Сайт був переміщений'},
            'It was never here': {ru: 'Он никогда не был здесь', uk: 'Він ніколи не був тут'},

            'Bad Request': {ru: 'Некорректный запрос', uk: 'Неправильний запит'},
            'The server did not understand the request': {
                ru: 'Сервер не смог обработать запрос из-за ошибки в нём',
                uk: 'Сервер не міг обробити запит через помилку в ньому'
            },
            'Unauthorized': {ru: 'Запрос не авторизован', uk: 'Несанкціонований доступ'},
            'The requested page needs a username and a password': {
                ru: 'Для доступа к странице требуется логин и пароль',
                uk: 'Щоб отримати доступ до сторінки, потрібний логін та пароль'
            },
            'Forbidden': {ru: 'Запрещено', uk: 'Заборонено'},
            'Access is forbidden to the requested page': {
                ru: 'Доступ к странице запрещён',
                uk: 'Доступ до сторінки заборонено'
            },
            'Not Found': {ru: 'Страница не найдена', uk: 'Сторінка не знайдена'},
            'The server can not find the requested page': {
                ru: 'Сервер не смог найти запрашиваемую страницу',
                uk: 'Сервер не міг знайти запитану сторінку'
            },
            'Method Not Allowed': {ru: 'Метод не поддерживается', uk: 'Неприпустимий метод'},
            'The method specified in the request is not allowed': {
                ru: 'Указанный в запросе метод не поддерживается',
                uk: 'Метод, зазначений у запиті, не підтримується'
            },
            'Proxy Authentication Required': {ru: 'Нужна аутентификация прокси', uk: 'Потрібна ідентифікація проксі'},
            'You must authenticate with a proxy server before this request can be served': {
                ru: 'Вы должны быть авторизованы на прокси сервере для обработки этого запроса',
                uk: 'Ви повинні увійти до проксі-сервера для обробки цього запиту'
            },
            'Request Timeout': {ru: 'Истекло время ожидания', uk: 'Час запиту закінчився'},
            'The request took longer than the server was prepared to wait': {
                ru: 'Отправка запроса заняла слишком много времени',
                uk: 'Надсилання запиту зайняв занадто багато часу'
            },
            'Conflict': {ru: 'Конфликт', uk: 'Конфлікт'},
            'The request could not be completed because of a conflict': {
                ru: 'Запрос не может быть обработан из-за конфликта',
                uk: 'Запит не може бути оброблений через конфлікт'
            },
            'Gone': {ru: 'Удалено', uk: 'Вилучений'},
            'The requested page is no longer available': {
                ru: 'Запрошенная страница была удалена',
                uk: 'Запитана сторінка була видалена'
            },
            'Length Required': {ru: 'Необходима длина', uk: 'Потрібно вказати розмір'},
            'The "Content-Length" is not defined. The server will not accept the request without it': {
                ru: 'Заголовок "Content-Length" не был передан. Сервер не может обработать запрос без него',
                uk: 'Заголовок "Content-Length" не був переданий. Сервер не може обробити запит без нього'
            },
            'Precondition Failed': {ru: 'Условие ложно', uk: 'Збій під час обробки попередньої умови'},
            'The pre condition given in the request evaluated to false by the server': {
                ru: 'Ни одно из условных полей заголовка запроса не было выполнено',
                uk: 'Жодна з умовних полів заголовка запиту не була виконана'
            },
            'Payload Too Large': {ru: 'Слишком большой запрос', uk: 'Занадто великий запит'},
            'The server will not accept the request, because the request entity is too large': {
                ru: 'Сервер не может обработать запрос, так как он слишком большой',
                uk: 'Сервер не може обробити запит, оскільки він занадто великий'
            },
            'Requested Range Not Satisfiable': {ru: 'Диапазон не достижим', uk: 'Запитуваний діапазон недосяжний'},
            'The requested byte range is not available and is out of bounds': {
                ru: 'Запрошенный диапазон данных недоступен или вне допустимых пределов',
                uk: 'Описаний діапазон даних недоступний або з допустимих меж'
            },
            'I\'m a teapot': {ru: 'Я чайник', uk: 'Я чайник'},
            'Attempt to brew coffee with a teapot is not supported': {
                ru: 'Попытка заварить кофе в чайнике обречена на фиаско',
                uk: 'Спроба виварити каву в чайник приречена на фіаско'
            },
            'Too Many Requests': {ru: 'Слишком много запросов', uk: 'Занадто багато запитів'},
            'Too many requests in a given amount of time': {
                ru: 'Отправлено слишком много запросов за короткое время',
                uk: 'Надіслано занадто багато запитів на короткий час'
            },
            'Internal Server Error': {ru: 'Внутренняя ошибка сервера', uk: 'Внутрішня помилка сервера'},
            'The server met an unexpected condition': {
                ru: 'Произошло что-то неожиданное на сервере',
                uk: 'На сервері було щось несподіване'
            },
            'Bad Gateway': {ru: 'Ошибка шлюза', uk: 'Помилка шлюзу'},
            'The server received an invalid response from the upstream server': {
                ru: 'Сервер получил некорректный ответ от вышестоящего сервера',
                uk: 'Сервер отримав неправильну відповідь з сервера Upstream'
            },
            'Service Unavailable': {ru: 'Сервис недоступен', uk: 'Сервіс недоступний'},
            'The server is temporarily overloading or down': {
                ru: 'Сервер временно не может обрабатывать запросы по техническим причинам',
                uk: 'Сервер тимчасово не може обробляти запити з технічних причин'
            },
            'Gateway Timeout': {ru: 'Шлюз не отвечает', uk: 'Шлюз не відповідає'},
            'The gateway has timed out': {
                ru: 'Сервер не дождался ответа от вышестоящего сервера',
                uk: 'Сервер не чекав відповіді від сервера Upstream'
            },
            'HTTP Version Not Supported': {ru: 'Версия HTTP не поддерживается', uk: 'Версія НТТР не підтримується'},
            'The server does not support the "http protocol" version': {
                ru: 'Сервер не поддерживает запрошенную версию HTTP протокола',
                uk: 'Сервер не підтримує запитану версію HTTP-протоколу'
            },

            'Host': {ru: 'Хост', uk: 'Хост'},
            'Original URI': {ru: 'Исходный URI', uk: 'Вихідний URI'},
            'Forwarded for': {ru: 'Перенаправлен', uk: 'Перенаправлений'},
            'Namespace': {ru: 'Пространство имён', uk: 'Простір імен'},
            'Ingress name': {ru: 'Имя Ingress', uk: 'Ім\'я Ingress'},
            'Service name': {ru: 'Имя сервиса', uk: 'Ім\'я сервісу'},
            'Service port': {ru: 'Порт сервиса', uk: 'Порт сервісу'},
            'Request ID': {ru: 'ID запроса', uk: 'ID запиту'},
            'Timestamp': {ru: 'Временная метка', uk: 'Тимчасова мітка'},

            'client-side error': {ru: 'ошибка на стороне клиента', uk: 'помилка на стороні клієнта'},
            'server-side error': {ru: 'ошибка на стороне сервера', uk: 'помилка на стороні сервера'},

            'Your Client': {ru: 'Ваш Браузер', uk: 'Ваш Браузер'},
            'Network': {ru: 'Сеть', uk: 'Сіть'},
            'Web Server': {ru: 'Web Сервер', uk: 'Web Сервер'},
            'What happened?': {ru: 'Что произошло?', uk: 'Що сталося?'},
            'What can i do?': {ru: 'Что можно сделать?', uk: 'Що можна зробити?'},
            'Please try again in a few minutes': {
                ru: 'Пожалуйста, попробуйте повторить запрос ещё раз чуть позже',
                uk: 'Будь ласка, спробуйте повторити запит ще раз трохи пізніше'
            },
            'Working': {ru: 'Работает', uk: 'Працює'},
            'Unknown': {ru: 'Неизвестно', uk: 'Невідомо'},
            'Please try to change the request method, headers, payload, or URL': {
                ru: 'Пожалуйста, попробуйте изменить метод запроса, заголовки, его содержимое или URL',
                uk: 'Будь ласка, спробуйте змінити метод запиту, заголовки, його вміст або URL-адресу'
            },
            'Please check your authorization data': {
                ru: 'Пожалуйста, проверьте данные авторизации',
                uk: 'Будь ласка, перевірте дані авторизації'
            },
            'Please double-check the URL and try again': {
                ru: 'Пожалуйста, дважды проверьте URL и попробуйте снова',
                uk: 'Будь ласка, двічі перевірте URL-адресу і спробуйте знову'
            },
        };

        /**
         * @param {string} token
         * @return {string}
         */
        const serializeToken = function (token) {
            return token.toLowerCase().replaceAll(/[^a-z0-9]/g, '');
        };

        // normalize the data keys
        for (const key in data) {
            Object.defineProperty(data, serializeToken(key), Object.getOwnPropertyDescriptor(data, key));
            delete data[key];
        }

        // detect browser locale (take only 2 first symbols)
        let activeLocale = navigator.language.substring(0, 2).toLowerCase();

        /**
         * @param {string} locale
         */
        this.setLocale = function (locale) {
            activeLocale = locale.toLowerCase();
        }

        /**
         * @param {string} token
         * @param {string|undefined?} def
         */
        this.translate = function (token, def) {
            const t = serializeToken(token);

            if (activeLocale === 'en' && data.hasOwnProperty(t)) {
                return token
            }

            if (data.hasOwnProperty(t) && data[t].hasOwnProperty(activeLocale)) {
                return data[t][activeLocale];
            }

            return def;
        };

        /**
         * Localize all elements with HTML attribute `data-l10n`.
         */
        this.localizeDocument = function () {
            const dataAttributeName = 'data-l10n';

            Array.prototype.forEach.call(document.querySelectorAll('[' + dataAttributeName + ']'), ($el) => {
                const attr = $el.getAttribute(dataAttributeName).trim(),
                    token = attr.length > 0 ? attr : $el.innerText.trim(),
                    localized = this.translate(token, undefined);

                if (attr.length === 0) {
                    $el.setAttribute(dataAttributeName, token);
                }

                if (localized !== undefined) {
                    $el.innerText = localized;
                } else {
                    console.debug(`Unsupported l10n token detected: "${token}" (locale "${activeLocale}")`, $el);
                }
            });
        };
    },
    writable: false,
    enumerable: false,
});

window.l10n.localizeDocument();
