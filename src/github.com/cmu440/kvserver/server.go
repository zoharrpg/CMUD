// Package kvserver implements the backend server for a
// geographically distributed, highly available, NoSQL key-value store.
package kvserver

import (
	"encoding/json"
	"fmt"
	"github.com/cmu440/actor"
	"net"
	"net/rpc"
	"strconv"
)

// Server
// A single server in the key-value store, running some number of query actors - nominally one per CPU core.
// Each query actor provides a key/value storage service on its own port.
//
// Different query actors (both within this server and across connected servers) periodically sync updates (Puts)
// following an eventually consistent, last-writer-wins strategy.
type Server struct {
	AS        *actor.ActorSystem
	ActorInfo []*actor.ActorRef
}

// OPTIONAL: Error handler for ActorSystem.OnError.
//
// Print the error or call debug.PrintStack() in this function.
// When starting an ActorSystem, call ActorSystem.OnError(errorHandler).
// This can help debug server-side errors more easily.
func errorHandler(err error) {
}

// NewServer Starts a server running queryActorCount query actors.
//
// The server's actor system listens for remote messages (from other actor systems) on startPort.
// The server listens for RPCs from kvclient.Clients on ports [startPort + 1, startPort + 2, ..., startPort + queryActorCount].
// Each of these "query RPC servers" answers queries by asking a specific query actor.
//
// remoteDescs contains a "description" string for each existing server in the key-value store.
// Specifically, each slice entry is the desc returned by an existing server's own NewServer call.
// The description strings are opaque to callers, but typically an implementation uses JSON-encoded data containing,
// e.g., actor.ActorRef's that remote servers' actors should contact.
//
// Before returning, NewServer starts the ActorSystem, all query actors, and all query RPC servers.
// If there is an error starting anything, that error is returned instead.
func NewServer(startPort int, queryActorCount int, remoteDescs []string) (server *Server, desc string, err error) {
	// Tips:
	// - The "HTTP service" example in the net/rpc docs does not support multiple RPC servers in the same process.
	// Instead, use the following template to start RPC servers (adapted from
	// https://groups.google.com/g/Golang-Nuts/c/JTn3LV_bd5M/m/cMO_DLyHPeUJ ):
	actorsInfo := make([]*actor.ActorRef, 0)
	actorSystem, _ := actor.NewActorSystem(startPort)

	for i := 1; i < queryActorCount+1; i++ {
		rpcServer := rpc.NewServer()
		q := &queryReceiver{}
		err := rpcServer.RegisterName("QueryReceiver", q)
		if err != nil {
			return nil, "", err
		}
		ln, _ := net.Listen("tcp", ":"+strconv.Itoa(startPort+i))
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go rpcServer.ServeConn(conn)
			}
		}()
		q.ActorSystem = actorSystem
		rf := actorSystem.StartActor(newQueryActor)
		q.ActorRef = rf
		actorsInfo = append(actorsInfo, rf)
	}

	// - To start query actors, call your ActorSystem's StartActor(newQueryActor), where newQueryActor is defined in ./query_actor.go.
	// Do this queryActorCount times. (For the checkpoint tests, queryActorCount will always be 1.)
	// - remoteDescs and desc: see doc comment above. For the checkpoint, it is okay to ignore remoteDescs and return "" for desc.
	RemoteServers := make([][]*actor.ActorRef, 0)
	for _, dec := range remoteDescs {
		var server []*actor.ActorRef
		err := json.Unmarshal([]byte(dec), &server)
		if err != nil {
			return nil, "", err

		}
		RemoteServers = append(RemoteServers, server)
	}

	for index, ref := range actorsInfo {
		actorSystem.Tell(ref, Init{ActorsInfo: actorsInfo, Me: index, RemoteInfo: RemoteServers})
	}
	for _, oldServer := range RemoteServers {
		for _, oldActor := range oldServer {
			actorSystem.Tell(oldActor, NotifyNewServer{actorsInfo})
		}
	}

	s := Server{AS: actorSystem, ActorInfo: actorsInfo}

	jsonData, err := json.Marshal(s.ActorInfo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, "", err
	}
	desc = string(jsonData)

	return &s, desc, nil
}

// Close OPTIONAL: Closes the server, including its actor system nd all RPC servers.
//
// You are not required to implement this function for full credit; the tests end by calling Close but do not check
// that it does anything. However, you may find it useful to implement this so that you can run multiple/repeated
// tests in the same "go test" command without cross-test interference (in particular, old test servers' squatting on ports.)
//
// Likewise, you may find it useful to close a partially-started server's resources if there is an error in NewServer.
func (server *Server) Close() {
}
