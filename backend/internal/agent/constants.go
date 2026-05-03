package agent

import "time"

// State transitions
const (
	SilenceTimeout      = 4500 * time.Millisecond // Working→Online silence threshold
	SilenceTimeoutBusy  = 15 * time.Second         // Extended silence when tool_in_flight
	StabilizeDebounce   = 1000 * time.Millisecond  // Working→Online debounce
	StatusPollInterval  = 500 * time.Millisecond   // Status poll frequency
	WorkingIntentWindow = 1500 * time.Millisecond  // Post-command Working intent window
	ShellReadyTimeout   = 3000 * time.Millisecond  // Shell ready timeout
	ShellReadyMinWait   = 500 * time.Millisecond   // Minimum wait before "activity = ready"
	ShellReadyMinBytes  = 1024                     // Output bytes to consider shell ready
	PostReadyStable     = 1200 * time.Millisecond  // Post-ready flow gate
	PostReadyTick       = 600 * time.Millisecond   // Post-ready no-output trigger interval
	ConnectTimeout      = 10 * time.Second
)

// Dispatch queue
const (
	DispatchQueueSize = 32
	MergeWindow       = 300 * time.Millisecond
	DedupWindow       = 30 * time.Second
	DedupRingSize     = 128
	ForceFlushTimeout = 30 * time.Second
	CommandConfirmDelay = 100 * time.Millisecond
	MaxMergedLength   = 4096 // Split merged messages beyond this
)

// Flow control
const (
	FlowHighWaterMark  = 200 * 1024 // 200KB pause threshold
	FlowLowWaterMark   = 20 * 1024  // 20KB resume threshold
	OutputEmitInterval = 16 * time.Millisecond // ~60fps
	OutputEmitMaxBytes = 64 * 1024
	OutputQueueCapacity = 256
)

// Outbox
const (
	OutboxPollInterval = 280 * time.Millisecond
	OutboxClaimLimit   = 8
	OutboxLeaseDuration = 8 * time.Second
	OutboxMaxRetries   = 6
	OutboxBaseBackoff  = 800 * time.Millisecond
	OutboxMaxBackoff   = 30 * time.Second
)

// Chat
const (
	ChatSilenceTimeout = 3000 * time.Millisecond
	ChatIdleDebounce   = 1000 * time.Millisecond
	ChatForceFlush     = 30 * time.Second
)

// Terminal
const (
	ScrollbackLines      = 2000
	RedrawSuppressWindow = 400 * time.Millisecond
	IdleTimeout          = 30 * time.Minute
	TerminalAckTimeout   = 5 * time.Second  // Timeout for terminal acks
	TerminalResizeDebounce = 100 * time.Millisecond // Debounce for resize events
)

// Session generation
const (
	SpawnEpochMax = 1000 // Maximum session generation epochs for recovery
)
