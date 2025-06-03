# SDCC – Sistema di Biglietteria Distribuito a Microservizi

Questo progetto implementa un'applicazione distribuita per la **gestione di biglietti per eventi** usando un'architettura **a microservizi**, con comunicazione via **gRPC** e orchestrazione asincrona tramite **RabbitMQ** (pattern Saga - choreography).

## Architettura

- `user-service` – registrazione/login utenti, autenticazione via JWT
- `auth-service` – generazione e validazione dei token JWT
- `ticket-service` – gestione eventi e acquisto biglietti
- `payment-service` – elaborazione dei pagamenti, pubblicazione esito
- `web-ui` – interfaccia Flask per utenti e admin
- `rabbitmq` – message broker (publish/subscribe)
- `mongo-` – database MongoDB per ciascun microservizio

## Requisiti

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- Porta `8080` libera per la web UI

## Avvio del progetto 

**Build del progetto**
   ```bash
      docker compose up --build
   ```
é possibile eseguire il progetto tramite docker swarm, in questo caso è necessario eseguire il comando:
   ```bash
      docker swarm init
      docker stack deploy -c docker-compose.swarm.yml sdcc
   ```
Per rimuovere lo stack in swarm:
```bash
      docker stack rm sdcc
```

## Testing

Per il testing dell'applicazione è possibile utilizzare la web UI per un utilizzo realistico, collegandosi a `http://localhost:8080`.
Per il testing della tolleranza ai guasti è possibile utilizzare il client creato appositamente per testare i vari scenari eseguendo:
```bash
  cd SDCC_Progetto/client
  go run main.go
```
Per spegnere i vari microservizi è possibile utilizzare il comando:
```bash
  docker compose stop
```
inserendo il nome del servizio che si vuole fermare, ad esempio:
```bash
  docker compose stop user-service
```

## Report

Nella cartella Report è possibile visualizzare la relazione del progetto in formato .pdf
