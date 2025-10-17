# Использование транзакций в проекте

## Структура использования транзакций

Транзакции в проекте используются на разных уровнях архитектуры с четкой иерархией:

```
РЕПОЗИТОРИЙ (Repository)
    ├─ Простой метод: GetByNumber(ctx Context, ...) - БЕЗ транзакции
    └─ TX-метод: UpdateStatusTx(*sqlx.Tx, ...) - ВНУТРИ транзакции

СЕРВИС (Service)
    ├─ Простой метод: GetOrder(ctx Context, ...) - БЕЗ транзакции
    └─ TX-метод: CompleteOrderTx(*sqlx.Tx, ...) - ВНУТРИ транзакции

ЮЗКЕЙС (UseCase)
    └─ TransactionManager.Execute(ctx, func(tx *sqlx.Tx) error { ... })
       └─ Оркестрирует все операции через TX-методы
```

---

## Пример 1: Работа на уровне Репозитория

### Без транзакции (БЕЗ контроля консистентности)

```go
// ❌ НЕ гарантирует атомарность
func (r *orderRepository) GetByNumber(ctx context.Context, number string) (*domain.Order, error) {
    // Просто читаем из БД
    return r.db.GetContext(ctx, ...)
}
```

**Используется когда:**
- Просто читаем данные
- Выполняем одну изолированную операцию

**Проблема:** Нет гарантии консистентности, если нужно выполнить несколько операций

---

### Внутри транзакции (ГАРАНТИРУЕТ консистентность)

```go
// ✅ ГАРАНТИРУЕТ атомарность
func (r *orderRepository) UpdateStatusTx(tx *sqlx.Tx, number, status string) error {
    // Используем TX вместо DB
    query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE number = $2`
    _, err := tx.Exec(query, status, number)  // tx вместо r.db
    return err
}
```

**Ключевые отличия:**
- Принимает `*sqlx.Tx` вместо использования `r.db`
- Использует обычные методы `tx.Exec()` вместо `tx.ExecContext()`
- Не закрывает транзакцию (она управляется на более высоком уровне)

**Используется когда:**
- Операция вызывается как часть большой транзакции
- Нужна консистентность с другими операциями

---

## Пример 2: Работа на уровне Сервиса

### Без транзакции (простая бизнес-логика)

```go
// ❌ БЕЗ контроля консистентности
func (s *orderService) CompleteOrder(ctx context.Context, number string) error {
    // Просто вызываем репозиторий
    return s.orderRepo.UpdateStatus(ctx, number, "completed")
}
```

**Используется для:**
- Простых операций без сложной логики
- Операций, которые не требуют координации с другими

---

### Передача TX из Юзкейса

```go
// ✅ ГАРАНТИРУЕТ атомарность при вызове из юзкейса
func (s *orderService) CompleteOrderTx(tx *sqlx.Tx, number string) error {
    // Передаем TX в репозиторий
    return s.orderRepo.UpdateStatusTx(tx, number, "completed")
}
```

**Ключевые моменты:**
- Принимает `*sqlx.Tx` вместо `context.Context`
- Передает TX непосредственно в репозиторий
- Не содержит сложной бизнес-логики (только координация)

---

### Собственный TransactionManager в Сервисе (когда операции ВСЕГДА неделимы)

```go
// ✅ Пример: Перевод между счетами - ВСЕГДА выполняется атомарно
func (s *balanceService) TransferBetweenAccounts(
    ctx context.Context, 
    fromID, toID int64, 
    amount float64,
) error {
    return s.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // Эти две операции НИКОГДА не разделяются
        if err := s.repo.WithdrawTx(tx, fromID, amount); err != nil {
            return err  // ROLLBACK обе
        }
        return s.repo.DepositTx(tx, toID, amount)  // COMMIT обе
    })
}
```

**Когда это имеет смысл:**
- Операции внутри Сервиса **логически неделимы**
- Они **ВСЕГДА** выполняются вместе
- Никогда не комбинируются с другими операциями в Юзкейсе

---

## Пример 3: Работа на уровне Юзкейса (ГДЕ ЖИВУТ ТРАНЗАКЦИИ)

### ✅ ПРАВИЛЬНО: Использование TransactionManager

```go
func (uc *OrderUseCase) CompleteOrderAndAccrueBalance(
    ctx context.Context, 
    orderNumber string, 
    userID int64, 
    accrualAmount float64,
) error {
    // ШАГ 1: Читаем данные БЕЗ транзакции (просто информация)
    order, err := uc.orderService.GetOrder(ctx, orderNumber)
    if err != nil {
        return err
    }

    // ШАГ 2: Используем TransactionManager для АТОМАРНОГО выполнения
    err = uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // ВСЕ эти операции выполняются В ОДНОЙ транзакции
        
        // Операция 1: Завершить заказ
        if err := uc.orderService.CompleteOrderTx(tx, orderNumber); err != nil {
            return err  // ← ROLLBACK: обе операции откатываются
        }

        // Операция 2: Добавить деньги пользователю
        if err := uc.userService.AddBalanceToUserTx(tx, userID, accrualAmount); err != nil {
            return err  // ← ROLLBACK: обе операции откатываются
        }

        return nil  // ← COMMIT: обе операции закреплены
    })

    return err
}
```

**Что происходит внутри TransactionManager:**

```
1. BEGIN TRANSACTION
2. Вызывает функцию (tx *sqlx.Tx) error
3. Если функция вернула ошибку:
   - ROLLBACK всё
4. Если функция успешна:
   - COMMIT всё
5. Если произойдет паника:
   - ROLLBACK всё и пробросить панику
```

---

## Разница между использованиями

| Уровень | Без транзакции | С транзакцией | Используется для |
|---------|---|---|---|
| **Репозиторий** | `GetByNumber(ctx, ...)` | `UpdateStatusTx(tx, ...)` | Работа с БД |
| **Сервис** | `GetOrder(ctx, ...)` | `CompleteOrderTx(tx, ...)` | Бизнес-логика отдельных объектов |
| **Юзкейс** | Для чтения | `tm.Execute(ctx, func(tx) {})` | Оркестрация, консистентность |

---

## Кейсы использования

### ✅ ИСПОЛЬЗУЙ транзакцию, если:

1. **Нужна консистентность данных**
   ```go
   // Нельзя завершить заказ БЕЗ добавления денег пользователю
   uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
       // Обе операции или обе откатываются
   })
   ```

2. **Несколько операций должны быть атомарными**
   ```go
   // Либо все, либо ничего
   uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
       op1 := uc.orderService.ProcessOrderTx(tx, orderNum)
       op2 := uc.userService.AddBalanceToUserTx(tx, userID, amount)
       // ...
   })
   ```

3. **Нужна защита от race conditions**
   ```go
   // Транзакция гарантирует, что никто не изменит данные между проверкой и обновлением
   ```

### ❌ НЕ используй транзакцию, если:

1. **Только чтение данных**
   ```go
   order, _ := uc.orderService.GetOrder(ctx, number)  // БЕЗ tx
   ```

2. **Одна операция, не связанная с другими**
   ```go
   _ = uc.orderService.CompleteOrder(ctx, number)  // БЕЗ tx
   ```

3. **Операции независимы и ошибка в одной не отменяет другие**
   ```go
   // Отправить уведомление юзеру
   _ = uc.notificationService.Send(ctx, userID, msg)  // БЕЗ tx
   ```

---

## Практический пример: Обработка заказа

### Сценарий: Пользователь купил товар

```
1. Заказ имеет статус "pending"
2. Система должна:
   - Изменить статус на "processing"
   - Зарезервировать деньги у пользователя
   - Если обе операции успешны → всё в порядке
   - Если одна из них упадет → откатить обе

Без транзакции:
❌ Меняем статус
❌ Но не можем зарезервировать деньги (ошибка БД)
❌ Теперь заказ в статусе "processing", но деньги не зарезервированы
❌ ДАННЫЕ НЕСОГЛАСОВАННЫЕ!

С транзакцией:
✅ Начинаем транзакцию
✅ Меняем статус
✅ Зарезервируем деньги
✅ Коммитим - всё согласованно
```

### Реализация в юзкейсе:

```go
func (uc *OrderUseCase) ProcessOrderAndReserveBalance(
    ctx context.Context,
    orderNumber string,
    userID int64,
    reserveAmount float64,
) error {
    // Читаем данные БЕЗ транзакции
    order, _ := uc.orderService.GetOrder(ctx, orderNumber)
    user, _ := uc.userService.GetUser(ctx, userID)
    
    // Проверяем валидность БЕЗ транзакции
    if user.Balance < reserveAmount {
        return fmt.Errorf("insufficient balance")
    }

    // ВСЕ изменения ВНУТРИ транзакции
    return uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // Операция 1: Изменить статус заказа
        if err := uc.orderService.ProcessOrderTx(tx, orderNumber); err != nil {
            return err
        }

        // Операция 2: Зарезервировать деньги
        newBalance := user.Balance - reserveAmount
        if err := uc.userService.SetBalanceTx(tx, userID, newBalance); err != nil {
            return err
        }

        // Обе операции выполнены? → COMMIT
        // Одна упала? → ROLLBACK обе
        return nil
    })
}
```

---

## Правила для транзакций

### 1. **Где использовать TransactionManager.Execute()?**

**Рекомендуется в Юзкейсе** — когда координируешь **несколько операций в разных Сервисах**:
```
ЮЗКЕЙС:
├─ Читать данные БЕЗ tx
├─ Проверить валидность БЕЗ tx
└─ Изменять данные ВНУТРИ tx ← TransactionManager.Execute()
```

**Может быть в Сервисе** — когда операции **ВСЕГДА логически неделимы**:
- Перевод между счетами (снять + положить вместе)
- Резервирование товара (уменьшить + заблокировать вместе)
- Другие неделимые бизнес-операции

### 2. **Методы с суффиксом `Tx`**
```go
// Каждый слой предлагает ОБА варианта:
func (s *Service) Operation(ctx context.Context, ...) error { }      // БЕЗ tx
func (s *Service) OperationTx(tx *sqlx.Tx, ...) error { }           // С tx
```

Это позволяет использовать операцию:
- Независимо (БЕЗ tx версия)
- Как часть большой транзакции (С tx версия)

### 3. **Ошибки откатываются автоматически**
```go
err := uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
    op1()  // ✅
    op2()  // ❌ Ошибка
    // Не добираемся сюда, TX откатилась
    return err
})
// op1 откачена вместе с op2!
```

### 4. **Не вкладывай одну транзакцию в другую**
```go
// ❌ НЕПРАВИЛЬНО - вложенные транзакции
uc.transactionManager.Execute(ctx, func(tx1 *sqlx.Tx) error {
    uc.transactionManager.Execute(ctx, func(tx2 *sqlx.Tx) error {
        // Конфликт!
    })
})

// ✅ ПРАВИЛЬНО - одна транзакция для всех
uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
    op1Tx(tx)
    op2Tx(tx)
    op3Tx(tx)
})
```

---

## Резюме

| Слой | Роль |
|------|---|
| **Репозиторий** | Предоставляет методы `...Tx()` для работы внутри транзакций |
| **Сервис** | Предоставляет методы `...Tx()` и может использовать `TransactionManager` для логически неделимых операций |
| **Юзкейс** | **Основное место** использования `TransactionManager.Execute()` для координации нескольких операций |

**Ключевой принцип:**
```
Используй TransactionManager.Execute() когда:
✅ Нужно координировать несколько операций в разных Сервисах (Юзкейс)
✅ Операции внутри Сервиса ВСЕГДА выполняются вместе (Сервис)

❌ Не используй вложенные транзакции
❌ Не вкладывай операции из разных TX-блоков
```