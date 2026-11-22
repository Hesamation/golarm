//go:build linux || darwin

package clock

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func terminalCells() cellSize {
	sz, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || sz.Col == 0 || sz.Row == 0 {
		return cellSize{}
	}
	return cellSize{int(sz.Xpixel / sz.Col), int(sz.Ypixel / sz.Row)}
}

func probeGraphics() (bool, []byte) {
	if os.Getenv("TMUX") != "" || os.Getenv("STY") != "" || os.Getenv("TERM") == "dumb" {
		return false, nil
	}
	input, output := os.Stdin.Fd(), os.Stdout.Fd()
	if !term.IsTerminal(input) || !term.IsTerminal(output) {
		return false, nil
	}
	inInfo, inErr := os.Stdin.Stat()
	outInfo, outErr := os.Stdout.Stat()
	if inErr != nil || outErr != nil || !os.SameFile(inInfo, outInfo) {
		return false, nil
	}
	state, err := term.MakeRaw(input)
	if err != nil {
		return false, nil
	}
	defer term.Restore(input, state)

	id := imageID()
	query := ansi.KittyGraphics([]byte("AAAA"), fmt.Sprintf("i=%d", id), "a=q", "s=1", "v=1", "t=d", "f=24")
	if _, err := os.Stdout.WriteString(query + ansi.RequestPrimaryDeviceAttributes); err != nil {
		return false, nil
	}
	deadline := time.Now().Add(120 * time.Millisecond)
	poll := []unix.PollFd{{Fd: int32(input), Events: unix.POLLIN}}
	var response []byte
	for time.Now().Before(deadline) && len(response) < 4096 {
		ready, err := unix.Poll(poll, int(time.Until(deadline).Milliseconds()))
		if err == unix.EINTR {
			continue
		}
		if err != nil || ready == 0 || poll[0].Revents&unix.POLLIN == 0 {
			break
		}
		var buf [256]byte
		n, err := unix.Read(int(input), buf[:])
		if n > 0 {
			response = append(response, buf[:n]...)
			if ok, done, keys := probeReply(response, id); done {
				return ok, keys
			}
		}
		if err != nil && err != unix.EINTR {
			break
		}
	}
	_, _, keys := probeReply(response, id)
	return false, keys
}
