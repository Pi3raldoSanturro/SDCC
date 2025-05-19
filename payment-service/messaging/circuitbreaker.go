package messaging

import (
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type CircuitBreaker struct {
	sync.Mutex
	state       string // "closed", "open"
	lastFailure time.Time
	cooldown    time.Duration
	Conn        *amqp091.Connection
	Channel     *amqp091.Channel
	QueueNames  []string
	rabbitmqURL string
}

// NewCircuitBreaker creates a breaker with initial channel setup
func NewCircuitBreaker(rabbitmqURL string, queues []string) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:       "closed",
		cooldown:    10 * time.Second,
		QueueNames:  queues,
		rabbitmqURL: rabbitmqURL,
	}
	cb.reconnect()
	return cb
}

// IsOpen checks if circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.Lock()
	defer cb.Unlock()
	if cb.state == "open" && time.Since(cb.lastFailure) > cb.cooldown {
		log.Println("⏳ Cooldown scaduto. Provo a riaprire il circuito...")
		return false
	}
	return cb.state == "open"
}

// MarkFailure sets breaker to open
func (cb *CircuitBreaker) MarkFailure() {
	cb.Lock()
	defer cb.Unlock()
	cb.state = "open"
	cb.lastFailure = time.Now()
	log.Println("🛑 Circuit breaker: connessione chiusa. Entra in stato OPEN.")
}

// TryRecover attempts to restore the connection and channel
func (cb *CircuitBreaker) TryRecover() bool {
	cb.Lock()
	defer cb.Unlock()
	return cb.reconnect()
}

// reconnect attempts to open a new connection/channel
func (cb *CircuitBreaker) reconnect() bool {
	log.Println("🔁 Tentativo di riconnessione a RabbitMQ...")

	conn, err := amqp091.Dial(cb.rabbitmqURL)
	if err != nil {
		log.Printf("❌ Connessione RabbitMQ fallita: %v", err)
		return false
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("❌ Apertura canale fallita: %v", err)
		conn.Close()
		return false
	}

	// Dichiara di nuovo le code
	for _, q := range cb.QueueNames {
		_, err := ch.QueueDeclare(q, true, false, false, false, nil)
		if err != nil {
			log.Printf("❌ Fallita dichiarazione coda '%s': %v", q, err)
			ch.Close()
			conn.Close()
			return false
		}
	}

	cb.Conn = conn
	cb.Channel = ch
	cb.state = "closed"

	if cb.Channel == nil {
		log.Println("❌ [DEBUG] cb.Channel è ancora nil dopo la riconnessione!")
		return false
	}

	log.Println("✅ Riconnessione e riapertura canale riuscita. Circuit breaker CHIUSO.")
	return true
}
