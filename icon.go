//go:build windows && !arm64
// +build windows,!arm64

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CDIcon creates a custom CD/DVD icon for the application
type CDIcon struct{}

func (c *CDIcon) Name() string {
	return "cd-icon"
}

func (c *CDIcon) Content() []byte {
	// SVG for a gold CD/DVD disc icon with transparent background
	svg := `<?xml version="1.0" encoding="UTF-8"?>
<svg width="64" height="64" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">
  <!-- Outer ring (gold) -->
  <circle cx="32" cy="32" r="30" fill="#FFD700" stroke="#DAA520" stroke-width="1"/>
  
  <!-- Inner reflecting arc (lighter gold) -->
  <path d="M 15 25 Q 32 20, 49 25" fill="none" stroke="#FFEC8B" stroke-width="2" opacity="0.6"/>
  
  <!-- Center hole (black) -->
  <circle cx="32" cy="32" r="8" fill="#333333" stroke="#1a1a1a" stroke-width="1"/>
  
  <!-- Center hole highlight -->
  <circle cx="32" cy="32" r="6" fill="#4d4d4d" stroke="none"/>
  
  <!-- Data tracks (subtle) -->
  <circle cx="32" cy="32" r="24" fill="none" stroke="#DAA520" stroke-width="0.5" opacity="0.3"/>
  <circle cx="32" cy="32" r="20" fill="none" stroke="#DAA520" stroke-width="0.5" opacity="0.3"/>
  <circle cx="32" cy="32" r="16" fill="none" stroke="#DAA520" stroke-width="0.5" opacity="0.3"/>
  <circle cx="32" cy="32" r="12" fill="none" stroke="#DAA520" stroke-width="0.5" opacity="0.3"/>
  
  <!-- Shine/reflection effect -->
  <path d="M 20 18 Q 32 15, 44 18 Q 32 21, 20 18" fill="white" opacity="0.4"/>
</svg>`
	return []byte(svg)
}

// GetAppIcon returns the custom CD icon resource
func GetAppIcon() fyne.Resource {
	return theme.NewThemedResource(&CDIcon{})
}
