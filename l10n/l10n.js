Object.defineProperty(window, 'l10n', {
    value: new function () {
        // language codes list: <https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes>
        const data = { // all keys should be in english (it is default/main locale)
            'Error': {
                fr: 'Erreur',
                ru: 'Ошибка',
                uk: 'Помилка',
            },
            'Good luck': {
                fr: 'Bonne chance',
                ru: 'Удачи',
                uk: 'Успіхів',
            },
            'UH OH': {
                fr: 'Oups',
                ru: 'Ох',
                uk: 'Ох',
            },
            'Request details': {
                fr: 'Détails de la demande',
                ru: 'Детали запроса',
                uk: 'Деталі запиту',
            },
            'Double-check the URL': {
                fr: 'Vérifiez l’URL',
                ru: 'Дважды проверьте URL',
                uk: 'Двічі перевіряйте URL-адресу',
            },
            'Alternatively, go back': {
                fr: 'Essayer de revenir en arrière',
                ru: 'Или можете вернуться назад',
                uk: 'Або ви можете повернутися',
            },
            'Here\'s what might have happened': {
                fr: 'Voici ce qui aurait pu se passer',
                ru: 'Из-за чего это могло случиться',
                uk: 'Що це може статися',
            },
            'You may have mistyped the URL': {
                fr: 'Vous avez peut-être mal tapé l’URL',
                ru: 'Вы могли ошибиться в URL',
                uk: 'Ви можете зробити помилку в URL-адресі',
            },
            'The site was moved': {
                fr: 'Le site a été déplacé',
                ru: 'Сайт был перемещён',
                uk: 'Сайт був переміщений',
            },
            'It was never here': {
                fr: 'Il n’a jamais été ici',
                ru: 'Он никогда не был здесь',
                uk: 'Він ніколи не був тут',
            },

            'Bad Request': {
                fr: 'Mauvaise demande',
                ru: 'Некорректный запрос',
                uk: 'Неправильний запит',
            },
            'The server did not understand the request': {
                fr: 'Le serveur ne comprend pas la demande',
                ru: 'Сервер не смог обработать запрос из-за ошибки в нём',
                uk: 'Сервер не міг обробити запит через помилку в ньому',
            },
            'Unauthorized': {
                fr: 'Non autorisé',
                ru: 'Запрос не авторизован',
                uk: 'Несанкціонований доступ',
            },
            'The requested page needs a username and a password': {
                fr: 'La page demandée nécessite un nom d’utilisateur et un mot de passe',
                ru: 'Для доступа к странице требуется логин и пароль',
                uk: 'Щоб отримати доступ до сторінки, потрібний логін та пароль',
            },
            'Forbidden': {
                fr: 'Interdit',
                ru: 'Запрещено',
                uk: 'Заборонено',
            },
            'Access is forbidden to the requested page': {
                fr: 'Accès interdit à la page demandée',
                ru: 'Доступ к странице запрещён',
                uk: 'Доступ до сторінки заборонено',
            },
            'Not Found': {
                fr: 'Pas trouvé',
                ru: 'Страница не найдена',
                uk: 'Сторінка не знайдена',
            },
            'The server can not find the requested page': {
                fr: 'Le serveur ne peut trouver la page demandée',
                ru: 'Сервер не смог найти запрашиваемую страницу',
                uk: 'Сервер не міг знайти запитану сторінку',
            },
            'Method Not Allowed': {
                fr: 'Méthode Non Autorisée',
                ru: 'Метод не поддерживается',
                uk: 'Неприпустимий метод',
            },
            'The method specified in the request is not allowed': {
                fr: 'La méthode spécifiée dans la requête n’est pas autorisée',
                ru: 'Указанный в запросе метод не поддерживается',
                uk: 'Метод, зазначений у запиті, не підтримується',
            },
            'Proxy Authentication Required': {
                fr: 'Authentification proxy requise',
                ru: 'Нужна аутентификация прокси',
                uk: 'Потрібна ідентифікація проксі',
            },
            'You must authenticate with a proxy server before this request can be served': {
                fr: 'Vous devez vous authentifier avec un serveur proxy avant que cette requête puisse être servie',
                ru: 'Вы должны быть авторизованы на прокси сервере для обработки этого запроса',
                uk: 'Ви повинні увійти до проксі-сервера для обробки цього запиту',
            },
            'Request Timeout': {
                fr: 'Demande expiré',
                ru: 'Истекло время ожидания',
                uk: 'Час запиту закінчився',
            },
            'The request took longer than the server was prepared to wait': {
                fr: 'La requête prend plus de temps que prévu',
                ru: 'Отправка запроса заняла слишком много времени',
                uk: 'Надсилання запиту зайняв занадто багато часу',
            },
            'Conflict': {
                fr: 'Conflit',
                ru: 'Конфликт',
                uk: 'Конфлікт',
            },
            'The request could not be completed because of a conflict': {
                fr: 'La requête n’a pas pu être complétée à cause d’un conflit',
                ru: 'Запрос не может быть обработан из-за конфликта',
                uk: 'Запит не може бути оброблений через конфлікт',
            },
            'Gone': {
                fr: 'Supprimé',
                ru: 'Удалено',
                uk: 'Вилучений',
            },
            'The requested page is no longer available': {
                fr: 'La page demandée n’est plus disponible',
                ru: 'Запрошенная страница была удалена',
                uk: 'Запитана сторінка була видалена',
            },
            'Length Required': {
                fr: 'Longueur requise',
                ru: 'Необходима длина',
                uk: 'Потрібно вказати розмір',
            },
            'The "Content-Length" is not defined. The server will not accept the request without it': {
                fr: 'Le "Content-Length" n’est pas défini. Le serveur ne prendra pas en compte la requête sans',
                ru: 'Заголовок "Content-Length" не был передан. Сервер не может обработать запрос без него',
                uk: 'Заголовок "Content-Length" не був переданий. Сервер не може обробити запит без нього',
            },
            'Precondition Failed': {
                fr: 'Échec de la condition préalable',
                ru: 'Условие ложно',
                uk: 'Збій під час обробки попередньої умови',
            },
            'The pre condition given in the request evaluated to false by the server': {
                fr: 'La précondition donnée dans la requête a été évaluée comme étant fausse par le serveur',
                ru: 'Ни одно из условных полей заголовка запроса не было выполнено',
                uk: 'Жодна з умовних полів заголовка запиту не була виконана',
            },
            'Payload Too Large': {
                fr: 'Charge trop volumineuse',
                ru: 'Слишком большой запрос',
                uk: 'Занадто великий запит',
            },
            'The server will not accept the request, because the request entity is too large': {
                fr: 'Le serveur ne prendra pas en compte la requête, car l’entité de la requête est trop volumineuse',
                ru: 'Сервер не может обработать запрос, так как он слишком большой',
                uk: 'Сервер не може обробити запит, оскільки він занадто великий',
            },
            'Requested Range Not Satisfiable': {
                fr: 'Demande non satisfaisante',
                ru: 'Диапазон не достижим',
                uk: 'Запитуваний діапазон недосяжний',
            },
            'The requested byte range is not available and is out of bounds': {
                fr: 'Le byte range demandé n’est pas disponible et est hors des limites',
                ru: 'Запрошенный диапазон данных недоступен или вне допустимых пределов',
                uk: 'Описаний діапазон даних недоступний або з допустимих меж',
            },
            'I\'m a teapot': {
                fr: 'Je suis une théière',
                ru: 'Я чайник',
                uk: 'Я чайник',
            },
            'Attempt to brew coffee with a teapot is not supported': {
                fr: 'Tenter de préparer du café avec une théière n’est pas pris en charge',
                ru: 'Попытка заварить кофе в чайнике обречена на фиаско',
                uk: 'Спроба виварити каву в чайник приречена на фіаско',
            },
            'Too Many Requests': {
                fr: 'Trop de demandes',
                ru: 'Слишком много запросов',
                uk: 'Занадто багато запитів',
            },
            'Too many requests in a given amount of time': {
                fr: 'Trop de requêtes dans un délai donné',
                ru: 'Отправлено слишком много запросов за короткое время',
                uk: 'Надіслано занадто багато запитів на короткий час',
            },
            'Internal Server Error': {
                fr: 'Erreur interne du serveur',
                ru: 'Внутренняя ошибка сервера',
                uk: 'Внутрішня помилка сервера',
            },
            'The server met an unexpected condition': {
                fr: 'Le serveur a rencontré une condition inattendue',
                ru: 'Произошло что-то неожиданное на сервере',
                uk: 'На сервері було щось несподіване',
            },
            'Bad Gateway': {
                fr: 'Mauvaise passerelle',
                ru: 'Ошибка шлюза',
                uk: 'Помилка шлюзу',
            },
            'The server received an invalid response from the upstream server': {
                fr: 'Le serveur a reçu une réponse invalide du serveur distant',
                ru: 'Сервер получил некорректный ответ от вышестоящего сервера',
                uk: 'Сервер отримав неправильну відповідь з сервера Upstream',
            },
            'Service Unavailable': {
                fr: 'Service indisponible',
                ru: 'Сервис недоступен',
                uk: 'Сервіс недоступний',
            },
            'The server is temporarily overloading or down': {
                fr: 'Le serveur est temporairement en surcharge ou indisponible',
                ru: 'Сервер временно не может обрабатывать запросы по техническим причинам',
                uk: 'Сервер тимчасово не може обробляти запити з технічних причин',
            },
            'Gateway Timeout': {
                fr: 'Expiration Passerelle',
                ru: 'Шлюз не отвечает',
                uk: 'Шлюз не відповідає',
            },
            'The gateway has timed out': {
                fr: 'Le temps d’attente de la passerelle est dépassé',
                ru: 'Сервер не дождался ответа от вышестоящего сервера',
                uk: 'Сервер не чекав відповіді від сервера Upstream',
            },
            'HTTP Version Not Supported': {
                fr: 'Version HTTP non prise en charge',
                ru: 'Версия HTTP не поддерживается',
                uk: 'Версія НТТР не підтримується',
            },
            'The server does not support the "http protocol" version': {
                fr: 'Le serveur ne supporte pas la version du protocole HTTP"',
                ru: 'Сервер не поддерживает запрошенную версию HTTP протокола',
                uk: 'Сервер не підтримує запитану версію HTTP-протоколу',
            },

            'Host': {
                fr: 'Hôte',
                ru: 'Хост',
                uk: 'Хост',
            },
            'Original URI': {
                fr: 'URI d’origine',
                ru: 'Исходный URI',
                uk: 'Вихідний URI',
            },
            'Forwarded for': {
                fr: 'Transmis pour',
                ru: 'Перенаправлен',
                uk: 'Перенаправлений',
            },
            'Namespace': {
                fr: 'Espace de noms',
                ru: 'Пространство имён',
                uk: 'Простір імен',
            },
            'Ingress name': {
                fr: 'Nom ingress',
                ru: 'Имя Ingress',
                uk: 'Ім\'я Ingress',
            },
            'Service name': {
                fr: 'Nom du service',
                ru: 'Имя сервиса',
                uk: 'Ім\'я сервісу',
            },
            'Service port': {
                fr: 'Port du service',
                ru: 'Порт сервиса',
                uk: 'Порт сервісу',
            },
            'Request ID': {
                fr: 'Identifiant de la demande',
                ru: 'ID запроса',
                uk: 'ID запиту',
            },
            'Timestamp': {
                fr: 'Horodatage',
                ru: 'Временная метка',
                uk: 'Тимчасова мітка',
            },

            'client-side error': {
                fr: 'Erreur du client',
                ru: 'ошибка на стороне клиента',
                uk: 'помилка на стороні клієнта',
            },
            'server-side error': {
                fr: 'Erreur du serveur',
                ru: 'ошибка на стороне сервера',
                uk: 'помилка на стороні сервера',
            },

            'Your Client': {
                fr: 'Votre client',
                ru: 'Ваш Браузер',
                uk: 'Ваш Браузер',
            },
            'Network': {
                fr: 'Réseau',
                ru: 'Сеть',
                uk: 'Сіть',
            },
            'Web Server': {
                fr: 'Serveur Web',
                ru: 'Web Сервер',
                uk: 'Web Сервер',
            },
            'What happened?': {
                fr: 'Que s’est-il passé ?',
                ru: 'Что произошло?',
                uk: 'Що сталося?',
            },
            'What can i do?': {
                fr: 'Que puis-je faire ?',
                ru: 'Что можно сделать?',
                uk: 'Що можна зробити?',
            },
            'Please try again in a few minutes': {
                fr: 'Veuillez réessayer dans quelques minutes',
                ru: 'Пожалуйста, попробуйте повторить запрос ещё раз чуть позже',
                uk: 'Будь ласка, спробуйте повторити запит ще раз трохи пізніше',
            },
            'Working': {
                fr: 'Opérationnel',
                ru: 'Работает',
                uk: 'Працює',
            },
            'Unknown': {
                fr: 'Inconnu',
                ru: 'Неизвестно',
                uk: 'Невідомо',
            },
            'Please try to change the request method, headers, payload, or URL': {
                fr: 'Veuillez essayer de changer la méthode de requête, les en-têtes, le contenu ou l’URL',
                ru: 'Пожалуйста, попробуйте изменить метод запроса, заголовки, его содержимое или URL',
                uk: 'Будь ласка, спробуйте змінити метод запиту, заголовки, його вміст або URL-адресу',
            },
            'Please check your authorization data': {
                fr: 'Veuillez vérifier vos données d’autorisation',
                ru: 'Пожалуйста, проверьте данные авторизации',
                uk: 'Будь ласка, перевірте дані авторизації',
            },
            'Please double-check the URL and try again': {
                fr: 'Veuillez vérifier l’URL et réessayer',
                ru: 'Пожалуйста, дважды проверьте URL и попробуйте снова',
                uk: 'Будь ласка, двічі перевірте URL-адресу і спробуйте знову',
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

            if (activeLocale === 'en' && Object.prototype.hasOwnProperty.call(data, t)) {
                return token
            }

            if (Object.prototype.hasOwnProperty.call(data, t) && Object.prototype.hasOwnProperty.call(data[t], activeLocale)) {
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
