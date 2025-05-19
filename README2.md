# SDCC - Sistema di Gestione Biglietti con Microservizi

## Descrizione del progetto

Questo progetto implementa un **sistema distribuito** per la **gestione dei biglietti di eventi** tramite un'architettura a **microservizi**. Ogni microservizio è isolato, scalabile e comunica tramite **gRPC** ed **eventi asincroni** su **RabbitMQ**.

L'obiettivo è:
- Gestire **prenotazione biglietti**.
- Gestire **pagamenti** simulati.
- Implementare **rollback automatico** in caso di fallimento del pagamento.
- Dimostrare l'uso di **RabbitMQ** per una **Saga pattern semplificata**.
- Introdurre **autenticazione e autorizzazione tramite JWT**.

## Architettura dei servizi

- **User-Service**
    - Gestione utenti (registrazione, login).
    - Richiede token JWT da Auth-Service.
    - MongoDB come database utenti.
    - Porta: `50051`

- **Auth-Service**
    - Microservizio dedicato alla **generazione e validazione di token JWT**.
    - Porta: `50054`

- **Ticket-Service**
    - Gestione eventi e biglietti disponibili.
    - Prenotazione biglietti (con decremento disponibilità).
    - Ripristino biglietti in caso di pagamento fallito.
    - MongoDB come database eventi.
    - RabbitMQ per comunicazione asincrona.
    - Circuit Breaker per tolleranza ai guasti.
    - Porta: `50052`

- **Payment-Service**
    - Simulazione del pagamento.
    - Salvataggio esito transazione.
    - Invio esito su RabbitMQ.
    - MongoDB come database pagamenti.
    - Circuit Breaker attivo.
    - Porta: `50053`

- **RabbitMQ**
    - Code usate:
        - `ticket-reserved-queue`
        - `payment-events-queue`

## Comandi principali

### Build dei servizi

```bash
docker-compose build user-service ticket-service payment-service auth-service
```

### Avvio dell'infrastruttura

```bash
docker-compose up -d
```

(Assicurati che MongoDB e RabbitMQ siano completamente avviati prima dei microservizi)

### Compilazione dei file proto

In caso di modifica ai `.proto` files:

```bash
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/user.proto
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/ticket.proto
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/payment.proto
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/auth.proto
```

### Esecuzione client

```bash
go run client/main.go
```

Il client supporta:
- Login e registrazione.
- Visualizzazione eventi.
- Acquisto biglietti.
- Azioni riservate all'admin.
- **Verifica della validità del token JWT (Auth-Service)**.

## Tecnologie utilizzate

- **Go (golang)** 1.22
- **gRPC** per comunicazione tra servizi
- **JWT** per autenticazione e autorizzazione (Auth-Service)
- **MongoDB** per la persistenza
- **RabbitMQ** per eventi asincroni
- **Docker + Docker Compose** per il deploy locale

## Prossimi miglioramenti

- ✅ Aggiunta di **UserId** reale e `EventInstanceId` per de-duplicazione eventi.
- ✅ Circuit Breaker per RabbitMQ (fault-tolerant messaging).
- ✅ Implementazione **Auth-Service** per JWT.
- ✅ Validazione token nel client (via Auth-Service).
- ⏳ Integrazione token JWT direttamente nei microservizi (es. Ticket-Service).
- ⏳ Implementazione access control nei microservizi usando `role`.

## Flusso di lavoro dei servizi

```
[Client] → [User-Service] → Richiesta Login/Register
                          → Richiesta JWT a Auth-Service
                          ← Token JWT restituito

[Client] → [Ticket-Service] → gRPC PurchaseTicket
                          → Publish RabbitMQ (ticket-reserved-queue)

[RabbitMQ] → [Payment-Service]
                          → Simulazione pagamento
                          → Publish RabbitMQ (payment-events-queue)

[RabbitMQ] → [Ticket-Service]
                          → Se SUCCESS: lascia quantità scalata
                          → Se FAIL: ripristina biglietti

[Client] → [Auth-Service] (opzione 5) → Validazione Token
```