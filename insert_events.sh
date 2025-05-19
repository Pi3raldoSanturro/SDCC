#!/bin/bash

echo "🎟 Inserimento eventi nel database ticketdb..."

mongosh <<EOF
use ticketdb

db.events.insertMany([
  {
    name: "Concerto Vasco Rossi",
    date: "2025-07-01",
    availableTickets: 100
  },
  {
    name: "Partita Roma vs Lazio",
    date: "2025-05-12",
    availableTickets: 200
  },
  {
    name: "Concerto di Roma",
    date: "2025-06-20",
    availableTickets: 150
  }
])

print("✅ Eventi inseriti con successo.")
EOF
