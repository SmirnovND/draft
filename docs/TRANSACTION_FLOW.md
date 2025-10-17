# Поток данных при использовании транзакций

## Визуальная схема: CompleteOrderAndAccrueBalance

```
┌─────────────────────────────────────────────────────────────────────┐
│ HTTP REQUEST                                                        │
│ POST /orders/complete                                               │
│ {"order_number":"ORD123", "user_id":1, "accrual_amount":100.50}    │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────┐
        │ OrderController                │
        │ CompleteOrder()                │
        └────────────┬───────────────────┘
                     │
                     │ вызывает
                     ▼
    ┌────────────────────────────────────┐
    │ OrderUseCase                       │
    │ CompleteOrderAndAccrueBalance()    │
    │                                    │
    │ ШАГ 1: Чтение БЕЗ TX             │
    │ ┌────────────────────────────────┐ │
    │ │ GetOrder(ctx, orderNumber)     │ │
    │ │ ↓                              │ │
    │ │ OrderService.GetOrder()        │ │
    │ │ ↓                              │ │
    │ │ OrderRepository.GetByNumber()  │ │
    │ │ ↓                              │ │
    │ │ SELECT * FROM orders WHERE ... │ │
    │ └────────────────────────────────┘ │
    │                                    │
    │ ШАГ 2: TransactionManager.Execute()│
    │ ┌────────────────────────────────┐ │
    │ │ BEGIN TRANSACTION              │ │
    │ └────────────────────────────────┘ │
    │                                    │
    │ ┌────────────────────────────────┐ │
    │ │ ОПЕРАЦИЯ 1: CompleteOrderTx()  │ │
    │ │ ├─ OrderService.CompleteOrderTx(tx)     │ │
    │ │ └─ OrderRepository.UpdateStatusTx(tx)   │ │
    │ │    └─ UPDATE orders SET status='completed' │ │
    │ └────────────────────────────────┘ │
    │                                    │
    │ ┌────────────────────────────────┐ │
    │ │ ОПЕРАЦИЯ 2: AddBalanceTx()     │ │
    │ │ ├─ UserService.AddBalanceToUserTx(tx)   │ │
    │ │ └─ UserRepository.IncrementBalanceTx(tx)│ │
    │ │    └─ UPDATE users SET balance=balance+100.50 │ │
    │ └────────────────────────────────┘ │
    │                                    │
    │ ШАГ 3: Результат                   │
    │ ┌─── УСПЕХ ──────────────────────┐ │
    │ │ COMMIT TRANSACTION             │ │
    │ │ ✅ Оба UPDATE'а выполнены      │ │
    │ └────────────────────────────────┘ │
    │                                    │
    │ ┌─── ОШИБКА ─────────────────────┐ │
    │ │ ROLLBACK TRANSACTION           │ │
    │ │ ⚠️  Оба UPDATE'а отменены      │ │
    │ └────────────────────────────────┘ │
    │                                    │
    └────────────────────┬───────────────┘
                         │
                         ▼
        ┌────────────────────────────┐
        │ HTTP RESPONSE              │
        │ 200 OK или 500 Error       │
        └────────────────────────────┘
```

---

## Сценарий 1: Успешное выполнение

```
ШАГИ:                              БД (ДО)                БД (ПОСЛЕ)

1. Читаем заказ БЕЗ TX
   SELECT * FROM orders ...

2. BEGIN TRANSACTION
   🔒 Таблицы заблокированы

3. UPDATE orders (операция 1)      orders.status="pending" → "completed"
   UPDATE users (операция 2)       users.balance=1000      → 1100.50

4. COMMIT
   🔓 Таблицы разблокированы       ✅ Данные сохранены
```

**Состояние БД:**
- Статус заказа: `pending` → `completed` ✅
- Баланс пользователя: `1000.00` → `1100.50` ✅
- **Консистентность:** ✅ Обе операции выполнены вместе

---

## Сценарий 2: Ошибка при выполнении второй операции

```
ШАГИ:                              БД (ДО)                БД (ПОСЛЕ)

1. Читаем заказ БЕЗ TX
   SELECT * FROM orders ...

2. BEGIN TRANSACTION
   🔒 Таблицы заблокированы

3. UPDATE orders (операция 1)      orders.status="pending" → "completed"
   ✅ Успешно

4. UPDATE users (операция 2)       ❌ ОШИБКА!
   Пример: user_id=999 не существует
   ERROR: violates foreign key constraint

5. ROLLBACK
   🔓 Таблицы разблокированы       ⚠️  Обе операции отменены
   
   orders.status: "completed" → "pending" (откачена)
   users.balance: не изменилась
```

**Состояние БД:**
- Статус заказа: остается `pending` ✅ (откачена)
- Баланс пользователя: не изменился ✅ (откачена)
- **Консистентность:** ✅ Либо обе операции, либо ничего (ACID)

---

## Сценарий 3: БЕЗ транзакции (ПРОБЛЕМА!)

```
⚠️  БЕЗ ТРАНЗАКЦИИ:

ШАГИ:                              БД (ДО)                БД (ПОСЛЕ)

1. UPDATE orders (операция 1)      orders.status="pending" → "completed"
   ✅ Успешно СРАЗУ коммитится

2. UPDATE users (операция 2)       ❌ ОШИБКА!
   Например: соединение упало

РЕЗУЛЬТАТ: НЕСОГЛАСОВАННЫЕ ДАННЫЕ!
- Статус заказа: "completed" ❌ (выполнен, но деньги не начислены)
- Баланс пользователя: 1000.00 ❌ (не изменился)
- Пользователь получит заказ БЕЗ оплаты! ⚠️
```

---

## Жизненный цикл транзакции

```
┌─────────────────────────────────────────────────────────────────┐
│ TransactionManager.Execute(ctx, func(tx *sqlx.Tx) error { ... }) │
└─────────────────────┬───────────────────────────────────────────┘
                      │
        ┌─────────────▼──────────┐
        │ 1. db.BeginTxx(ctx)    │ ← BEGIN TRANSACTION
        │ Создается TX объект    │
        └─────────────┬──────────┘
                      │
        ┌─────────────▼──────────────────────┐
        │ 2. Вызывается fn(tx)               │
        │ fn := func(tx *sqlx.Tx) error {}   │
        │                                   │
        │ ВНУТРИ fn:                         │
        │ - tx.Exec()                        │ ← Операция 1
        │ - tx.Exec()                        │ ← Операция 2
        │ - tx.Exec()                        │ ← Операция 3
        │                                   │
        └─────────────┬──────────────────────┘
                      │
        ┌─────────────▼──────────────────────┐
        │ 3. defer обработка результатов     │
        │                                   │
        │ если fn вернула ошибку:            │
        │   tx.Rollback() ← ROLLBACK         │
        │                                   │
        │ если fn успешна:                   │
        │   tx.Commit() ← COMMIT             │
        │                                   │
        │ если паника внутри fn:             │
        │   tx.Rollback() ← ROLLBACK         │
        │   panic(r) ← пробросить панику    │
        │                                   │
        └─────────────┬──────────────────────┘
                      │
        ┌─────────────▼──────────┐
        │ 4. Возвращаем ошибку   │
        │ (если была)            │
        └────────────────────────┘
```

---

## Детальный пример: шаг за шагом

### Код:
```go
err := uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
    // Операция 1
    err1 := uc.orderService.CompleteOrderTx(tx, "ORD123")
    if err1 != nil {
        return err1  // ← ROLLBACK
    }

    // Операция 2
    err2 := uc.userService.AddBalanceToUserTx(tx, 1, 100.50)
    if err2 != nil {
        return err2  // ← ROLLBACK
    }

    return nil  // ← COMMIT
})
```

### Выполнение:

| Шаг | Что происходит | SQL | Состояние TX |
|-----|---|---|---|
| 1 | `BeginTxx()` | `BEGIN` | 🔒 Активна |
| 2 | `CompleteOrderTx()` | `UPDATE orders SET status='completed'` | 🔒 Активна |
| 3 | Проверка ошибки | - | 🔒 Активна |
| 4 | `AddBalanceToUserTx()` | `UPDATE users SET balance=balance+100.50` | 🔒 Активна |
| 5 | Проверка ошибки | - | 🔒 Активна |
| 6 | `return nil` | - | 🔒 Активна |
| 7 | defer: успех | `COMMIT` | ✅ Коммитена |
| 8 | Возврат | - | ✅ Завершена |

---

## Архитектурная схема в проекте

```
HTTP Layer
    ↓
┌───────────────────────────┐
│ OrderController           │
│ CompleteOrder(w, r)       │
└──────────────┬────────────┘
               │ Инжектируется через DI
               ▼
┌───────────────────────────────────────┐
│ OrderUseCase                          │
│                                       │
│ CompleteOrderAndAccrueBalance()       │
│ ├─ Читает данные БЕЗ TX               │
│ │  └─ orderService.GetOrder(ctx, ...)│
│ │                                    │
│ └─ ЗАПУСКАЕТ TX через TransactionManager:
│    ├─ orderService.CompleteOrderTx(tx, ...)
│    │  └─ orderRepository.UpdateStatusTx(tx, ...)
│    │
│    └─ userService.AddBalanceToUserTx(tx, ...)
│       └─ userRepository.IncrementBalanceTx(tx, ...)
└───────────────────────────────────────┘
               ↓
┌──────────────────────────────────┐
│ TransactionManager               │
│ Execute(ctx, func(tx) { ... })   │
│                                  │
│ ├─ BEGIN                         │
│ ├─ fn(tx)                        │
│ ├─ COMMIT или ROLLBACK           │
│ └─ Возврат ошибки (если была)    │
└──────────────────────────────────┘
               ↓
        PostgreSQL DB
```

---

## Правило: Где живут транзакции?

### ✅ ДА: TransactionManager в Юзкейсе

```go
// internal/usecases/order_usecase.go
func (uc *OrderUseCase) CompleteOrderAndAccrueBalance(...) error {
    return uc.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // ✅ ПРАВИЛЬНО!
    })
}
```

### ❌ НЕТ: TransactionManager В Сервисе

```go
// internal/services/order_service.go
func (s *OrderService) Complete(...) error {
    // ❌ НЕПРАВИЛЬНО!
    return s.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // ...
    })
}
```

**Почему?**
- Сервис не должен знать о транзакциях
- Транзакции - это дело Юзкейса (бизнес-логика)
- Сервис просто предоставляет операции

### ❌ НЕТ: TransactionManager В Репозитории

```go
// internal/repositories/order_repository.go
func (r *orderRepository) Update(...) error {
    // ❌ НЕПРАВИЛЬНО!
    return r.transactionManager.Execute(ctx, func(tx *sqlx.Tx) error {
        // ...
    })
}
```

**Почему?**
- Репозиторий только работает с БД
- Он не должен управлять транзакциями
- TX передается ему готовым объектом

---

## Контрольный список при использовании транзакций

- [ ] TransactionManager используется ТОЛЬКО в Юзкейсе
- [ ] Сервис предоставляет ОБА варианта: `Operation()` и `OperationTx()`
- [ ] Репозиторий имеет TX-методы с суффиксом `Tx`
- [ ] TX-методы принимают `*sqlx.Tx` и НЕ используют контекст
- [ ] ОБА вызова сервиса внутри TX получают ОДИН TX объект
- [ ] Ошибка в одной операции откатывает ВСЕ операции
- [ ] Нет вложенных транзакций (`Execute` внутри `Execute`)
- [ ] Все операции, требующие консистентности, внутри одного `Execute()`