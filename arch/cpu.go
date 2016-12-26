package arch

// Inst is an interface for executing one single instruction
type inst interface {
	I(cpu *cpu, in uint32) *Excep
}

// CPU defines the structure of a processing unit.
type cpu struct {
	regs []uint32
	ring byte

	phyMem    *phyMemory
	virtMem   *virtMemory
	interrupt *interrupt
	calls     *calls

	inst     inst
	index    byte
	ncycle   uint64
	sleeping bool
}

// newCPU creates a CPU with memroy and instruction binding
func newCPU(mem *phyMemory, calls *calls, i inst, index byte) *cpu {
	if index >= 32 {
		panic("too many cores")
	}

	ret := new(cpu)
	ret.regs = makeRegs()
	ret.phyMem = mem
	ret.virtMem = newVirtMemory(ret.phyMem)
	ret.calls = calls
	ret.index = index

	intPage := ret.phyMem.Page(pageInterrupt) // page 1 is the interrupt page
	if intPage == nil {
		panic("memory too small")
	}
	ret.interrupt = newInterrupt(intPage, index)
	ret.inst = i

	ret.regs[PC] = InitPC

	return ret
}

// UserMode returns true when the CPU is in user mode.
func (c *cpu) UserMode() bool {
	return c.ring > 0
}

// Reset resets the CPU's internal states, i.e., registers,
// the page table, and disables interrupt
func (c *cpu) Reset() {
	for i := 0; i < Nreg; i++ {
		c.regs[i] = 0
	}
	c.regs[PC] = InitPC
	c.virtMem.SetTable(0)
	c.ring = 0
	c.interrupt.Disable()
}

func (c *cpu) tick() *Excep {
	c.ncycle++
	pc := c.regs[PC]
	inst, e := c.readWord(pc)
	if e != nil {
		return e
	}

	c.regs[PC] = pc + 4
	if c.inst != nil {
		e = c.inst.I(c, inst)

		if e != nil {
			c.regs[PC] = pc // restore saved original PC
			return e
		}
	}

	return nil
}

const (
	intFrameSP   = 0
	intFrameRET  = 4
	intFrameArg  = 8
	intFrameCode = 12
	intFrameRing = 13

	intFrameSize = 16
)

// Interrupt issues an interrupt to the core
func (c *cpu) Interrupt(code byte) {
	c.interrupt.Issue(code)
}

func (c *cpu) readWord(addr uint32) (uint32, *Excep) {
	return c.virtMem.ReadWord(addr, c.ring)
}

func (c *cpu) readByte(addr uint32) (uint8, *Excep) {
	return c.virtMem.ReadByte(addr, c.ring)
}

func (c *cpu) writeWord(addr uint32, v uint32) *Excep {
	return c.virtMem.WriteWord(addr, c.ring, v)
}

func (c *cpu) writeByte(addr uint32, v uint8) *Excep {
	return c.virtMem.WriteByte(addr, c.ring, v)
}

// Ienter enters a interrupt routine.
func (c *cpu) Ienter(code byte, arg uint32) *Excep {
	hsp := c.interrupt.handlerSP()
	base := hsp - intFrameSize

	writeWord := func(off uint32, v uint32) *Excep {
		return c.virtMem.WriteWord(base+off, 0, v)
	}
	writeByte := func(off uint32, b uint8) *Excep {
		return c.virtMem.WriteByte(base+off, 0, b)
	}
	if e := writeWord(intFrameSP, c.regs[SP]); e != nil {
		return e
	}
	if e := writeWord(intFrameRET, c.regs[RET]); e != nil {
		return e
	}
	if e := writeWord(intFrameArg, arg); e != nil {
		return e
	}
	if e := writeByte(intFrameCode, code); e != nil {
		return e
	}
	if e := writeByte(intFrameRing, c.ring); e != nil {
		return e
	}

	c.interrupt.Disable()
	c.regs[SP] = hsp
	c.regs[RET] = c.regs[PC]
	c.regs[PC] = c.interrupt.handlerPC()
	c.ring = 0

	return nil
}

// Syscall jumps to the system call handler and switches to ring 0.
func (c *cpu) Syscall() *Excep {
	userSP := c.regs[SP]
	syscallSP := c.interrupt.syscallSP()

	if e := c.virtMem.WriteWord(syscallSP-4, 0, userSP); e != nil {
		return e
	}

	c.regs[SP] = syscallSP
	c.regs[RET] = c.regs[PC]
	c.regs[PC] = c.interrupt.syscallPC()
	c.ring = 0

	return nil
}

// Iret restores from an interrupt.
// It restores the SP, RET, PC registers, restores the ring level,
// clears the served interrupt bit and enables interrupt again.
// The interrupt trap frame is saved on the current stack.
func (c *cpu) Iret() *Excep {
	if c.ring != 0 {
		panic("iret in userland")
	}

	sp := c.regs[SP]
	base := sp - intFrameSize
	sp, e := c.readWord(base + intFrameSP)
	if e != nil {
		return e
	}
	ret, e := c.readWord(base + intFrameRET)
	if e != nil {
		return e
	}
	code, e := c.readByte(base + intFrameCode)
	if e != nil {
		return e
	}
	ring, e := c.readByte(base + intFrameRing)
	if e != nil {
		return e
	}

	c.regs[PC] = c.regs[RET]
	c.regs[RET] = ret
	c.regs[SP] = sp
	c.ring = ring
	if code > 0 {
		c.interrupt.Clear(code)
	}
	c.interrupt.Enable()

	return nil
}

// Tick executes one instruction, and increases the program counter
// by 4 by default. If an exception is met, it will handle it.
func (c *cpu) Tick() *Excep {
	poll, code := c.interrupt.Poll()
	if poll {
		return c.Ienter(code, 0)
	}

	// no interrupt to dispatch, let's proceed
	e := c.tick()
	if e == nil {
		return nil
	}

	// proceed attempt failed, this is a fault.
	c.interrupt.Issue(e.Code)       // put the fault on to interrupt
	poll, code = c.interrupt.Poll() // see if it is handlable
	if poll {
		if code != e.Code {
			panic("interrupt code is different")
		}
		return c.Ienter(code, e.Arg) // pass it to the handler
	}

	// the interrupt handler is not handling it
	// this fault will be thrown out to the simulator
	return e
}
