package actor

import (
	"net/rpc"
	"sync"
)

type RemoteTellArgs struct {
	Ref  *ActorRef
	Mars []byte
}

// remoteTellReply represents the reply for the remoteTell RPC.
type RemoteTellReply struct {
	// You can define fields here if needed.
}

var mailBox *Mailbox
var mux sync.Mutex

// Calls system.tellFromRemote(ref, mars) on the remote ActorSystem listening
// on ref.Address.
//
// This function should NOT wait for a reply from the remote system before
// returning, to allow sending multiple messages in a row more quickly.
// It should ensure that messages are delivered in-order to the remote system.
// (You may assume that remoteTell is not called multiple times
// concurrently with the same ref.Address).
func remoteTell(client *rpc.Client, ref *ActorRef, mars []byte) {
	// TODO (3B): implement this!

	args := &RemoteTellArgs{Ref: ref, Mars: mars}
	mailBox.Push(args)

	go func() {
		mux.Lock()
		defer mux.Unlock()
		next, ok := mailBox.Pop()
		if !ok {

			return
		}
		client.Go("RemoteTellHandler.RemoteTell", next, nil, nil)
	}()
	//if err != nil {
	//	fmt.Printf("Error calling RemoteTell RPC: %v\n", err)
	//	// Handle the error as needed.
	//}
}

// Registers an RPC handler on server for remoteTell calls to system.
//
// You do not need to start the server's listening on the network;
// just register a handler struct that handles remoteTell RPCs by calling
// system.tellFromRemote(ref, mars).
func registerRemoteTells(system *ActorSystem, server *rpc.Server) error {
	// TODO (3B): implement this!
	mailBox = NewMailbox()
	handler := &RemoteTellHandler{
		ActorSys: system,
	}

	err := server.RegisterName("RemoteTellHandler", handler)
	if err != nil {
		return err
	}

	return nil
}

type RemoteTellHandler struct {
	ActorSys *ActorSystem
}

// TODO (3B): implement your remoteTell RPC handler below!

// RemoteTell handles the remoteTell RPC.
func (h *RemoteTellHandler) RemoteTell(args *RemoteTellArgs, reply *RemoteTellReply) error {
	// Call system.tellFromRemote(ref, mars) using the provided arguments.
	h.ActorSys.tellFromRemote(args.Ref, args.Mars)

	return nil
}
