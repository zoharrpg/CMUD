package kvserver

import (
	"encoding/gob"
	"fmt"
	"github.com/cmu440/actor"
	"strings"
	"time"
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
	gob.Register(GetResult{})
	gob.Register(PutResult{})
	gob.Register(InitLocal{})
	gob.Register(LocalSynSignal{})
	gob.Register(LocalSync{})
	gob.Register(NotifyNewServer{})
	gob.Register(RemoteSyncSignal{})
}

type queryActor struct {
	// TODO (3B): implement this!
	context     *actor.ActorContext
	store       map[string]StoreValue
	ActorsInfo  []*actor.ActorRef
	RemoteInfo  [][]*actor.ActorRef
	Me          int
	Logs        map[string]MPut
	ActorSystem *actor.ActorSystem
}
type StoreValue struct {
	Value     string
	Timestamp int64 //millisecond resolution
	Uid       string
}
type MGet struct {
	Key    string
	Sender *actor.ActorRef
}
type MPut struct {
	Key       string
	Value     string
	Sender    *actor.ActorRef
	Timestamp int64
}
type MList struct {
	Prefix string
	Sender *actor.ActorRef
}
type ListResult struct {
	Pair map[string]string
}
type GetResult struct {
	Ok    bool
	Value string
}
type PutResult struct {
}
type InitLocal struct {
	ActorsInfo []*actor.ActorRef
	Me         int
	RemoteInfo [][]*actor.ActorRef
}
type LocalSynSignal struct {
}
type LocalSync struct {
	Data map[string]MPut
}
type NotifyNewServer struct {
	Refs []*actor.ActorRef
}
type RemoteSyncSignal struct {
}

// "Constructor" for queryActors, used in ActorSystem.StartActor.
func newQueryActor(context *actor.ActorContext) actor.Actor {

	return &queryActor{
		// TODO (3B): implement this!
		context:    context,
		store:      make(map[string]StoreValue),
		ActorsInfo: make([]*actor.ActorRef, 0),
		Me:         -1,
		Logs:       make(map[string]MPut),
		RemoteInfo: make([][]*actor.ActorRef, 0),
	}
}

// OnMessage implements actor.Actor.OnMessage.
func (actor *queryActor) OnMessage(message any) error {
	// TODO (3B): implement this!
	switch m := message.(type) {
	case NotifyNewServer:
		actor.RemoteInfo = append(actor.RemoteInfo, m.Refs)
		actor.context.Tell(actor.ActorsInfo[actor.Me], RemoteSyncSignal{})

	case RemoteSyncSignal:
		for _, remote := range actor.RemoteInfo {
			for _, server := range remote {
				actor.context.Tell(server, LocalSync{Data: actor.Logs})
			}
		}

	case LocalSynSignal:
		//fmt.Println("LocSynSignal Received")
		for index, a := range actor.ActorsInfo {
			if index != actor.Me {
				actor.context.Tell(a, LocalSync{Data: actor.Logs})

			}
		}
		for _, remote := range actor.RemoteInfo {
			for _, server := range remote {
				actor.context.Tell(server, LocalSync{Data: actor.Logs})
			}
		}

		actor.Logs = make(map[string]MPut)
		actor.context.TellAfter(actor.ActorsInfo[actor.Me], LocalSynSignal{}, 100*time.Millisecond)
	case LocalSync:
		//fmt.Println("LocSyn Received")
		for _, data := range m.Data {
			doPut := false

			if v, ok := actor.store[data.Key]; ok {
				if data.Timestamp > v.Timestamp {
					doPut = true
				} else if data.Timestamp == v.Timestamp && data.Sender.Uid() < v.Uid {
					doPut = true
				}
			} else {
				doPut = true
			}

			if doPut {
				actor.store[data.Key] = StoreValue{data.Value, data.Timestamp, data.Sender.Uid()}

			}

		}

	case InitLocal:
		actor.ActorsInfo = append(actor.ActorsInfo, m.ActorsInfo...)
		actor.RemoteInfo = append(actor.RemoteInfo, m.RemoteInfo...)
		actor.Me = m.Me
		actor.context.Tell(actor.ActorsInfo[actor.Me], LocalSynSignal{})

	case MGet:
		v, exist := actor.store[m.Key]
		result := GetResult{Value: v.Value, Ok: exist}
		actor.context.Tell(m.Sender, result)
	case MPut:
		m.Timestamp = time.Now().UnixMilli()
		doPut := false
		if v, ok := actor.store[m.Key]; ok {
			if m.Timestamp > v.Timestamp {
				doPut = true
			} else if m.Timestamp == v.Timestamp && m.Sender.Uid() < v.Uid {
				doPut = true
			}
		} else {
			doPut = true
		}
		if doPut {
			actor.store[m.Key] = StoreValue{m.Value, m.Timestamp, m.Sender.Uid()}
			actor.Logs[m.Key] = m
		}
		result := PutResult{}
		actor.context.Tell(m.Sender, result)

	case MList:
		result := ListResult{Pair: make(map[string]string)}
		for k, v := range actor.store {
			if strings.HasPrefix(k, m.Prefix) {
				result.Pair[k] = v.Value
			}
		}
		actor.context.Tell(m.Sender, result)
	default:
		return fmt.Errorf("Unexpected counterActor message type: %T", m)
	}
	return nil
}
