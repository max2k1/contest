package xcontext

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackgroundContext(t *testing.T) {
	ctx := Background()

	var blocked bool
	select {
	case <-ctx.Until(nil):
	default:
		blocked = true
	}

	require.True(t, blocked)
	require.Nil(t, ctx.Err())
	require.Nil(t, ctx.Notifications())
}

func TestWaiterGC(t *testing.T) {
	goroutines := runtime.NumGoroutine()

	ctx, cancelFunc := WithCancel(nil)
	var wg sync.WaitGroup
	wg.Add(1)
	readyToSelect := make(chan struct{})
	startSelect := make(chan struct{})
	go func() {
		defer wg.Done()
		ch := ctx.Until(nil)
		runtime.GC()
		select {
		case <-ch:
			t.Fail()
		default:
		}
		close(readyToSelect)
		<-startSelect
		<-ch
	}()
	<-readyToSelect
	cancelFunc()
	close(startSelect)
	wg.Wait()

	runtime.GC()
	require.Equal(t, goroutines, runtime.NumGoroutine())
}

func TestContextCanceled(t *testing.T) {
	ctx, cancelFunc := WithCancel(nil)
	require.NotNil(t, ctx)

	cancelFunc()

	<-ctx.Done()
	<-ctx.Until(nil)

	require.Equal(t, Canceled, ctx.Err())
	require.Nil(t, ctx.Notifications())

	var paused bool
	select {
	case <-ctx.Until(Paused):
		paused = true
	default:
	}
	require.False(t, paused)
}

func TestContextPaused(t *testing.T) {
	ctx, pauseFunc := WithNotify(nil, Paused)
	require.NotNil(t, ctx)

	pauseFunc()

	<-ctx.Until(Paused)
	<-ctx.Until(nil)

	require.Nil(t, ctx.Err())
	require.Equal(t, []error{Paused}, ctx.Notifications())

	var canceled bool
	select {
	case <-ctx.Done():
		canceled = true
	default:
	}
	require.False(t, canceled)
}

func TestMultipleCancelPause(t *testing.T) {
	ctx, cancelFunc := WithCancel(nil)
	ctx, pauseFunc := WithNotify(ctx, Paused)
	require.NotNil(t, ctx)

	pauseFunc()
	cancelFunc()
	pauseFunc()
	cancelFunc()

	<-ctx.Done()
	<-ctx.Until(nil)
	<-ctx.Until(Paused)

	require.Equal(t, Canceled, ctx.Err())
	require.Equal(t, []error{Paused, Paused}, ctx.Notifications())
}

func TestGrandGrandGrandChild(t *testing.T) {
	ctx0, notifyFunc := WithNotify(nil, Paused)
	ctx1, _ := WithCancel(ctx0.Clone())
	ctx2, _ := WithNotify(ctx1.Clone(), Paused)

	require.False(t, ctx2.IsSignaledWith(Paused))
	notifyFunc()
	<-ctx2.Until(Paused)
}

func TestWithParent(t *testing.T) {
	t.Run("pause_propagated", func(t *testing.T) {
		parentCtx, pauseFunc := WithNotify(nil, Paused)
		childCtx := parentCtx.Clone()

		pauseFunc()
		<-childCtx.Until(Paused)
		require.Equal(t, []error{Paused}, childCtx.Notifications())
		<-childCtx.Until(nil)

		require.Nil(t, childCtx.Err())
	})
	t.Run("cancel_propagated", func(t *testing.T) {
		parentCtx, parentCancelFunc := WithCancel(nil)
		childCtx := parentCtx.Clone()

		parentCancelFunc()
		<-childCtx.Done()
		require.Equal(t, Canceled, childCtx.Err())
		<-childCtx.Until(nil)

		require.Nil(t, childCtx.Notifications())
	})
	t.Run("pause_cancel_happened_before_context_created", func(t *testing.T) {
		parentCtx, cancelFunc := WithCancel(nil)
		parentCtx, pauseFunc := WithNotify(parentCtx, Paused)
		pauseFunc()
		cancelFunc()

		childCtx := parentCtx.Clone()
		<-childCtx.Done()
		<-childCtx.Until(Paused)
		<-childCtx.Until(nil)

		require.Equal(t, Canceled, childCtx.Err())
		require.Equal(t, []error{Paused}, childCtx.Notifications())
	})
	t.Run("child_pause_cancel_do_not_affect_parent", func(t *testing.T) {
		parentCtxs := map[string]Context{}
		parentCtxs["background"] = Background()
		parentCtxs["with_cancel"], _ = WithCancel(nil)
		for parentName, parentCtx := range parentCtxs {
			t.Run(parentName, func(t *testing.T) {
				childCtx, cancelFunc := WithCancel(parentCtx.Clone())
				_, pauseFunc := WithNotify(childCtx, Paused)

				pauseFunc()
				cancelFunc()

				var blocked bool
				select {
				case <-parentCtx.Until(nil):
				default:
					blocked = true
				}

				require.True(t, blocked)
			})
		}
	})
}
