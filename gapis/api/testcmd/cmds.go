// Copyright (C) 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testcmd provides fake commands used for testing.
package testcmd

import (
	"context"
	"reflect"

	"github.com/google/gapid/core/data"
	"github.com/google/gapid/core/data/protoconv"
	"github.com/google/gapid/core/image"
	"github.com/google/gapid/core/os/device"
	"github.com/google/gapid/gapil/constset"
	"github.com/google/gapid/gapis/api"
	"github.com/google/gapid/gapis/api/testcmd/test_pb"
	"github.com/google/gapid/gapis/memory"
	"github.com/google/gapid/gapis/replay/builder"
	"github.com/google/gapid/gapis/service/box"
)

type A struct {
	ID    api.CmdID
	Flags api.CmdFlags
}

func (a *A) Thread() uint64         { return 1 }
func (a *A) SetThread(uint64)       {}
func (a *A) CmdName() string        { return "A" }
func (a *A) API() api.API           { return nil }
func (a *A) CmdFlags() api.CmdFlags { return a.Flags }
func (a *A) Extras() *api.CmdExtras { return nil }
func (a *A) Mutate(context.Context, *api.State, *builder.Builder) error {
	return nil
}

type B struct {
	ID   api.CmdID
	Bool bool
}

func (a *B) Thread() uint64         { return 1 }
func (a *B) SetThread(uint64)       {}
func (a *B) CmdName() string        { return "B" }
func (a *B) API() api.API           { return nil }
func (a *B) CmdFlags() api.CmdFlags { return 0 }
func (a *B) Extras() *api.CmdExtras { return nil }
func (a *B) Mutate(context.Context, *api.State, *builder.Builder) error {
	return nil
}

type (
	Pointer struct {
		addr uint64
		pool memory.PoolID
	}

	StringːString map[string]string

	IntːStructPtr map[int]*Struct

	Struct struct {
		Str string
		Ref *Struct
	}
)

// Interface compliance checks
var _ memory.Pointer = &Pointer{}

func (p Pointer) String() string                            { return memory.PointerToString(p) }
func (p Pointer) IsNullptr() bool                           { return p.addr == 0 && p.pool == memory.ApplicationPool }
func (p Pointer) Address() uint64                           { return p.addr }
func (p Pointer) Pool() memory.PoolID                       { return p.pool }
func (p Pointer) Offset(n uint64) memory.Pointer            { panic("not implemented") }
func (p Pointer) ElementSize(m *device.MemoryLayout) uint64 { return 1 }
func (p Pointer) ElementType() reflect.Type                 { return reflect.TypeOf(byte(0)) }
func (p Pointer) ISlice(start, end uint64, m *device.MemoryLayout) memory.Slice {
	panic("not implemented")
}

var _ data.Assignable = &Pointer{}

func (p *Pointer) Assign(o interface{}) bool {
	if o, ok := o.(memory.Pointer); ok {
		*p = Pointer{o.Address(), o.Pool()}
		return true
	}
	return false
}

type X struct {
	Str  string        `param:"Str"`
	Sli  []bool        `param:"Sli"`
	Ref  *Struct       `param:"Ref"`
	Ptr  Pointer       `param:"Ptr"`
	Map  StringːString `param:"Map"`
	PMap IntːStructPtr `param:"PMap"`
}

func (X) Thread() uint64         { return 1 }
func (X) SetThread(uint64)       {}
func (X) CmdName() string        { return "X" }
func (X) API() api.API           { return api.Find(APIID) }
func (X) CmdFlags() api.CmdFlags { return 0 }
func (X) Extras() *api.CmdExtras { return nil }
func (X) Mutate(context.Context, *api.State, *builder.Builder) error {
	return nil
}

type API struct{}

func (API) Name() string                 { return "foo" }
func (API) ID() api.ID                   { return APIID }
func (API) Index() uint8                 { return 15 }
func (API) ConstantSets() *constset.Pack { return nil }
func (API) GetFramebufferAttachmentInfo(*api.State, uint64, api.FramebufferAttachment) (uint32, uint32, uint32, *image.Format, error) {
	return 0, 0, 0, nil, nil
}
func (API) Context(*api.State, uint64) api.Context { return nil }
func (API) CreateCmd(name string) api.Cmd {
	switch name {
	case "X":
		return &X{}
	default:
		return nil
	}
}

var (
	APIID = api.ID{1, 2, 3}

	P = &X{
		Str:  "aaa",
		Sli:  []bool{true, false, true},
		Ref:  &Struct{Str: "ccc", Ref: &Struct{Str: "ddd"}},
		Ptr:  Pointer{0x123, 0x456},
		Map:  StringːString{"cat": "meow", "dog": "woof"},
		PMap: IntːStructPtr{},
	}

	Q = &X{
		Str: "xyz",
		Sli: []bool{false, true, false},
		Ptr: Pointer{0x321, 0x654},
		Map: StringːString{"bird": "tweet", "fox": "?"},
		PMap: IntːStructPtr{
			100: &Struct{Str: "baldrick"},
		},
	}
)

func init() {
	api.Register(API{})
	protoconv.Register(func(ctx context.Context, a *X) (*test_pb.X, error) {
		return &test_pb.X{Data: box.NewValue(*a)}, nil
	}, func(ctx context.Context, b *test_pb.X) (*X, error) {
		var a X
		if err := b.Data.AssignTo(&a); err != nil {
			return nil, err
		}
		return &a, nil
	})
}
