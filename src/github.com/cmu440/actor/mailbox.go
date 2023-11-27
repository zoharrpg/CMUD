package actor

import (
	"sync"
)

// A mailbox, i.e., a thread-safe unbounded FIFO queue.
//
// You can think of Mailbox like a Go channel with an infinite buffer.
//
// Mailbox is only exported outside of the actor package for use in tests;
// we do not expect you to use it, just implement it.
type Mailbox struct {
	mu      sync.Mutex
	message []any
	closed  bool
	cond    *sync.Cond

	// TODO (3A): implement this!
}

// Returns a new mailbox that is ready for use.
func NewMailbox() *Mailbox {
	mailbox := &Mailbox{}
	mailbox.cond = sync.NewCond(&mailbox.mu)
	// TODO (3A): implement this!
	return mailbox
}

// Pushes message onto the end of the mailbox's FIFO queue.
//
// This function should NOT block.
//
// If mailbox.Close() has already been called, this may ignore
// the message. It still should NOT block.
//
// Note: message is not a literal actor message; it is an ActorSystem
// wrapper around a marshalled actor message.
func (mailbox *Mailbox) Push(message any) {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	if !mailbox.closed {
		mailbox.message = append(mailbox.message, message)
		mailbox.cond.Signal()
	}
	// TODO (3A): implement this!
}

// Pops a message from the front of the mailbox's FIFO queue,
// blocking until a message is available.
//
// If mailbox.Close() is called (either before or during a Pop() call),
// this should unblock and return (nil, false). Otherwise, it should return
// (message, true).
func (mailbox *Mailbox) Pop() (message any, ok bool) {
	// TODO (3A): implement this!
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()

	for len(mailbox.message) == 0 && !mailbox.closed {
		mailbox.cond.Wait()
	}
	if len(mailbox.message) > 0 && !mailbox.closed {
		message = mailbox.message[0]
		mailbox.message = mailbox.message[1:]
		ok = true
		return message, ok
	}

	return nil, false
}

// Closes the mailbox, causing future Pop() calls to return (nil, false)
// and terminating any goroutines running in the background.
//
// If Close() has already been called, this may exhibit undefined behavior,
// including blocking indefinitely.
func (mailbox *Mailbox) Close() {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()

	if !mailbox.closed {
		mailbox.closed = true
		mailbox.cond.Broadcast()
	}
	// TODO (3A): implement this!
}
