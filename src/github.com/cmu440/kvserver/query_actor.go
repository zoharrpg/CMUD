package kvserver

import (
	"encoding/gob"
	"fmt"
	"github.com/cmu440/actor"
	"strings"
	"time"
)

// Implement your queryActor in this file.
func init() {
	gob.Register(GetResult{})
	gob.Register(Init{})
	gob.Register(ListResult{})
	gob.Register(SynMsg{})
	gob.Register(SynSignal{})
	gob.Register(MGet{})
	gob.Register(MPut{})
	gob.Register(MList{})
	gob.Register(PutResult{})
	gob.Register(NotifyNewServer{})
}

// queryActor represents an actor that handles GET, PUT, and LIST requests.
type queryActor struct {
	ActorsInfo  []*actor.ActorRef
	ActorSystem *actor.ActorSystem
	Context     *actor.ActorContext
	Logs        map[string]MPut
	Me          int
	RemoteInfo  [][]*actor.ActorRef
	Store       map[string]StoreValue
}

// StoreValue is the value stored in the store
type StoreValue struct {
	Sender    *actor.ActorRef
	Timestamp int64 //millisecond resolution
	Value     string
}

// MGet is the message type for GET requests.
type MGet struct {
	Key    string
	Sender *actor.ActorRef
}

// MPut is the message type for PUT requests.
type MPut struct {
	Key       string
	Sender    *actor.ActorRef
	Timestamp int64
	Value     string
}

// MList is the message type for LIST requests.
type MList struct {
	Prefix string
	Sender *actor.ActorRef
}

// ListResult is the message type for LIST responses.
type ListResult struct {
	Pair map[string]string
}

// GetResult is the message type for GET responses.
type GetResult struct {
	Ok    bool
	Value string
}

// PutResult is the message type for PUT responses.
type PutResult struct {
}

// Init is the message type for initializing the queryActor
type Init struct {
	ActorsInfo []*actor.ActorRef
	RemoteInfo [][]*actor.ActorRef
	Me         int
}

// SynSignal is the message type for triggering synchronization
type SynSignal struct {
}

// SynMsg is the message type for synchronization
type SynMsg struct {
	Data map[string]MPut
}

// NotifyNewServer is the message type for notifying that a new server came online
type NotifyNewServer struct {
	Refs []*actor.ActorRef
}

// "Constructor" for queryActors, used in ActorSystem.StartActor.
func newQueryActor(context *actor.ActorContext) actor.Actor {
	return &queryActor{
		ActorsInfo: make([]*actor.ActorRef, 0),
		Context:    context,
		Logs:       make(map[string]MPut),
		Me:         -1,
		RemoteInfo: make([][]*actor.ActorRef, 0),
		Store:      make(map[string]StoreValue),
	}
}

// OnMessage implements actor.Actor.OnMessage.
// Sync Strategy:
//  1. When a new server joins, it will send a NotifyNewServer message to all servers.
//  2. When a server receives a NotifyNewServer message, it will send a SynMsg message to the new server.
//  3. When a server receives a SynMsg message, it will update its own store and logs.
//  4. Every 100ms, a server will send a SynSignal message to itself
//  5. When a server receives a SynSignal message, it will send a SynMsg message to all servers.
func (actor *queryActor) OnMessage(message any) error {
	switch m := message.(type) {
	case NotifyNewServer:
		actor.RemoteInfo = append(actor.RemoteInfo, m.Refs)
		logs := make(map[string]MPut)

		for k, v := range actor.Store {
			logs[k] = MPut{Key: k, Value: v.Value, Sender: v.Sender, Timestamp: v.Timestamp}
		}

		for _, ref := range m.Refs {
			actor.Context.Tell(ref, SynMsg{Data: logs})
		}

	case SynSignal:
		for index, a := range actor.ActorsInfo {
			if index != actor.Me {
				actor.Context.Tell(a, SynMsg{Data: actor.Logs})
			}
		}
		for _, remote := range actor.RemoteInfo {
			actor.Context.Tell(remote[0], SynMsg{Data: actor.Logs})
		}
		actor.Logs = make(map[string]MPut)
		actor.Context.TellAfter(actor.ActorsInfo[actor.Me], SynSignal{}, 100*time.Millisecond)

	case SynMsg:
		for _, data := range m.Data {
			doPut := false

			if v, ok := actor.Store[data.Key]; ok {
				if data.Timestamp > v.Timestamp {
					doPut = true
				} else if data.Timestamp == v.Timestamp && data.Sender.Uid() < v.Sender.Uid() {
					doPut = true
				}
			} else {
				doPut = true
			}

			if doPut {
				actor.Store[data.Key] = StoreValue{data.Sender, data.Timestamp, data.Value}
				actor.Logs[data.Key] = data
			}
		}

	case Init:
		actor.ActorsInfo = append(actor.ActorsInfo, m.ActorsInfo...)
		actor.RemoteInfo = append(actor.RemoteInfo, m.RemoteInfo...)
		actor.Me = m.Me
		actor.Context.Tell(actor.ActorsInfo[actor.Me], SynSignal{})

	case MGet:
		v, exist := actor.Store[m.Key]
		result := GetResult{Value: v.Value, Ok: exist}
		actor.Context.Tell(m.Sender, result)

	case MPut:
		m.Timestamp = time.Now().UnixMilli()
		doPut := false
		if v, ok := actor.Store[m.Key]; ok {
			if m.Timestamp > v.Timestamp {
				doPut = true
			} else if m.Timestamp == v.Timestamp && m.Sender.Uid() < v.Sender.Uid() {
				doPut = true
			}
		} else {
			doPut = true
		}
		if doPut {
			actor.Store[m.Key] = StoreValue{m.Sender, m.Timestamp, m.Value}
			actor.Logs[m.Key] = m
		}
		result := PutResult{}
		actor.Context.Tell(m.Sender, result)

	case MList:
		result := ListResult{Pair: make(map[string]string)}
		for k, v := range actor.Store {
			if strings.HasPrefix(k, m.Prefix) {
				result.Pair[k] = v.Value
			}
		}
		actor.Context.Tell(m.Sender, result)

	default:
		return fmt.Errorf("Unexpected counterActor message type: %T", m)
	}
	return nil
}
