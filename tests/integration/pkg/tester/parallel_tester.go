package tester

import (
	"context"
	"fmt"
	"sync"
)

type Tester interface {
	Name() string
	Start()
	Stop() []TestResult
}

func NewTester(name string, testFn func() error, numberOfThreads int) Tester {
	return &tester{
		name:            name,
		testFn:          testFn,
		numberOfWorkers: numberOfThreads,
	}
}

type TestResult struct {
	WorkerName string
	TestCount  int
	Err        error
}

type tester struct {
	name            string
	testFn          func() error
	numberOfWorkers int
	resultChans     []chan TestResult
	cancel          func()
	waitGroup       *sync.WaitGroup
}

type worker struct {
	name       string
	test       func() error
	testCount  int
	err        error
	resultChan chan TestResult
}

func (w *worker) doWorkInBackground(ctx context.Context, group *sync.WaitGroup) {
	group.Add(1)
	go func() {
		defer group.Done()
		w.doWork(ctx)
	}()
}

func (w *worker) sendResult() {
	w.resultChan <- TestResult{
		WorkerName: w.name,
		TestCount:  w.testCount,
		Err:        w.err,
	}
	close(w.resultChan)
}

func (w *worker) doWork(ctx context.Context) {
	defer w.sendResult()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.testCount++
			err := w.test()
			if err != nil {
				w.err = fmt.Errorf("test %d done by worker %s failed with error %v", w.testCount, w.name, err)
				return
			}
		}
	}
}

func (t *tester) Name() string {
	return t.name
}

func (t *tester) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.waitGroup = &sync.WaitGroup{}
	t.resultChans = make([]chan TestResult, t.numberOfWorkers)

	for i := 0; i < t.numberOfWorkers; i++ {
		t.resultChans[i] = make(chan TestResult, 1)
		w := worker{
			name:       fmt.Sprintf("%s-%d", t.name, i),
			test:       t.testFn,
			resultChan: t.resultChans[i],
		}

		w.doWorkInBackground(ctx, t.waitGroup)
	}
}

func (t *tester) Stop() []TestResult {
	results := make([]TestResult, 0)
	t.cancel()
	t.waitGroup.Wait()
	for _, resultChan := range t.resultChans {
		result := <-resultChan
		results = append(results, result)
	}
	return results
}
