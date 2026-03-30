# Contributing

## Формат коммитов

В проекте используется [Conventional Commits](https://www.conventionalcommits.org/). Каждый коммит в PR проверяется линтером автоматически.

### Формат

```
<тип>(scope): описание
```

`scope` — необязательный, указывает на область изменений.

### Типы

| Тип | Когда использовать |
|---|---|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `docs` | Изменения в документации |
| `style` | Форматирование, пробелы, точки с запятой (не влияет на логику) |
| `refactor` | Рефакторинг без изменения поведения |
| `test` | Добавление или изменение тестов |
| `chore` | Обновление зависимостей, CI, конфигов |
| `perf` | Улучшение производительности |
| `ci` | Изменения в CI/CD |
| `build` | Изменения в системе сборки (Docker, go.mod) |

### Примеры

```
feat(jira): add issue sync endpoint
fix(db): handle nil pointer on empty query result
chore: update Go dependencies
ci: add commitlint workflow
docs: add contributing guide
refactor(config): simplify database connection setup
test(db): add unit tests for read functions
build: update Dockerfile to multi-stage
```

### PR

- Название PR должно следовать тому же формату: `feat(jira): add issue sync endpoint`
- Один PR — одна логическая задача
- Перед мержем убедись, что CI зеленый
