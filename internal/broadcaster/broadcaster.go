package broadcaster

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

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		pubChan:   make(chan T, 1),
		subChan:   make(chan chan T, 1),
		unsubChan: make(chan chan T, 1),
		stopChan:  make(chan struct{}),
	}
}

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

func (b *Broadcaster[T]) Stop() {
	close(b.stopChan)
}

func (b *Broadcaster[T]) Subscribe() chan T {
	c := make(chan T)
	// Tell the running Start method that a new subscriber channel was added
	b.subChan <- c
	return c
}

func (b *Broadcaster[T]) Unsubscribe(c chan T) {
	b.unsubChan <- c
}

func (b *Broadcaster[T]) Publish(msg T) {
	b.pubChan <- msg
}
