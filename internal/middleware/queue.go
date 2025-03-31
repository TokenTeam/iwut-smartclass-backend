package middleware

import (
	"context"
	"fmt"
	"iwut-smart-timetable-backend/internal/config"
	"sync"
	"time"
)

type Job interface {
	Execute() error
}

// WorkQueue 工作队列
type WorkQueue struct {
	name         string        // 队列名称
	jobQueue     chan Job      // 任务通道
	workerPool   chan struct{} // 控制 Worker 并发量的信号
	workerCount  int           // Worker 数量
	wg           sync.WaitGroup
	ctx          context.Context
	cancelFunc   context.CancelFunc
	shutdownChan chan struct{} // 关闭信号通道
}

var (
	queues     = make(map[string]*WorkQueue)
	queueMutex sync.Mutex
)

// NewWorkQueue 初始化工作队列
func NewWorkQueue(name string, workerCount int, queueSize int) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if q, exists := queues[name]; exists {
		return q
	}

	ctx, cancel := context.WithCancel(context.Background())
	queue := &WorkQueue{
		name:         name,
		jobQueue:     make(chan Job, queueSize),
		workerPool:   make(chan struct{}, workerCount),
		workerCount:  workerCount,
		ctx:          ctx,
		cancelFunc:   cancel,
		shutdownChan: make(chan struct{}),
	}

	queues[name] = queue
	Logger.Log("INFO", fmt.Sprintf("[Queue] Created work queue: %s with %d workers", name, workerCount))

	return queue
}

// Start 启动工作队列
func (q *WorkQueue) Start() {
	Logger.Log("INFO", fmt.Sprintf("[Queue] Starting work queue: %s", q.name))

	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.Worker(i)
	}
}

// Worker 工作函数
func (q *WorkQueue) Worker(id int) {
	defer q.wg.Done()
	workerName := fmt.Sprintf("%s-worker-%d", q.name, id)
	Logger.Log("DEBUG", fmt.Sprintf("[Queue] Started %s", workerName))

	for {
		select {
		case <-q.ctx.Done():
			Logger.Log("DEBUG", fmt.Sprintf("[Queue] Shutting down %s", workerName))
			return
		case job, ok := <-q.jobQueue:
			if !ok {
				return
			}

			// 获取信号
			q.workerPool <- struct{}{}

			// 执行任务
			start := time.Now()
			err := job.Execute()
			duration := time.Since(start)

			// 释放信号
			<-q.workerPool

			if err != nil {
				Logger.Log("ERROR", fmt.Sprintf("[Queue] %s job failed after %s: %s", workerName, duration.String(), err.Error()))
			} else {
				Logger.Log("DEBUG", fmt.Sprintf("[Queue] %s job completed in %s", workerName, duration.String()))
			}
		}
	}
}

// AddJob 添加任务到队列
func (q *WorkQueue) AddJob(job Job) {
	select {
	case <-q.ctx.Done():
		Logger.Log("WARN", fmt.Sprintf("[Queue] Attempting to add job to stopped queue: %s", q.name))
		return
	default:
		q.jobQueue <- job
	}
}

// Stop 停止工作队列
func (q *WorkQueue) Stop() {
	Logger.Log("INFO", fmt.Sprintf("[Queue] Stopping queue: %s", q.name))
	q.cancelFunc()
	close(q.jobQueue)
	q.wg.Wait()
	close(q.shutdownChan)
	Logger.Log("INFO", fmt.Sprintf("[Queue] Queue stopped: %s", q.name))
}

// GetQueue 获取已创建的队列
func GetQueue(name string) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return queues[name]
}

// InitQueues 初始化所有工作队列
func InitQueues(cfg *config.Config) {
	// Summary Service
	summaryQueue := NewWorkQueue("SummaryServiceQueue", cfg.SummaryWorkerCount, cfg.SummaryQueueSize)
	summaryQueue.Start()
}
