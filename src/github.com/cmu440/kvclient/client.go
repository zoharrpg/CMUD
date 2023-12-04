// Package kvclient implements the client library for a
// geographically distributed, highly available, NoSQL key-value store.
package kvclient

import (
	"fmt"
	"github.com/cmu440/kvcommon"
	"net/rpc"
)

// QueryRouter
// Type for client.router.
//
// Returns the address of the RPC server to contact for the next query function call (Get, List, Put).
type QueryRouter interface {
	NextAddr() string
}

// Client
// A client for the key-value store.
//
// All client functions are thread-safe.
type Client struct {
	// Router used to simulate load balancing as described in the handout. You are given the router in NewClient.
	//
	// In each query function call (Get, List, Put), first call router.NextAddr() to get
	// the address of the RPC server to call.
	// It will give you the address of an RPC server that has an implementation of kvcommon.QueryReceiver registered
	// under the name  "QueryReceiver".
	//
	// Once you get the address, use rpc.Dial to get a rpc.Client.
	// For compatibility with our tests, use network "tcp".
	router QueryRouter
}

// NewClient
// Returns a client for connecting to the key-value store.
func NewClient(router QueryRouter) *Client {
	return &Client{
		router,
	}
}

// Send RPCs to type and name "QueryReceiver", defined in kvcommon/rpc_types.go.
// Your implementation should be thread-safe: there may be concurrent Get/Put/List calls, or multiple outstanding RPCs.

// Get
// Returns the value associated with key, if present.
// If key not present, ("", false, nil) is returned.
//
// If there is a network error contacting the RPC server indicated by router.NextAddr(), that error is returned instead.
func (client *Client) Get(key string) (value string, ok bool, err error) {
	addr := client.router.NextAddr()
	conn, err := rpc.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return "", false, nil
	}
	args := kvcommon.GetArgs{Key: key}
	reply := kvcommon.GetReply{}
	err = conn.Call("QueryReceiver.Get", args, &reply)

	if err != nil {
		fmt.Println(err)
		return "", false, nil

	}

	return reply.Value, reply.Ok, nil
}

// List
// Returns a map containing all (key, value) pairs whose key starts with prefix,
// similar to recursively listing all files in a folder.
//
// If there is a network error contacting the RPC server indicated by router.NextAddr(), that error is returned instead.
func (client *Client) List(prefix string) (entries map[string]string, err error) {
	addr := client.router.NextAddr()
	conn, err := rpc.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	args := kvcommon.ListArgs{Prefix: prefix}
	reply := kvcommon.ListReply{}
	err = conn.Call("QueryReceiver.List", args, &reply)

	if err != nil {
		fmt.Println(err)
		return nil, err

	}

	return reply.Entries, nil
}

// Put
// Sets the value associated with key.
//
// If there is a network error contacting the RPC server indicated by router.NextAddr(), that error is returned instead.
func (client *Client) Put(key string, value string) error {
	addr := client.router.NextAddr()
	conn, err := rpc.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	args := kvcommon.PutArgs{Key: key, Value: value}
	reply := kvcommon.PutReply{}
	err = conn.Call("QueryReceiver.Put", args, &reply)

	if err != nil {
		fmt.Println(err)
		return err

	}

	return nil
}

// Close
// OPTIONAL: Closes the client, including all of its RPC clients.
//
// You are not required to implement this function for full credit; the tests end by calling Close but do not check
// that it does anything. However, you may find it useful to implement this so that you can run multiple/repeated
// tests in the same "go test" command without cross-test interference.
func (client *Client) Close() {
}
