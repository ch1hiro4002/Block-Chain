package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/util"
	"github.com/stretchr/testify/assert"
)

func TestStack_Push_Pop(t *testing.T) {
	stack := NewStack(1024)

	stack.Push(1)
	stack.Push(2)

	value := stack.Pop()
	assert.Equal(t, value.(int), 2)
}

func TestVM_Int_Add(t *testing.T) {
	// 1 + 2 = 3
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0c}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(3), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Int_Sub(t *testing.T) {
	// 3 - 2 = 1

	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0d}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(1), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Byte_Pack(t *testing.T) {
	// stack: [byte, byte, byte, size]
	// "FOO" ==> []byte{0x46, 0x4f, 0x4f}

	data := []byte{0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x03, 0x0a, 0x02, 0x0a, 0x0d}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	result1 := vm.stack.Pop().(int)
	assert.Equal(t, 1, result1)
	result2 := vm.stack.Pop().([]byte)
	assert.Equal(t, "FOO", string(result2))
}

func TestVM_State(t *testing.T) {
	// State: map[string][]byte
	// "FOO" => []byte{3 + 2}
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}

	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	value, err := vm.state.Get([]byte("FOO"))
	assert.Nil(t, err)
	assert.Equal(t, value, util.SerializeInt64(int64(5)))
}

func TestVM_Get(t *testing.T) {
	fooData := []byte{0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x12}
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}
	data = append(data, fooData...)

	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	value := vm.stack.Pop().([]byte)
	assert.Equal(t, value, util.SerializeInt64(int64(5)))
}

func TestVM_Int_Mul(t *testing.T) {
	// 3 * 4 = 12
	data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0e}

	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(12), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Int_Div(t *testing.T) {
	// 8 / 2 = 4
	data := []byte{0x08, 0x0a, 0x02, 0x0a, 0x0f}

	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(4), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Int_Div_ByZero(t *testing.T) {
	data := []byte{0x08, 0x0a, 0x00, 0x0a, 0x0f}

	state := NewState()
	vm := NewVM(data, state)

	assert.Error(t, vm.Run())
}
