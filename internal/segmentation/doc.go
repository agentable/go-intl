// Package segmentation exposes text segmentation backend capabilities.
//
// It keeps supported-locale claims tied to the active Unicode segmentation
// implementation instead of advertising unsupported dictionary behavior.
//
// Only the Segmenter implementation should use this package; public callers use
// segmenter APIs.
package segmentation
