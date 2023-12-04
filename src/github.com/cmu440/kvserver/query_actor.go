package kvserver

import (
	"encoding/gob"
	"fmt"
	"github.com/cmu440/actor"
	"strings"
)

// Implement your queryActor in this file.
// See example/counter_actor.go for an example actor using the
// github.com/cmu440/actor package.

// TODO (3B): define your message types as structs

func init() {
	// TODO (3B): Register message types, e.g.:
	gob.Register(MGet{})
	gob.Register(MPut{})
	gob.Register(MList{})
	gob.Register(ListResult{})
	gob.Register(Result{})
	gob.Register(PutResult{})
}

type queryActor struct {
	// TODO (3B): implement this!
	context *actor.ActorContext
	store   map[string]string
}
type MGet struct {
	Key    string
	Sender *actor.ActorRef
}
type MPut struct {
	Key    string
	Value  string
	Sender *actor.ActorRef
}
type MList struct {
	Prefix string
	Sender *actor.ActorRef
}
type ListResult struct {
	Pair map[string]string
}
type Result struct {
	Ok    bool
	Value string
}
type PutResult struct {
}

// "Constructor" for queryActors, used in ActorSystem.StartActor.
func newQueryActor(context *actor.ActorContext) actor.Actor {
	return &queryActor{
		// TODO (3B): implement this!
		context: context,
		store:   make(map[string]string),
	}
}

// OnMessage implements actor.Actor.OnMessage.
func (actor *queryActor) OnMessage(message any) error {
	// TODO (3B): implement this!
	switch m := message.(type) {
	case MGet:
		value, exist := actor.store[m.Key]
		result := Result{Value: value, Ok: exist}
		actor.context.Tell(m.Sender, result)
	case MPut:
		actor.store[m.Key] = m.Value
		result := PutResult{}
		actor.context.Tell(m.Sender, result)
	case MList:
		result := ListResult{Pair: make(map[string]string)}
		for k, v := range actor.store {
			if strings.HasPrefix(k, m.Prefix) {
				result.Pair[k] = v
			}
		}
		actor.context.Tell(m.Sender, result)
	default:
		return fmt.Errorf("Unexpected counterActor message type: %T", m)
	}
	return nil
}
