package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"iwut-smartclass-backend/internal/config"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Job interface {
	Execute() error
	GetID() string
	GetData() interface{}
	GetType() string
}

type WorkQueue struct {
	name           string        // 队列名称
	jobQueue       chan Job      // 任务通道
	workerPool     chan struct{} // 控制 Worker 并发量的信号
	workerCount    int           // Worker 数量
	wg             sync.WaitGroup
	ctx            context.Context
	cancelFunc     context.CancelFunc
	shutdownChan   chan struct{}        // 关闭信号通道
	persistenceDir string               // 持久化目录
	jobLoaders     map[string]JobLoader // Job 加载器
}

type JobLoader func([]byte, *config.Config) (Job, error)

var (
	queues        = make(map[string]*WorkQueue)
	queueMutex    sync.Mutex
	globalLoaders = make(map[string]JobLoader)
)

func NewWorkQueue(name string, workerCount int, queueSize int) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if q, exists := queues[name]; exists {
		return q
	}

	// 创建持久化目录
	persistenceDir := filepath.Join("data", "queues", name)
	if err := os.MkdirAll(persistenceDir, 0755); err != nil {
		Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to create persistence dir: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	queue := &WorkQueue{
		name:           name,
		jobQueue:       make(chan Job, queueSize),
		workerPool:     make(chan struct{}, workerCount),
		workerCount:    workerCount,
		ctx:            ctx,
		cancelFunc:     cancel,
		shutdownChan:   make(chan struct{}),
		persistenceDir: persistenceDir,
		jobLoaders:     make(map[string]JobLoader),
	}

	// 注册全局 Loaders
	for jobType, loader := range globalLoaders {
		queue.jobLoaders[jobType] = loader
	}

	queues[name] = queue
	Logger.Log("INFO", fmt.Sprintf("[Queue] Created work queue: %s with %d workers", name, workerCount))

	return queue
}

func RegisterGlobalLoader(jobType string, loader JobLoader) {
	globalLoaders[jobType] = loader
}

func (q *WorkQueue) RegisterLoader(jobType string, loader JobLoader) {
	q.jobLoaders[jobType] = loader
}

func (q *WorkQueue) Start(cfg *config.Config) {
	Logger.Log("INFO", fmt.Sprintf("[Queue] Starting work queue: %s", q.name))

	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.Worker(i)
	}

	// 恢复未完成的任务
	go q.Recover(cfg)
}

func (q *WorkQueue) Recover(cfg *config.Config) {
	files, err := os.ReadDir(q.persistenceDir)
	if err != nil {
		Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to read persistence dir: %v", err))
		return
	}

	count := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(q.persistenceDir, file.Name()))
		if err != nil {
			Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to read job file %s: %v", file.Name(), err))
			continue
		}

		var wrapper struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(content, &wrapper); err != nil {
			Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to unmarshal job wrapper %s: %v", file.Name(), err))
			continue
		}

		loader, ok := q.jobLoaders[wrapper.Type]
		if !ok {
			Logger.Log("WARN", fmt.Sprintf("[Queue] No loader found for job type %s", wrapper.Type))
			continue
		}

		job, err := loader(wrapper.Data, cfg)
		if err != nil {
			Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to load job %s: %v", file.Name(), err))
			continue
		}

		q.jobQueue <- job
		count++
	}
	if count > 0 {
		Logger.Log("INFO", fmt.Sprintf("[Queue] Recovered %d jobs for queue %s", count, q.name))
	}
}

func (q *WorkQueue) saveJob(job Job) error {
	data := map[string]interface{}{
		"type": job.GetType(),
		"data": job.GetData(),
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	filename := filepath.Join(q.persistenceDir, fmt.Sprintf("%s.json", job.GetID()))
	return os.WriteFile(filename, bytes, 0644)
}

func (q *WorkQueue) deleteJob(job Job) error {
	filename := filepath.Join(q.persistenceDir, fmt.Sprintf("%s.json", job.GetID()))
	return os.Remove(filename)
}

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
				Logger.Log("ERROR", fmt.Sprintf("[Queue] %s job failed after %s: %v", workerName, duration.String(), err.Error()))
			} else {
				Logger.Log("DEBUG", fmt.Sprintf("[Queue] %s job completed in %s", workerName, duration.String()))
				// 任务成功完成后删除持久化文件
				if deleteErr := q.deleteJob(job); deleteErr != nil {
					Logger.Log("WARN", fmt.Sprintf("[Queue] Failed to delete persisted job: %v", deleteErr))
				}
			}
		}
	}
}

func (q *WorkQueue) AddJob(job Job) {
	select {
	case <-q.ctx.Done():
		Logger.Log("WARN", fmt.Sprintf("[Queue] Attempting to add job to stopped queue: %s", q.name))
		return
	default:
		// 持久化任务
		if err := q.saveJob(job); err != nil {
			Logger.Log("ERROR", fmt.Sprintf("[Queue] Failed to persist job: %v", err))
		}
		q.jobQueue <- job
	}
}

func (q *WorkQueue) Stop() {
	Logger.Log("INFO", fmt.Sprintf("[Queue] Stopping queue: %s", q.name))
	q.cancelFunc()
	close(q.jobQueue)
	q.wg.Wait()
	close(q.shutdownChan)
	Logger.Log("INFO", fmt.Sprintf("[Queue] Queue stopped: %s", q.name))
}

func GetQueue(name string) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return queues[name]
}

func InitQueues(cfg *config.Config) {
	// Summary Service
	summaryQueue := NewWorkQueue("SummaryServiceQueue", cfg.SummaryWorkerCount, cfg.SummaryQueueSize)
	summaryQueue.Start(cfg)
}
