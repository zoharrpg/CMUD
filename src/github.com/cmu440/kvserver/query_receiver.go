package kvserver

import (
	"github.com/cmu440/actor"
	"github.com/cmu440/kvcommon"
)

// RPC handler implementing the kvcommon.QueryReceiver interface.
// There is one queryReceiver per queryActor, each running on its own port,
// created and registered for RPCs in NewServer.

// queryReceiver
// MUST answer RPCs by sending a message to its query actor and getting a response message from that query actor (via
// ActorSystem's NewChannelRef). It must NOT attempt to answer queries using its own state,
// and it must NOT directly coordinate with other queryReceivers - all coordination is done within the actor system
// by its query actor.
type queryReceiver struct {
	ActorSystem *actor.ActorSystem
	ActorRef    *actor.ActorRef
}

// Get implements kvcommon.QueryReceiver.Get.
func (rcvr *queryReceiver) Get(args kvcommon.GetArgs, reply *kvcommon.GetReply) error {
	ref, channel := rcvr.ActorSystem.NewChannelRef()
	rcvr.ActorSystem.Tell(rcvr.ActorRef, MGet{Key: args.Key, Sender: ref})
	tmp := <-channel
	reply.Value = tmp.(GetResult).Value
	reply.Ok = tmp.(GetResult).Ok
	return nil
}

// List implements kvcommon.QueryReceiver.List.
func (rcvr *queryReceiver) List(args kvcommon.ListArgs, reply *kvcommon.ListReply) error {
	ref, channel := rcvr.ActorSystem.NewChannelRef()

	rcvr.ActorSystem.Tell(rcvr.ActorRef, MList{Prefix: args.Prefix, Sender: ref})
	tmp := <-channel
	reply.Entries = tmp.(ListResult).Pair
	return nil
}

// Put implements kvcommon.QueryReceiver.Put.
func (rcvr *queryReceiver) Put(args kvcommon.PutArgs, reply *kvcommon.PutReply) error {
	ref, _ := rcvr.ActorSystem.NewChannelRef()
	//currentTime := time.Now().UnixMilli()

	rcvr.ActorSystem.Tell(rcvr.ActorRef, MPut{Key: args.Key, Value: args.Value, Sender: ref})
	return nil
}
