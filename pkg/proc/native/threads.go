package native

import (
	"fmt"

	"github.com/go-delve/delve/pkg/proc"
)

// Thread represents a single thread in the traced process
// ID represents the thread id or port, Process holds a reference to the
// Process struct that contains info on the process as
// a whole, and Status represents the last result of a `wait` call
// on this thread.
type Thread struct {
	ID     int         // Thread ID or mach port
	Status *WaitStatus // Status returned from last wait call

	dbp            *Process
	singleStepping bool
	os             *OSSpecificDetails
	common         proc.CommonThread
}

// Continue the execution of this thread.
//
// If we are currently at a breakpoint, we'll clear it
// first and then resume execution. Thread will continue until
// it hits a breakpoint or is signaled.
func (t *Thread) Continue() error {
	return t.resume()
}

// StepInstruction steps a single instruction.
//
// Executes exactly one instruction and then returns.
// If the thread is at a breakpoint, we first clear it,
// execute the instruction, and then replace the breakpoint.
// Otherwise we simply execute the next instruction.
func (t *Thread) StepInstruction() (err error) {
	t.singleStepping = true
	defer func() {
		t.singleStepping = false
	}()

	err = t.singleStep()
	if err != nil {
		if _, exited := err.(proc.ErrProcessExited); exited {
			return err
		}
		return fmt.Errorf("step failed: %s", err.Error())
	}
	return nil
}

// Common returns information common across Process
// implementations.
func (t *Thread) Common() *proc.CommonThread {
	return &t.common
}

// ThreadID returns the ID of this thread.
func (t *Thread) ThreadID() int {
	return t.ID
}

// ClearBreakpoint clears the specified breakpoint.
func (t *Thread) ClearBreakpoint(pc uint64, originalData []byte) error {
	if _, err := t.WriteMemory(uintptr(pc), originalData); err != nil {
		return fmt.Errorf("could not clear breakpoint %s", err)
	}
	return nil
}

// Registers obtains register values from the debugged process.
func (t *Thread) Registers(floatingPoint bool) (proc.Registers, error) {
	return registers(t, floatingPoint)
}

// RestoreRegisters will set the value of the CPU registers to those
// passed in via 'savedRegs'.
func (t *Thread) RestoreRegisters(savedRegs proc.Registers) error {
	return t.restoreRegisters(savedRegs)
}

// PC returns the current program counter value for this thread.
func (t *Thread) PC() (uint64, error) {
	regs, err := t.Registers(false)
	if err != nil {
		return 0, err
	}
	return regs.PC(), nil
}
