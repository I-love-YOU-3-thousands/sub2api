package service

import (
	"context"
	"testing"
	"time"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/stretchr/testify/require"
)

func TestDetachUpstreamContextPreservesAttachedCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	upstream, release := detachUpstreamContext(withAttachedUpstreamContext(parent), time.Hour)
	defer release()
	attempt, releaseAttempt := pkghttputil.NewUpstreamAttemptContext(upstream, time.Hour)
	defer releaseAttempt()

	release()
	require.NoError(t, parent.Err())
	require.NoError(t, attempt.Err())

	cancel()

	require.ErrorIs(t, upstream.Err(), context.Canceled)
	require.ErrorIs(t, attempt.Err(), context.Canceled)
}

func TestDetachUpstreamContextPreservesAttachedDeadline(t *testing.T) {
	deadline := time.Now().Add(-time.Second)
	parent, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	upstream, release := detachUpstreamContext(withAttachedUpstreamContext(parent), time.Hour)
	defer release()
	attempt, releaseAttempt := pkghttputil.NewUpstreamAttemptContext(upstream, time.Hour)
	defer releaseAttempt()

	attemptDeadline, ok := attempt.Deadline()
	require.True(t, ok)
	require.Equal(t, deadline, attemptDeadline)
	require.ErrorIs(t, upstream.Err(), context.DeadlineExceeded)
	require.ErrorIs(t, attempt.Err(), context.DeadlineExceeded)
}

func TestWithAttachedUpstreamContextHandlesNil(t *testing.T) {
	upstream, release := detachUpstreamContext(withAttachedUpstreamContext(nil))
	defer release()

	require.NotNil(t, upstream)
	require.NoError(t, upstream.Err())
}
