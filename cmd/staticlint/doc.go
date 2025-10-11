/*
Package staticlint содержит кастомный multichecker.

Запуск:

	go run cmd/staticlint/main.go <пакеты для анализа>

Этот multichecker включает:
1. Стандартные анализаторы (printf, shadow, structtag, unreachable)
2. Анализаторы SA из staticcheck.io
3. Дополнительный анализатор ST1000 (ошибки именования структур)
4. Два публичных анализатора:
  - nilerr: обнаружение игнорирования ошибок
  - bodyclose: проверка закрытия HTTP-ответов

5. Кастомный анализатор exitchecker, запрещающий os.Exit в main.

Пример использования:

	go run cmd/staticlint/main.go ./...
*/
package main
