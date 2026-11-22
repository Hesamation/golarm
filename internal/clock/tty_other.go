//go:build !linux && !darwin

package clock

func terminalCells() cellSize { return cellSize{} }

func probeGraphics() (bool, []byte) { return false, nil }
