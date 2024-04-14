Задание Авито
=============

**Тестовое задание для стажёра Backend в Авито**
------------------------------------------------

В Авито есть большое количество неоднородного контента, для которого необходимо иметь единую систему управления. В частности, необходимо показывать разный контент пользователям в зависимости от их принадлежности к какой-либо группе. Данный контент мы будем предоставлять с помощью баннеров.

**Описание задачи**
-------------------

Необходимо реализовать сервис, который позволяет показывать пользователям баннеры, в зависимости от требуемой фичи и тега пользователя, а также управлять баннерами и связанными с ними тегами и фичами.

**Общие вводные**
-----------------

**Баннер** — это документ, описывающий какой-либо элемент пользовательского интерфейса. Технически баннер представляет собой  JSON-документ неопределенной структуры.  **Тег** — это сущность для обозначения группы пользователей; представляет собой число (ID тега).  **Фича** — это домен или функциональность; представляет собой число (ID фичи).

1.  Один баннер может быть связан только с одной фичей и несколькими тегами

2.  При этом один тег, как и одна фича, могут принадлежать разным баннерам одновременно

3.  Фича и тег однозначно определяют баннер

Так как баннеры являются для пользователя вспомогательным функционалом, допускается, если пользователь в течение короткого срока будет получать устаревшую информацию. При этом существует часть пользователей (порядка 10%), которым обязательно получать самую актуальную информацию. Для таких пользователей нужно предусмотреть механизм получения информации напрямую из БД.

Инструкция по запуску Makefile:
-------------------------------

Для запуска требуется находится в директории master/main

*   Команда запуска приложения  
      
    `make run`

*   Команда запуска всех тестовых файлов  
      
    `make test`

*   Команда запуска docker контейнера  
      
    `make docker`

*   Команда запуска линтера  
      
    `make linter`

*   Команда запуска интеграционного теста на windows, (также есть закомменченные 2 вида запуска для linux)  
      
    `make integration_test`

Проблемы, с которыми столкнулся:
--------------------------------

1.  Возник сразу вопрос с тем, как обрабатывать токены, но после прочтения api файла понял, что просто стоит брать токены из headers, никакой авторизации не надо добавлять

2.  При прочтении пункта 5 основных задач, задумался над решением прописанного условия, решил, что лучше всего будет прописать на Go кэш самому, его можно увидеть в папке internal/cache

3.  Решил добавить отдельную функцию проверки токенов и прав, которые они имеют для лучшей реализации ограничений для пользователей

4.  Первую дополнительную задачу решил таким способом: пусть каждое получение ключа из кэша добавляет веса этому ключу относительно других, при чистке удаляются сравнивается вес ключа относительно общего кол-во запросов в кэш, если ключ набрал 10 процентов запросов, то он остаётся, иначе чистится.

5.  Также добавил в кэш ограничение по количеству ключей так, что при появлении 20 ключе тоже происходит чистка.

6.  Добавил в api и реализовал методы для получения трёх прошлых версий баннера и для удаления баннеров по фиче или тэгу

7.  Конфигурация линтера представлена в файле `.go-arch-lint.yml`

Описание эндпоинтов:
--------------------

Пример методов в POSTMAN есть в JSON файле [**banners.postman\_collection.json**](https://github.com/KilyakArata/test-task/blob/main/banners.postman_collection.json)


GET

`/user_banner`

получение баннера по тэгу и фиче

GET

`/banner`

получение баннеров по тэгу или фиче

POST

`/banner`

добавление баннера

PATCH

`/banner/{id}`

обновление баннера

DELETE

`/banner/{id}`

удаление баннера по id

DELETE

`/banner`

удаления баннера по фиче или тэгу

GET

`/banner/{id}`

получение прошлых версий баннера
