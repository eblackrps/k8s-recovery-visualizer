package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) fitWindowToScreen(ctx context.Context) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		runtime.WindowSetSize(ctx, preferredWindowWidth, preferredWindowHeight)
		runtime.WindowCenter(ctx)
		return
	}

	screen := screens[0]
	for _, candidate := range screens {
		if candidate.IsCurrent {
			screen = candidate
			break
		}
		if candidate.IsPrimary {
			screen = candidate
		}
	}

	width := fitWindowDimension(screen.Size.Width, preferredWindowWidth, 96, 720)
	height := fitWindowDimension(screen.Size.Height, preferredWindowHeight, 128, 540)

	runtime.WindowSetMinSize(ctx, minInt(width, minimumWindowWidth), minInt(height, minimumWindowHeight))
	runtime.WindowSetSize(ctx, width, height)
	runtime.WindowCenter(ctx)
}

func fitWindowDimension(screenSize, preferred, reserved, floor int) int {
	if screenSize <= 0 {
		return preferred
	}

	target := minInt(preferred, screenSize-reserved)
	if target < floor {
		target = maxInt(screenSize-48, floor)
	}
	if target > screenSize {
		target = screenSize
	}
	return maxInt(target, 360)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
