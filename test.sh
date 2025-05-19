#!/bin/bash

# Script di test SENZA jq per SDCC_Progetto3

set -e  # Ferma il processo in caso di errore

# === 1. Registrazione Utente ===
echo "=== 1. Registrazione Utente ==="
grpcurl -plaintext -import-path proto/ -proto user.proto -d '{"username":"pieraldo","password":"password123"}' localhost:50051 user.UserService/Register || true

# === 2. Login Utente ===
echo "=== 2. Login Utente ==="
grpcurl -plaintext -import-path proto/ -proto user.proto -d '{"username":"pieraldo","password":"password123"}' localhost:50051 user.UserService/Login || true

# === 3. Lista Eventi Disponibili ===
echo "=== 3. Lista Eventi Disponibili ==="
grpcurl -plaintext -import-path proto/ -proto ticket.proto -d '{}' localhost:50052 ticket.TicketService/ListEvents

# Fissiamo manualmente l'ID dell'evento
# (aggiorna manualmente se cambi evento nel database)
event_id="681085079e8b285943d861e0"
echo "Evento selezionato ID: $event_id"

# === 4. Acquisto Biglietto (fallimento atteso) ===
echo "=== 4. Acquisto Biglietto (fallimento atteso) ==="
grpcurl -plaintext -import-path proto/ -proto ticket.proto -d '{"eventId":"'$event_id'","quantity":1}' localhost:50052 ticket.TicketService/PurchaseTicket

sleep 5

# === 5. Lista Eventi Dopo Fallimento ===
echo "=== 5. Lista Eventi Dopo Fallimento (biglietto ripristinato) ==="
grpcurl -plaintext -import-path proto/ -proto ticket.proto -d '{}' localhost:50052 ticket.TicketService/ListEvents

# === 6. Acquisto Biglietto (successo atteso) ===
echo "=== 6. Acquisto Biglietto (successo atteso) ==="
grpcurl -plaintext -import-path proto/ -proto ticket.proto -d '{"eventId":"'$event_id'","quantity":2}' localhost:50052 ticket.TicketService/PurchaseTicket

sleep 5

# === 7. Lista Eventi Dopo Successo ===
echo "=== 7. Lista Eventi Dopo Successo (biglietto scalato) ==="
grpcurl -plaintext -import-path proto/ -proto ticket.proto -d '{}' localhost:50052 ticket.TicketService/ListEvents


# Fine test
echo "\n👌 Tutti i test eseguiti correttamente!"
