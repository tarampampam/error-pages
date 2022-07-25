Object.defineProperty(window, 'l10n', {
    value: new function () {
        // language codes list: <https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes>
        const data = { // all keys should be in english (it is default/main locale)
            'Error': {
                fr: 'Erreur',
                ru: 'Ошибка',
                uk: 'Помилка',
                pt: 'Erro',
                nl: 'Fout',
                de: 'Fehler',
                es: 'Error',
            },
            'Good luck': {
                fr: 'Bonne chance',
                ru: 'Удачи',
                uk: 'Успіхів',
                pt: 'Boa sorte',
                nl: 'Veel succes',
                de: 'Viel Glück',
                es: 'Buena Suerte',
            },
            'UH OH': {
                fr: 'Oups',
                ru: 'Ох',
                uk: 'Ох',
                pt: 'Ops',
                nl: 'Oeps',
                de: 'Hoppla',
                es: 'Oups',
            },
            'Request details': {
                fr: 'Détails de la requête',
                ru: 'Детали запроса',
                uk: 'Деталі запиту',
                pt: 'Detalhes da solicitação',
                nl: 'Details van verzoek',
                de: 'Details der Anfrage',
                es: 'Detalles de la petición',
            },
            'Double-check the URL': {
                fr: 'Vérifiez l’URL',
                ru: 'Дважды проверьте URL',
                uk: 'Двічі перевіряйте URL-адресу',
                pt: 'Verifique novamente a URL',
                nl: 'Controleer de URL',
                de: 'Überprüfen Sie die URL',
                es: 'Verifique la url',
            },
            'Alternatively, go back': {
                fr: 'Essayer de revenir en arrière',
                ru: 'Или можете вернуться назад',
                uk: 'Або ви можете повернутися',
                pt: "Como alternativa, tente voltar",
                nl: 'Of ga terug',
                de: 'Alternativ gehen Sie zurück',
                es: 'Como alternativa, vuelva atrás',
            },
            'Here\'s what might have happened': {
                fr: 'Voici ce qui aurait pu se passer',
                ru: 'Из-за чего это могло случиться',
                uk: 'Що це може статися',
                pt: 'Aqui está o que pode ter acontecido',
                nl: 'Wat er gebeurd kan zijn',
                de: 'Folgendes könnte passiert sein',
                es: 'Esto es lo que ha podido pasar',
            },
            'You may have mistyped the URL': {
                fr: 'Vous avez peut-être mal tapé l’URL',
                ru: 'Вы могли ошибиться в URL',
                uk: 'Ви можете зробити помилку в URL-адресі',
                pt: 'Você pode ter digitado incorretamente a URL',
                nl: 'De URL bevat een typefout',
                de: 'Möglicherweise haben Sie die URL falsch eingegeben',
                es: 'Igual ha escrito mal la URL',
            },
            'The site was moved': {
                fr: 'Le site a été déplacé',
                ru: 'Сайт был перемещён',
                uk: 'Сайт був переміщений',
                pt: 'O site foi movido',
                nl: 'De site is verplaatst',
                de: 'Die Seite wurde verschoben',
                es: 'El sitio se ha movido',
            },
            'It was never here': {
                fr: 'Il n’a jamais été ici',
                ru: 'Он никогда не был здесь',
                uk: 'Він ніколи не був тут',
                pt: 'Nunca esteve aqui',
                nl: 'Het was hier nooit',
                de: 'Es war nie hier',
                es: 'Nunca ha estado aquí',
            },
            'Bad Request': {
                fr: 'Mauvaise requête',
                ru: 'Некорректный запрос',
                uk: 'Неправильний запит',
                pt: 'Requisição inválida',
                nl: 'Foutieve anvraag',
                de: 'Fehlerhafte Anfrage',
                es: 'Petición inválida'
            },
            'The server did not understand the request': {
                fr: 'Le serveur ne comprend pas la requête',
                ru: 'Сервер не смог обработать запрос из-за ошибки в нём',
                uk: 'Сервер не міг обробити запит через помилку в ньому',
                pt: 'O servidor não entendeu a solicitação',
                nl: 'De server begreep het verzoek niet',
                de: 'Der Server hat die Anfrage nicht verstanden',
                es: 'El servidor no entendió la petición'
            },
            'Unauthorized': {
                fr: 'Non autorisé',
                ru: 'Запрос не авторизован',
                uk: 'Несанкціонований доступ',
                pt: 'Não autorizado',
                nl: 'Niet geautoriseerd',
                de: 'Nicht autorisiert',
                es: 'No autorizado',
            },
            'The requested page needs a username and a password': {
                fr: 'La page demandée nécessite un nom d’utilisateur et un mot de passe',
                ru: 'Для доступа к странице требуется логин и пароль',
                uk: 'Щоб отримати доступ до сторінки, потрібний логін та пароль',
                pt: 'A página solicitada precisa de um nome de usuário e uma senha',
                nl: 'De pagina heeft een gebruikersnaam en wachtwoord nodig',
                de: 'Die angeforderte Seite benötigt einen Benutzernamen und ein Passwort',
                es: 'La página solicitada necesita un usuario y una contraseña',
            },
            'Forbidden': {
                fr: 'Interdit',
                ru: 'Запрещено',
                uk: 'Заборонено',
                pt: 'Proibido',
                nl: 'Verboden',
                de: 'Verboten',
                es: 'Prohibido',
            },
            'Access is forbidden to the requested page': {
                fr: 'Accès interdit à la page demandée',
                ru: 'Доступ к странице запрещён',
                uk: 'Доступ до сторінки заборонено',
                pt: 'É proibido o acesso à página solicitada',
                nl: 'Toegang tot de pagina is verboden',
                de: 'Der Zugriff auf die angeforderte Seite ist verboten',
                es: 'El acceso está prohibido para la página solicitada',
            },
            'Not Found': {
                fr: 'Introuvable',
                ru: 'Страница не найдена',
                uk: 'Сторінка не знайдена',
                pt: 'Não encontrado',
                nl: 'Niet gevonden',
                de: 'Nicht gefunden',
                es: 'No encontrado',
            },
            'The server can not find the requested page': {
                fr: 'Le serveur ne peut trouver la page demandée',
                ru: 'Сервер не смог найти запрашиваемую страницу',
                uk: 'Сервер не міг знайти запитану сторінку',
                pt: 'O servidor não consegue encontrar a página solicitada',
                nl: 'De server kan de pagina niet vinden',
                de: 'Der Server kann die angeforderte Seite nicht finden',
                es: 'El servidor no puede encontrar la página solicitada',
            },
            'Method Not Allowed': {
                fr: 'Méthode Non Autorisée',
                ru: 'Метод не поддерживается',
                uk: 'Неприпустимий метод',
                pt: 'Método não permitido',
                nl: 'Methode niet toegestaan',
                de: 'Methode nicht erlaubt',
                es: 'Método no permitido',
            },
            'The method specified in the request is not allowed': {
                fr: 'La méthode spécifiée dans la requête n’est pas autorisée',
                ru: 'Указанный в запросе метод не поддерживается',
                uk: 'Метод, зазначений у запиті, не підтримується',
                pt: 'O método especificado na solicitação não é permitido',
                nl: 'De methode in het verzoek is niet toegestaan',
                de: 'Die in der Anfrage angegebene Methode ist nicht zulässig',
                es: 'El método especificado en la petición no está permitido'
            },
            'Proxy Authentication Required': {
                fr: 'Authentification proxy requise',
                ru: 'Нужна аутентификация прокси',
                uk: 'Потрібна ідентифікація проксі',
                pt: 'Autenticação de proxy necessária',
                nl: 'Authenticatie op de proxyserver verplicht',
                de: 'Proxy-Authentifizierung benötigt',
                es: 'Autenticación de proxy requerida',
            },
            'You must authenticate with a proxy server before this request can be served': {
                fr: 'Vous devez vous authentifier avec un serveur proxy avant que cette requête puisse être servie',
                ru: 'Вы должны быть авторизованы на прокси сервере для обработки этого запроса',
                uk: 'Ви повинні увійти до проксі-сервера для обробки цього запиту',
                pt: 'Você deve se autenticar com um servidor proxy antes que esta solicitação possa ser atendida',
                nl: 'Je moet authenticeren bij een proxyserver voordat dit verzoek uitgevoerd kan worden',
                de: 'Sie müssen sich bei einem Proxy-Server authentifizieren, bevor diese Anfrage bedient werden kann',
                es: 'Debes autentificarte con un servidor proxy antes de que esta petición pueda ser atendida',
            },
            'Request Timeout': {
                fr: 'Requête expiré',
                ru: 'Истекло время ожидания',
                uk: 'Час запиту закінчився',
                pt: 'Tempo limite de solicitação excedido',
                nl: 'Aanvraagtijd verstreken',
                de: 'Zeitüberschreitung der Anforderung',
                es: 'Tiempo límite de la petición excedido',
            },
            'The request took longer than the server was prepared to wait': {
                fr: 'La requête prend plus de temps que prévu',
                ru: 'Отправка запроса заняла слишком много времени',
                uk: 'Надсилання запиту зайняв занадто багато часу',
                pt: 'A solicitação demorou mais do que o servidor estava preparado para esperar',
                nl: 'Het verzoek duurde langer dan de server wilde wachten',
                de: 'Die Anfrage hat länger gedauert, als der Server bereit war zu warten',
                es: 'La petición esta tardando más de lo que el servidor estaba preparado para esperar',
            },
            'Conflict': {
                fr: 'Conflit',
                ru: 'Конфликт',
                uk: 'Конфлікт',
                pt: 'Conflito',
                nl: 'Conflict',
                de: 'Konflikt',
                es: 'Conflicto',
            },
            'The request could not be completed because of a conflict': {
                fr: 'La requête n’a pas pu être complétée à cause d’un conflit',
                ru: 'Запрос не может быть обработан из-за конфликта',
                uk: 'Запит не може бути оброблений через конфлікт',
                pt: 'A solicitação não pôde ser concluída devido a um conflito',
                nl: 'Het verzoek kon niet worden verwerkt vanwege een conflict',
                de: 'Die Anfrage konnte aufgrund eines Konflikts nicht abgeschlossen werden',
                es: 'La petición no ha podido ser completada por un conflicto',
            },
            'Gone': {
                fr: 'Supprimé',
                ru: 'Удалено',
                uk: 'Вилучений',
                pt: 'Removido',
                nl: 'Verdwenen',
                de: 'Verschwunden',
                es: 'Eliminado',
            },
            'The requested page is no longer available': {
                fr: 'La page demandée n’est plus disponible',
                ru: 'Запрошенная страница была удалена',
                uk: 'Запитана сторінка була видалена',
                pt: 'A página solicitada não está mais disponível',
                nl: 'De pagina is niet langer beschikbaar',
                de: 'Die angeforderte Seite ist nicht mehr verfügbar',
                es: 'La página solicitada no está ya disponible',
            },
            'Length Required': {
                fr: 'Longueur requise',
                ru: 'Необходима длина',
                uk: 'Потрібно вказати розмір',
                pt: 'Content-Length necessário',
                nl: 'Lengte benodigd',
                de: 'Länge benötigt',
                es: 'Longitud requerida',
            },
            'The "Content-Length" is not defined. The server will not accept the request without it': {
                fr: 'Le "Content-Length" n’est pas défini. Le serveur ne prendra pas en compte la requête',
                ru: 'Заголовок "Content-Length" не был передан. Сервер не может обработать запрос без него',
                uk: 'Заголовок "Content-Length" не був переданий. Сервер не може обробити запит без нього',
                pt: 'O "Content-Length" não está definido. O servidor não aceitará a solicitação sem ele',
                nl: 'De "Content-Length" is niet gespecificeerd. De server accepteert het verzoek niet zonder',
                de: 'Die "Content-Length" ist nicht definiert. Ohne sie akzeptiert der Server die Anfrage nicht',
                es: 'El "Content-Legth" no eta definido. Este servidor no aceptará la petición sin él',
            },
            'Precondition Failed': {
                fr: 'Échec de la condition préalable',
                ru: 'Условие ложно',
                uk: 'Збій під час обробки попередньої умови',
                pt: 'Falha na pré-condição',
                nl: 'Niet voldaan aan vooraf gestelde voorwaarde',
                de: 'Vorbedingung fehlgeschlagen',
                es: 'Precondición fallida',
            },
            'The pre condition given in the request evaluated to false by the server': {
                fr: 'La précondition donnée dans la requête a été évaluée comme étant fausse par le serveur',
                ru: 'Ни одно из условных полей заголовка запроса не было выполнено',
                uk: 'Жодна з умовних полів заголовка запиту не була виконана',
                pt: 'A pré-condição dada na solicitação avaliada como falsa pelo servidor',
                nl: 'De vooraf gestelde voorwaarde is afgewezen door de server',
                de: 'Die in der Anfrage angegebene Vorbedingung wird vom Server als falsch bewertet',
                es: 'La precondición ha sido evaluada como negativa para esta petición por el servidor'
            },
            'Payload Too Large': {
                fr: 'Charge trop volumineuse',
                ru: 'Слишком большой запрос',
                uk: 'Занадто великий запит',
                pt: 'Payload muito grande',
                nl: 'Aanvraag te grood',
                de: 'Anfrage zu groß',
                es: 'Carga muy grande',
            },
            'The server will not accept the request, because the request entity is too large': {
                fr: 'Le serveur ne prendra pas en compte la requête, car l’entité de la requête est trop volumineuse',
                ru: 'Сервер не может обработать запрос, так как он слишком большой',
                uk: 'Сервер не може обробити запит, оскільки він занадто великий',
                pt: 'O servidor não aceitará a solicitação porque a entidade da solicitação é muito grande',
                nl: 'De server accepteert het verzoek niet omdat de aanvraag te groot is',
                de: 'Der Server akzeptiert die Anfrage nicht, da die Datenmenge zu groß ist',
                es: 'El servidor no aceptará esta petición, porque la carga es demasiado grande',
            },
            'Requested Range Not Satisfiable': {
                fr: 'Requête non satisfaisante',
                ru: 'Диапазон не достижим',
                uk: 'Запитуваний діапазон недосяжний',
                pt: 'Intervalo Solicitado Não Satisfatório',
                nl: 'Aangevraagd gedeelte niet opvraagbaar',
                de: 'Anfrage-Bereich nicht erfüllbar',
                es: 'Intervalo solicitado no satisfactorio',
            },
            'The requested byte range is not available and is out of bounds': {
                fr: 'Le byte range demandé n’est pas disponible et est hors des limites',
                ru: 'Запрошенный диапазон данных недоступен или вне допустимых пределов',
                uk: 'Описаний діапазон даних недоступний або з допустимих меж',
                pt: 'O intervalo de bytes solicitado não está disponível e está fora dos limites',
                nl: 'De aangevraagde bytes zijn buiten het limiet',
                de: 'Der angefragte Teilbereich der Ressource existiert nicht oder ist ungültig',
                es: 'El intervalo de bytes requerido no está disponible o se encuentra fuera de los límites',
            },
            'I\'m a teapot': {
                fr: 'Je suis une théière',
                ru: 'Я чайник',
                uk: 'Я чайник',
                pt: 'Eu sou um bule',
                nl: 'Ik ben een theepot',
                de: 'Ich bin eine Teekanne',
                es: 'Soy una tetera'
            },
            'Attempt to brew coffee with a teapot is not supported': {
                fr: 'Tenter de préparer du café avec une théière n’est pas pris en charge',
                ru: 'Попытка заварить кофе в чайнике обречена на фиаско',
                uk: 'Спроба виварити каву в чайник приречена на фіаско',
                pt: 'A tentativa de preparar café com um bule não é suportada',
                nl: 'Koffie maken met een theepot is niet ondersteund',
                de: 'Der Versuch, Kaffee mit einer Teekanne zuzubereiten, wird nicht unterstützt',
                es: 'Intentar hacer un café en una tetera no está soportado',
            },
            'Too Many Requests': {
                fr: 'Trop de requêtes',
                ru: 'Слишком много запросов',
                uk: 'Занадто багато запитів',
                pt: 'Excesso de solicitações',
                nl: 'Te veel requests',
                de: 'Zu viele Anfragen',
                es: 'Demasiadas peticiones'
            },
            'Too many requests in a given amount of time': {
                fr: 'Trop de requêtes dans un délai donné',
                ru: 'Отправлено слишком много запросов за короткое время',
                uk: 'Надіслано занадто багато запитів на короткий час',
                pt: 'Excesso de solicitações em um determinado período de tempo',
                nl: 'Te veel verzoeken binnen een bepaalde tijd',
                de: 'Der Client hat zu viele Anfragen in einem bestimmten Zeitraum gesendet',
                es: 'Demasiadas peticiones en un determinado periodo de tiempo',
            },
            'Internal Server Error': {
                fr: 'Erreur interne du serveur',
                ru: 'Внутренняя ошибка сервера',
                uk: 'Внутрішня помилка сервера',
                pt: 'Erro do Servidor Interno',
                nl: 'Interne serverfout',
                de: 'Interner Server-Fehler',
                es: 'Error Interno',
            },
            'The server met an unexpected condition': {
                fr: 'Le serveur a rencontré une condition inattendue',
                ru: 'Произошло что-то неожиданное на сервере',
                uk: 'На сервері було щось несподіване',
                pt: 'O servidor encontrou uma condição inesperada',
                nl: 'De server ondervond een onverwachte conditie',
                de: 'Der Server hat einen internen Fehler festgestellt',
                es: 'El servidor ha encontrado una condición no esperada',
            },
            'Bad Gateway': {
                fr: 'Mauvaise passerelle',
                ru: 'Ошибка шлюза',
                uk: 'Помилка шлюзу',
                pt: 'Gateway inválido',
                nl: 'Ongeldige Gateway',
                de: 'Fehlerhafter Gateway',
                es: 'Puerta de enlace no valida',
            },
            'The server received an invalid response from the upstream server': {
                fr: 'Le serveur a reçu une réponse invalide du serveur distant',
                ru: 'Сервер получил некорректный ответ от вышестоящего сервера',
                uk: 'Сервер отримав неправильну відповідь з сервера Upstream',
                pt: 'O servidor recebeu uma resposta inválida do servidor upstream',
                nl: 'De server ontving een ongeldig antwoord van een bovenliggende server',
                de: 'Der Server hat eine ungültige Antwort vom Upstream-Server erhalten',
                es: 'El servidor ha recibido una respuesta no válida del servidor de origen'
            },
            'Service Unavailable': {
                fr: 'Service indisponible',
                ru: 'Сервис недоступен',
                uk: 'Сервіс недоступний',
                pt: 'Serviço não disponível',
                nl: 'Dienst niet beschikbaar',
                de: 'Dienst nicht verfügbar',
                es: 'Servicio no disponible',
            },
            'The server is temporarily overloading or down': {
                fr: 'Le serveur est temporairement en surcharge ou indisponible',
                ru: 'Сервер временно не может обрабатывать запросы по техническим причинам',
                uk: 'Сервер тимчасово не може обробляти запити з технічних причин',
                pt: 'O servidor está temporariamente sobrecarregado ou inativo',
                nl: 'De server is tijdelijk overbelast of niet bereikbaar',
                de: 'Der Server ist vorübergehend überlastet oder ausgefallen',
                es: 'El servidor está temporalmente sobrecargado o inactivo',
            },
            'Gateway Timeout': {
                fr: 'Expiration Passerelle',
                ru: 'Шлюз не отвечает',
                uk: 'Шлюз не відповідає',
                pt: 'Tempo limite do gateway excedido',
                nl: 'Gateway Verlopen',
                de: 'Gateway Zeitüberschreitung',
                es: 'Tiempo límite de puerta de enlace excedido',
            },
            'The gateway has timed out': {
                fr: 'Le temps d’attente de la passerelle est dépassé',
                ru: 'Сервер не дождался ответа от вышестоящего сервера',
                uk: 'Сервер не чекав відповіді від сервера Upstream',
                pt: 'O gateway esgotou o tempo limite',
                nl: 'De verbinding naar de bovenliggende server is verlopen',
                de: 'Das Zeitlimit für den Verbindungsaufbau mit dem Upstream-Server ist abgelaufen',
                es: 'La puerta de enlace ha sobrepasado el tiempo límite',
            },
            'HTTP Version Not Supported': {
                fr: 'Version HTTP non prise en charge',
                ru: 'Версия HTTP не поддерживается',
                uk: 'Версія НТТР не підтримується',
                pt: 'Versão HTTP não suportada',
                nl: 'HTTP-versie wordt niet ondersteunt',
                de: 'HTTP-Version wird nicht unterstützt',
                es: 'Versión de HTTP no soportada',
            },
            'The server does not support the "http protocol" version': {
                fr: 'Le serveur ne supporte pas la version du protocole HTTP',
                ru: 'Сервер не поддерживает запрошенную версию HTTP протокола',
                uk: 'Сервер не підтримує запитану версію HTTP-протоколу',
                pt: 'O servidor não suporta a versão do protocolo HTTP',
                nl: 'De server ondersteunt deze HTTP-versie niet',
                de: 'Der Server unterstützt die HTTP-Protokoll-Version nicht',
                es: 'El servidor no soporta la versión del protocolo HTTP',
            },

            'Host': {
                fr: 'Hôte',
                ru: 'Хост',
                uk: 'Хост',
                pt: 'Hospedeiro',
                nl: 'Host',
                de: 'Host',
                es: 'Host',
            },
            'Original URI': {
                fr: 'URI d’origine',
                ru: 'Исходный URI',
                uk: 'Вихідний URI',
                pt: 'URI original',
                nl: 'Originele URI',
                de: 'Originale URI',
                es: 'URI original',
            },
            'Forwarded for': {
                fr: 'Transmis pour',
                ru: 'Перенаправлен',
                uk: 'Перенаправлений',
                pt: 'Encaminhado para',
                nl: 'Doorgestuurd voor',
                de: 'Weitergeleitet für',
                es: 'Remitido por',
            },
            'Namespace': {
                fr: 'Espace de noms',
                ru: 'Пространство имён',
                uk: 'Простір імен',
                pt: 'Namespace',
                nl: 'Elementnaam',
                de: 'Namensraum',
                es: 'Namespace',
            },
            'Ingress name': {
                fr: 'Nom ingress',
                ru: 'Имя Ingress',
                uk: 'Ім\'я Ingress',
                pt: 'Nome Ingress',
                nl: 'Ingress naam',
                de: 'Ingress Name',
                es: 'Nombre Ingress',
            },
            'Service name': {
                fr: 'Nom du service',
                ru: 'Имя сервиса',
                uk: 'Ім\'я сервісу',
                pt: 'Nome do Serviço',
                nl: 'Service naam',
                de: 'Service Name',
                es: 'Nombre del servicio'
            },
            'Service port': {
                fr: 'Port du service',
                ru: 'Порт сервиса',
                uk: 'Порт сервісу',
                pt: 'Porta do serviço',
                nl: 'Service poort',
                de: 'Service Port',
                es: 'Puerto del servicio',
            },
            'Request ID': {
                fr: 'Identifiant de la requête',
                ru: 'ID запроса',
                uk: 'ID запиту',
                pt: 'ID da solicitação',
                nl: 'ID van het verzoek',
                de: 'Anfrage ID',
                es: 'ID de la petición',
            },
            'Timestamp': {
                fr: 'Horodatage',
                ru: 'Временная метка',
                uk: 'Тимчасова мітка',
                pt: 'Timestamp',
                nl: 'Tijdstempel',
                de: 'Zeitstempel',
                es: 'Timestamp',
            },

            'client-side error': {
                fr: 'Erreur Client',
                ru: 'ошибка на стороне клиента',
                uk: 'помилка на стороні клієнта',
                pt: 'erro do lado do cliente',
                nl: 'fout aan de gebruikerskant',
                de: 'Clientseitiger Fehler',
                es: 'Error del lado del cliente',
            },
            'server-side error': {
                fr: 'Erreur Serveur',
                ru: 'ошибка на стороне сервера',
                uk: 'помилка на стороні сервера',
                pt: 'erro do lado do servidor',
                nl: 'fout aan de serverkant',
                de: 'Serverseitiger Fehler',
                es: 'Error del lado del servidor',
            },

            'Your Client': {
                fr: 'Votre Client',
                ru: 'Ваш Браузер',
                uk: 'Ваш Браузер',
                pt: 'Seu Cliente',
                nl: 'Jouw Client',
                de: 'Ihr Client',
                es: 'Tu cliente',
            },
            'Network': {
                fr: 'Réseau',
                ru: 'Сеть',
                uk: 'Сіть',
                pt: 'Rede',
                nl: 'Netwerk',
                de: 'Netzwerk',
                es: 'Red',
            },
            'Web Server': {
                fr: 'Serveur Web',
                ru: 'Web Сервер',
                uk: 'Web Сервер',
                pt: 'Servidor web',
                nl: 'Web Server',
                de: 'Webserver',
                es: 'Servidor web',
            },
            'What happened?': {
                fr: 'Que s’est-il passé ?',
                ru: 'Что произошло?',
                uk: 'Що сталося?',
                pt: 'O que aconteceu?',
                nl: 'Wat is er gebeurd?',
                de: 'Was ist passiert?',
                es: '¿Que ha pasado?',
            },
            'What can i do?': {
                fr: 'Que puis-je faire ?',
                ru: 'Что можно сделать?',
                uk: 'Що можна зробити?',
                pt: 'O que eu posso fazer?',
                nl: 'Wat kan ik doen?',
                de: 'Was kann ich machen?',
                es: '¿Que puedo hacer?',
            },
            'Please try again in a few minutes': {
                fr: 'Veuillez réessayer dans quelques minutes',
                ru: 'Пожалуйста, попробуйте повторить запрос ещё раз чуть позже',
                uk: 'Будь ласка, спробуйте повторити запит ще раз трохи пізніше',
                pt: 'Por favor, tente novamente em alguns minutos',
                nl: 'Probeer het alstublieft opnieuw over een paar minuten',
                de: 'Bitte versuchen Sie es in ein paar Minuten erneut',
                es: 'Por favor, intente nuevamente en unos minutos',
            },
            'Working': {
                fr: 'Opérationnel',
                ru: 'Работает',
                uk: 'Працює',
                pt: 'Funcionando',
                nl: 'Functioneel',
                de: 'Funktioniert',
                es: 'Trabajando',
            },
            'Unknown': {
                fr: 'Inconnu',
                ru: 'Неизвестно',
                uk: 'Невідомо',
                pt: 'Desconhecido',
                nl: 'Onbekend',
                de: 'Unbekannt',
                es: 'Desconocido',
            },
            'Please try to change the request method, headers, payload, or URL': {
                fr: 'Veuillez essayer de changer la méthode de requête, les en-têtes, le contenu ou l’URL',
                ru: 'Пожалуйста, попробуйте изменить метод запроса, заголовки, его содержимое или URL',
                uk: 'Будь ласка, спробуйте змінити метод запиту, заголовки, його вміст або URL-адресу',
                pt: 'Tente alterar o método de solicitação, cabeçalhos, payload ou URL',
                nl: 'Probeer het opnieuw met een andere methode, headers, payload of URL',
                de: 'Bitte versuchen Sie, die Anfragemethode, Header, Payload oder URL zu ändern',
                es: 'Por favor intente cambiar el método de la petición, cabeceras, carga o URL'
            },
            'Please check your authorization data': {
                fr: 'Veuillez vérifier vos données d’autorisation',
                ru: 'Пожалуйста, проверьте данные авторизации',
                uk: 'Будь ласка, перевірте дані авторизації',
                pt: 'Verifique seus dados de autorização',
                nl: 'Controleer de authenticatiegegevens',
                de: 'Bitte überprüfen Sie Ihre Zugangsdaten',
                es: 'Verifique sus datos de autorización',
            },
            'Please double-check the URL and try again': {
                fr: 'Veuillez vérifier l’URL et réessayer',
                ru: 'Пожалуйста, дважды проверьте URL и попробуйте снова',
                uk: 'Будь ласка, двічі перевірте URL-адресу і спробуйте знову',
                pt: 'Verifique novamente o URL e tente novamente',
                nl: 'Controleer de URL en probeer het opnieuw',
                de: 'Bitte überprüfen Sie die URL und versuchen Sie es erneut',
                es: 'Verifique de nuevo la URL y vuelva a probar',
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
