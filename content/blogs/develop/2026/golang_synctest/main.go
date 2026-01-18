package main

import (
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

func matchString(a, b string) (bool, error) {
	return a == b, nil
}

func main() {
	testSuite := []testing.InternalTest{
		{
			Name: "TestWithRealSleep",
			F:    TestWithRealSleep,
		},
		{
			Name: "TestWithSynctest",
			F:    TestWithSynctest,
		},
	}
	testing.Main(matchString, testSuite, nil, nil)
}

func simulateLongRunningJob(d time.Duration) {
	time.Sleep(d)
}

func TestWithRealSleep(t *testing.T) {

	waitDuration := 100 * time.Millisecond
	start := time.Now()

	fmt.Println("Starting real sleep test...")
	simulateLongRunningJob(waitDuration)

	elapsed := time.Since(start)
	fmt.Printf("TestWithRealSleep exec time: %v\n", elapsed)

	if elapsed < waitDuration {
		t.Errorf("Expected at least 100 ms, got %v", elapsed)
	}
}

func TestWithSynctest(t *testing.T) {
	start := time.Now()

	fmt.Println("Starting synctest...")

	synctest.Test(t, func(t *testing.T) {
		done := make(chan struct{})

		go func() {
			defer close(done)
			simulateLongRunningJob(2 * time.Second)
		}()

		<-done
	})

	elapsed := time.Since(start)
	fmt.Printf("TestWithSynctest exec time: %v\n", elapsed)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected instantaneous execution, got %v", elapsed)
	}
}
