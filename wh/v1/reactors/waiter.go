package reactors

// Waiter allows something to say when it's done so other things can wait for
// it. It differs from a WaitGroup in that the owner can call Done many times,
// and the waiting client will only know when Done was done at least once.
type Waiter struct {
	c    chan struct{}
	done bool
}

// Create a new waiter.
func NewWaiter() *Waiter {
	return &Waiter{
		c: make(chan struct{}),
	}
}

// Set the waiter as done.
func (w *Waiter) Done() {
	if !w.done {
		w.done = true
		close(w.c)
	}
}

// Returns a channel that closes when done.
func (w *Waiter) Wait() <-chan struct{} {
	return w.c
}
