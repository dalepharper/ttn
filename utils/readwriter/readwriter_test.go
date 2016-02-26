// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package readwriter

import (
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestReadWriter(t *testing.T) {
	{
		Desc(t, "Write to an empty buffer")
		rw := New(nil)
		rw.Write([]byte{1, 2, 3, 4})
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1, 2, 3, 4}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// -------------

	{
		Desc(t, "Write to a non empty buffer")
		rw := New([]byte{1, 2, 3, 4})
		rw.Write([]byte{1, 2})
		checkErrors(t, nil, rw.Err())
	}

	// -------------

	{
		Desc(t, "Append to an existing buffer")
		rw := New(nil)
		rw.Write([]byte{1, 2, 3, 4})
		data, _ := rw.Bytes()

		rw = New(data)
		rw.Write([]byte{5, 6})
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1, 2, 3, 4}, data) })
		rw.Read(func(data []byte) { checkData(t, []byte{5, 6}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// -------------

	{
		Desc(t, "Read from empty buffer")
		rw := New(nil)
		rw.Read(func(data []byte) { checkNotCalled(t) })
		checkErrors(t, pointer.String(string(errors.Behavioural)), rw.Err())
	}

	// --------------

	{
		Desc(t, "Write after read from non empty")
		rw := New(nil)
		rw.Write([]byte{1, 2})
		rw.Write([]byte{3, 4})
		data, _ := rw.Bytes()

		rw = New(data)
		rw.Read(func(data []byte) {})
		rw.Write([]byte{5, 6})
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{3, 4}, data) })
		rw.Read(func(data []byte) { checkData(t, []byte{5, 6}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write single byte")
		rw := New(nil)
		rw.Write(byte(1))
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write string")
		rw := New(nil)
		rw.Write("TheThingsNetwork")
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte("TheThingsNetwork"), data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write lorawan.AES128Key")
		rw := New(nil)
		rw.Write(lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}))
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write lorawan.EUI64")
		rw := New(nil)
		rw.Write(lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}))
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write lorawan.DevAddr")
		rw := New(nil)
		rw.Write(lorawan.DevAddr([4]byte{1, 2, 3, 4}))
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{1, 2, 3, 4}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write empty slice")
		rw := New(nil)
		rw.Write([]byte{})
		data, err := rw.Bytes()
		checkErrors(t, nil, err)

		rw = New(data)
		rw.Read(func(data []byte) { checkData(t, []byte{}, data) })
		checkErrors(t, nil, rw.Err())
	}

	// --------------

	{
		Desc(t, "Write invalid data")
		rw := New(nil)
		chwait := make(chan bool)
		go func() {
			defer func() {
				recover()
				close(chwait)
			}()
			rw.Write(14)
			checkNotCalled(t)
		}()
		<-chwait
	}
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == errors.Nature(*want) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkData(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check data")
		return
	}
	Ko(t, "Expected data to be %v but got %v", want, got)
}

func checkNotCalled(t *testing.T) {
	Ko(t, "Unexpected call on method")
}
