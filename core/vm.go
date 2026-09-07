package core

import "fmt"

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // 10
	InstrAdd      Instruction = 0x0b // 11
	InstrSub      Instruction = 0x0c // 12
	InstrPushByte Instruction = 0x0d // 13
	InstrPack     Instruction = 0x0e // 14

)

type Stack struct {
	data []any
	sp   int
	size int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
		size: size,
	}
}

func (s *Stack) Push(v any) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) Pop() any {
	if s.sp == 0 {
		return nil
	}

	s.sp--
	value := s.data[s.sp]
	s.data[s.sp] = nil

	return value
}

type VM struct {
	data  []byte
	ip    int // instruction pointer
	stack *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: NewStack(128),
	}
}

func (vm *VM) Run() error {
	for ; vm.ip < len(vm.data); vm.ip++ {
		instr := Instruction(vm.data[vm.ip])

		if err := vm.Exec(instr); err != nil {
			return err
		}
	}

	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := b - a
		vm.stack.Push(c)
	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))
	case InstrPack:
		size := vm.stack.Pop().(int)
		if size < 0 {
			return fmt.Errorf("pack: invalid size %d", size)
		}

		data := make([]byte, size)

		// LIFO: Fill from back to front, restoring the byte order prior to pushing onto the stack.
		for i := size - 1; i >= 0; i-- {
			b := vm.stack.Pop().(byte)
			data[i] = b
		}

		vm.stack.Push(data)
	}

	return nil
}
