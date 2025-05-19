# SDCC - Sistema di Gestione Biglietti con Microservizi

## Descrizione del progetto

Questo progetto implementa un **sistema distribuito** per la **gestione dei biglietti di eventi** tramite un'architettura a **microservizi**. Ogni microservizio è isolato, scalabile e comunica tramite **gRPC** ed **eventi asincroni** su **RabbitMQ**.

L'obiettivo è:
- Gestire **prenotazione biglietti**.
- Gestire **pagamenti** simulati.
- Implementare **rollback automatico** in caso di fallimento del pagamento.
- Dimostrare l'uso di **RabbitMQ** per una **Saga pattern semplificata**.

---

## Architettura dei servizi

- **User-Service**
    - Gestione utenti (registrazione, login).
    - MongoDB come database utenti.
    - Porta: `50051`

- **Ticket-Service**
    - Gestione eventi e biglietti disponibili.
    - Prenotazione biglietti (con decremento disponibilità immediato).
    - Ripristino biglietti in caso di pagamento fallito.
    - MongoDB come database eventi.
    - RabbitMQ per inviare eventi di prenotazione e ricevere esito pagamento.
    - Porta: `50052`

- **Payment-Service**
    - Simulazione pagamento.
    - Salvataggio esito pagamento.
    - Comunicazione asincrona del risultato tramite RabbitMQ.
    - MongoDB come database pagamenti.
    - Porta: `50053`

- **RabbitMQ**
    - Code usate:
        - `ticket-reserved-queue`
        - `payment-events-queue`


---

## Comandi principali

### Build dei servizi

```bash
docker-compose build user-service ticket-service payment-service
```

### 📅 Avvio dell'infrastruttura

```bash
docker-compose up -d
```

(Assicurati che MongoDB e RabbitMQ partano correttamente)


### Compilazione dei file proto

In caso di modifica ai `.proto` files:

```bash
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/user.proto
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/ticket.proto
protoc --proto_path=proto/ --go_out=proto/ --go-grpc_out=proto/ proto/payment.proto
```


### Esecuzione test automatico

Abbiamo creato un **test.sh** che simula il flusso completo:

```bash
chmod +x test.sh
./test.sh
```

Il test esegue:
- Registrazione utente
- Login
- Lista eventi disponibili
- Prenotazione con pagamento fallito (e ripristino biglietti)
- Prenotazione con pagamento riuscito (e decremento biglietti)

---

## Tecnologie utilizzate

- **Go (golang)** 1.22
- **gRPC** per comunicazione microservizi
- **MongoDB** come database di persistenza
- **RabbitMQ** per la gestione eventi asincroni
- **Docker + Docker Compose** per il deploy locale


---

## Prossimi miglioramenti

- ✅ Aggiunta di **UserId** reale in TicketReservedEvent.
- ✅ Aggiunta di **Circuit Breaker** per aumentare la tolleranza ai guasti.
- ✅ Implementazione di un **Validation-Service** (es. AWS Lambda).
- ✅ Scrittura di **Retry Policies** sui consumer RabbitMQ.


---

## Flusso di lavoro dei servizi

```
Utente --> [User-Service] --(gRPC Login)--> OK
       --> [Ticket-Service] --(gRPC Purchase)--> Riserva biglietto
                                          --> Invia evento su RabbitMQ
RabbitMQ (ticket-reserved-queue)
       --> [Payment-Service] --(Simula pagamento)
                                          --> Invia risultato pagamento
RabbitMQ (payment-events-queue)
       --> [Ticket-Service]
             - Se SUCCESS --> lascia decrementato
             - Se FAIL --> ripristina biglietto
```

---

