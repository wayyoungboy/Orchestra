package agent

import (
	"context"
	"sync"
	"time"
)

// FlowController implements output throttling and backpressure.
// It limits the rate of output sent to WebSocket clients.
//
// The controller:
// - Throttles output to OutputEmitInterval (~60fps)
// - Buffers up to OutputEmitMaxBytes per emit
// - Signals pause when buffered bytes exceed FlowHighWaterMark
// - Signals resume when buffered bytes drop below FlowLowWaterMark
type FlowController struct {
	mu           sync.Mutex
	buffer       []byte
	buffered     int
	paused       bool
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	emitChan     chan []byte // channel for emitted data
	OnPause      func()      // callback when high water mark exceeded
	OnResume     func()      // callback when low water mark crossed
	totalQueued  int64       // total bytes queued (for stats)
	totalEmitted int64       // total bytes emitted (for stats)
}

// NewFlowController creates a new flow controller.
func NewFlowController() *FlowController {
	return &FlowController{
		buffer:   make([]byte, 0, OutputEmitMaxBytes*2),
		emitChan: make(chan []byte, OutputQueueCapacity),
	}
}

// Start begins the background emit goroutine.
func (f *FlowController) Start(ctx context.Context) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return
	}

	f.ctx, f.cancel = context.WithCancel(ctx)
	f.running = true

	go f.emitLoop()
}

// Stop stops the controller.
func (f *FlowController) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return
	}

	f.running = false
	if f.cancel != nil {
		f.cancel()
	}
	close(f.emitChan)
}

// Queue adds data to the buffer, applying backpressure if needed.
// Returns true if the data was queued, false if the controller is stopped.
func (f *FlowController) Queue(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return false
	}

	// Add to buffer
	f.buffer = append(f.buffer, data...)
	f.buffered += len(data)
	f.totalQueued += int64(len(data))

	// Check water marks
	f.checkWaterMarksLocked()

	return true
}

// QueueAndGetSignal adds data and returns whether the consumer should pause.
func (f *FlowController) QueueAndGetSignal(data []byte) (queued bool, shouldPause bool) {
	if len(data) == 0 {
		return true, false
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return false, false
	}

	// Add to buffer
	f.buffer = append(f.buffer, data...)
	f.buffered += len(data)
	f.totalQueued += int64(len(data))

	// Check water marks
	f.checkWaterMarksLocked()

	return true, f.paused
}

// emitLoop is the background goroutine that drains at throttle rate.
func (f *FlowController) emitLoop() {
	ticker := time.NewTicker(OutputEmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			// Drain remaining before exit
			f.emitRemaining()
			return
		case <-ticker.C:
			f.emitBatch()
		}
	}
}

// emitBatch emits one batch from the buffer.
func (f *FlowController) emitBatch() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.buffered == 0 {
		return
	}

	// Determine batch size (max OutputEmitMaxBytes)
	batchSize := f.buffered
	if batchSize > OutputEmitMaxBytes {
		batchSize = OutputEmitMaxBytes
	}

	// Extract batch
	batch := make([]byte, batchSize)
	copy(batch, f.buffer[:batchSize])

	// Shift buffer
	f.buffer = f.buffer[batchSize:]
	f.buffered -= batchSize
	f.totalEmitted += int64(batchSize)

	// Check water marks (might resume now)
	f.checkWaterMarksLocked()

	// Send to emit channel (non-blocking)
	select {
	case f.emitChan <- batch:
	default:
		// Channel full, drop batch (consumer too slow)
	}
}

// emitRemaining drains all remaining buffer before shutdown.
func (f *FlowController) emitRemaining() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for f.buffered > 0 {
		batchSize := f.buffered
		if batchSize > OutputEmitMaxBytes {
			batchSize = OutputEmitMaxBytes
		}

		batch := make([]byte, batchSize)
		copy(batch, f.buffer[:batchSize])

		f.buffer = f.buffer[batchSize:]
		f.buffered -= batchSize
		f.totalEmitted += int64(batchSize)

		// Send to emit channel (non-blocking, but we're shutting down)
		select {
		case f.emitChan <- batch:
		default:
		}
	}
}

// checkWaterMarksLocked checks water marks and fires callbacks.
// Must be called with mu held.
func (f *FlowController) checkWaterMarksLocked() {
	if !f.paused && f.buffered >= FlowHighWaterMark {
		// Cross high water mark → pause
		f.paused = true
		if f.OnPause != nil {
			go f.OnPause() // Fire async to avoid deadlock
		}
	} else if f.paused && f.buffered <= FlowLowWaterMark {
		// Cross low water mark → resume
		f.paused = false
		if f.OnResume != nil {
			go f.OnResume() // Fire async to avoid deadlock
		}
	}
}

// EmitChan returns the channel for emitted batches.
func (f *FlowController) EmitChan() <-chan []byte {
	return f.emitChan
}

// IsPaused returns true if the controller is in backpressure mode.
func (f *FlowController) IsPaused() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.paused
}

// BufferedBytes returns the current buffer size.
func (f *FlowController) BufferedBytes() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.buffered
}

// Stats returns flow controller statistics.
func (f *FlowController) Stats() (queued, emitted int64, buffered int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.totalQueued, f.totalEmitted, f.buffered
}

// SetCallbacks sets the pause/resume callbacks.
func (f *FlowController) SetCallbacks(onPause, onResume func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.OnPause = onPause
	f.OnResume = onResume
}
