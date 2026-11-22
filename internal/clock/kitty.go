package clock

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"math/rand/v2"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

var primaryAttributes = regexp.MustCompile("\\x1b\\[\\?[0-9;]*c")

func probeReply(data []byte, id int) (supported, complete bool, keys []byte) {
	da := primaryAttributes.FindIndex(data)
	ack := []byte(ansi.KittyGraphics([]byte("OK"), fmt.Sprintf("i=%d", id)))
	supported = bytes.Contains(data, ack)
	keys = bytes.ReplaceAll(data, ack, nil)
	// A matching error still belongs to the probe, even on partial implementations.
	start := []byte(fmt.Sprintf("\x1b_Gi=%d;", id))
	if i := bytes.Index(keys, start); i >= 0 {
		if j := bytes.Index(keys[i:], []byte("\x1b\\")); j >= 0 {
			keys = bytes.Join([][]byte{keys[:i], keys[i+j+2:]}, nil)
		}
	}
	return supported, da != nil, primaryAttributes.ReplaceAll(keys, nil)
}

func imageID() int { return int(rand.Uint32N((1<<30)-1)) + 1 }

// kittyDisplay holds only the two graphics placements owned by this clock.
type kittyDisplay struct {
	dialID, handsID int
}

func newKittyDisplay() kittyDisplay {
	dialID := imageID()
	return kittyDisplay{dialID, dialID + (1 << 30)}
}

func (k kittyDisplay) clear() string {
	remove := func(id int) string {
		return ansi.KittyGraphics(nil, "a=d", "d=I", "q=2", fmt.Sprintf("i=%d", id))
	}
	return remove(k.dialID) + remove(k.handsID)
}

func kittyLayer(frame *image.RGBA, layout dialLayout, id, z int) string {
	var pngData bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(&pngData, frame); err != nil {
		return ""
	}
	payload := base64.StdEncoding.EncodeToString(pngData.Bytes())
	x, y := layout.pixelOrigin()
	var cmds strings.Builder
	cmds.WriteString(ansi.SaveCursor)
	cmds.WriteString(ansi.CursorPosition(x/layout.width+1, y/layout.height+1))
	options := []string{
		"a=T", "f=100", "q=2", "C=1", fmt.Sprintf("i=%d", id), fmt.Sprintf("z=%d", z),
		fmt.Sprintf("X=%d", x%layout.width), fmt.Sprintf("Y=%d", y%layout.height),
	}
	for len(payload) > kitty.MaxChunkSize {
		cmds.WriteString(ansi.KittyGraphics([]byte(payload[:kitty.MaxChunkSize]), append(options, "m=1")...))
		options = []string{"q=2"}
		payload = payload[kitty.MaxChunkSize:]
	}
	cmds.WriteString(ansi.KittyGraphics([]byte(payload), append(options, "m=0")...))
	cmds.WriteString(ansi.RestoreCursor)
	return cmds.String()
}

// kittyOutput keeps uploads and cleanup in Bubble Tea's own output stream.
// Preserving the file methods lets Bubble Tea continue handling TTY resizes.
type kittyOutput struct {
	*os.File
	display kittyDisplay
	active  bool
}

func (out *kittyOutput) WriteString(s string) (int, error) { return out.Write([]byte(s)) }

func (out *kittyOutput) Write(p []byte) (int, error) {
	var cleanup string
	hasGraphics := bytes.Contains(p, []byte("\x1b_G"))
	if out.active && bytes.Contains(p, []byte(ansi.ResetAltScreenSaveCursorMode)) {
		cleanup = out.display.clear()
		out.active = false
	}
	if hasGraphics {
		out.active = !bytes.Contains(p, []byte("\x1b_Ga=d"))
	}
	if !hasGraphics && cleanup == "" {
		return out.File.Write(p)
	}
	packet := make([]byte, 0, len(p)+len(cleanup)+24)
	packet = append(packet, ansi.SetSynchronizedOutputMode...)
	packet = append(packet, cleanup...)
	packet = append(packet, p...)
	packet = append(packet, ansi.ResetSynchronizedOutputMode...)
	if _, err := out.File.Write(packet); err != nil {
		return 0, err
	}
	return len(p), nil
}
