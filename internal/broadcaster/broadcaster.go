// This package provides the [Broadcaster] type which is used to
// publish a message to many subscribers.
package broadcaster

// A [Broadcaster] is used to send messages to all subscribed
// listeners.
type Broadcaster[T any] struct {
	// Channel used for publishing messages to.
	pubChan chan T
	// Channel used for registering new subscription channels.
	subChan chan chan T
	// Channel used for unregistering subscription channels.
	unsubChan chan chan T
	// Channel used to stop the broadcaster.
	stopChan chan struct{}
}

// Creates a new [Broadcaster] instance. T is the type to be
// broadcasted.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		pubChan:   make(chan T, 1),
		subChan:   make(chan chan T, 1),
		unsubChan: make(chan chan T, 1),
		stopChan:  make(chan struct{}),
	}
}

// Start the [Broadcaster], listening for subscribers and processing
// messages.
func (b *Broadcaster[T]) Start() {
	subscribers := map[chan T]struct{}{}
	for {
		select {
		// Stop the broadcaster
		case <-b.stopChan:
			return
		// Register a new subscriber channel
		case newChan := <-b.subChan:
			subscribers[newChan] = struct{}{}
		// Unregister an existing subscriber channel
		case chanToDelete := <-b.unsubChan:
			delete(subscribers, chanToDelete)
		case msg := <-b.pubChan:
			for sub := range subscribers {
				// Use a non-blocking send to protect the broadcaster from blocking
				select {
				case sub <- msg:
				default:
				}
			}
		}
	}
}

// Stop the [Broadcaster].
func (b *Broadcaster[T]) Stop() {
	close(b.stopChan)
}

// Add a new subscriber to this [Broadcaster].
// Returns a channel that receives all messages sent to the 
// [Broadcaster.]
func (b *Broadcaster[T]) Subscribe() chan T {
	c := make(chan T)
	// Tell the running Start method that a new subscriber channel was added
	b.subChan <- c
	return c
}

// Unsubscribe the given channel from this [Broadcaster].
func (b *Broadcaster[T]) Unsubscribe(c chan T) {
	b.unsubChan <- c
}

// Publish a message to all subscribed listeners.
func (b *Broadcaster[T]) Publish(msg T) {
	b.pubChan <- msg
}
