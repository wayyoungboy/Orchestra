package terminal

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

func startPty(cmd *exec.Cmd) (*os.File, error) {
	// 设置进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func resizePty(f *os.File, cols, rows int) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct {
			Rows uint16
			Cols uint16
			X    uint16
			Y    uint16
		}{
			Rows: uint16(rows),
			Cols: uint16(cols),
		})),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func init() {
	// 忽略 SIGCHLD，让子进程自动回收
	signal.Ignore(syscall.SIGCHLD)
}