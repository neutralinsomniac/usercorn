package arm

import (
	"fmt"
	sysnum "github.com/lunixbochs/ghostrace/ghost/sys/num"
	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"

	"github.com/lunixbochs/usercorn/go/models"
	"github.com/lunixbochs/usercorn/go/syscalls"
)

var LinuxRegs = []int{uc.ARM_REG_R0, uc.ARM_REG_R1, uc.ARM_REG_R2, uc.ARM_REG_R3, uc.ARM_REG_R4, uc.ARM_REG_R5, uc.ARM_REG_R6}
var StaticUname = models.Uname{"Linux", "usercorn", "3.13.0-24-generic", "normal copy of Linux minding my business", "arm"}

func LinuxInit(u models.Usercorn, args, env []string) error {
	return u.PosixInit(args, env, nil)
}

func linux_uname(u syscalls.U, args []uint64) uint64 {
	addr, _ := u.RegRead(LinuxRegs[0])
	StaticUname.Pad(64)
	syscalls.Uname(u, addr, &StaticUname)
	return 0
}

var overrides = map[string]*syscalls.Syscall{
	"set_tls": {syscalls.Stub, A{PTR}, INT},
	"uname":   {linux_uname, A{PTR}, INT},
}

func LinuxSyscall(u models.Usercorn) {
	// TODO: handle errors or something
	num, _ := u.RegRead(uc.ARM_REG_R7)
	// TODO: EABI has a different syscall base (OABI is 0x900000)
	// TODO: does the generator handle this? it needs to.
	if num > 0x900000 {
		num -= 0x900000
	}
	name, _ := sysnum.Linux_arm[int(num)]
	override, _ := overrides[name]
	ret, _ := u.Syscall(int(num), name, syscalls.RegArgs(u, LinuxRegs), override)
	u.RegWrite(uc.ARM_REG_R0, ret)
}

func LinuxInterrupt(u models.Usercorn, intno uint32) {
	if intno == 2 {
		LinuxSyscall(u)
		return
	}
	panic(fmt.Sprintf("unhandled ARM interrupt: %d", intno))
}

func init() {
	Arch.RegisterOS(&models.OS{Name: "linux", Init: LinuxInit, Interrupt: LinuxInterrupt})
}
