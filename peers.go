/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// peers.go defines how processes find and communicate with their peers.

package groupcache

import (
	"context"

	pb "github.com/udhos/groupcache/v2/groupcachepb"
)

// ProtoGetter is the interface that must be implemented by a peer.
type ProtoGetter interface {
	Get(context context.Context, in *pb.GetRequest, out *pb.GetResponse) error
	Remove(context context.Context, in *pb.GetRequest) error
	Set(context context.Context, in *pb.SetRequest) error
	// GetURL returns the peer URL
	GetURL() string
}

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	// PickPeer returns the peer that owns the specific key
	// and true to indicate that a remote peer was nominated.
	// It returns nil, false if the key owner is the current peer.
	PickPeer(key string) (peer ProtoGetter, ok bool)
	// GetAll returns all the peers in the group
	GetAll() []ProtoGetter
}

// NoPeers is an implementation of PeerPicker that never finds a peer.
type NoPeers struct{}

func (NoPeers) PickPeer(key string) (peer ProtoGetter, ok bool) { return }
func (NoPeers) GetAll() []ProtoGetter                           { return []ProtoGetter{} }

// RegisterPeerPickerWithWorkspace registers the peer initialization function.
// It is called once, when the first group is created.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPeerPickerWithWorkspace(ws *workspace, fn func() PeerPicker) {
	if ws.portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	ws.portPicker = func(_ string) PeerPicker { return fn() }
}

// RegisterPeerPicker registers the peer initialization function.
// It is called once, when the first group is created.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPeerPicker(fn func() PeerPicker) {
	RegisterPeerPickerWithWorkspace(DefaultWorkspace, fn)
}

// RegisterPerGroupPeerPickerWithWorkspace registers the peer initialization function,
// which takes the groupName, to be used in choosing a PeerPicker.
// It is called once, when the first group is created.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPerGroupPeerPickerWithWorkspace(ws *workspace, fn func(groupName string) PeerPicker) {
	if ws.portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	ws.portPicker = fn
}

// RegisterPerGroupPeerPicker registers the peer initialization function,
// which takes the groupName, to be used in choosing a PeerPicker.
// It is called once, when the first group is created.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPerGroupPeerPicker(fn func(groupName string) PeerPicker) {
	RegisterPerGroupPeerPickerWithWorkspace(DefaultWorkspace, fn)
}

func getPeers(ws *workspace, groupName string) PeerPicker {
	if ws.portPicker == nil {
		return NoPeers{}
	}
	pk := ws.portPicker(groupName)
	if pk == nil {
		pk = NoPeers{}
	}
	return pk
}
