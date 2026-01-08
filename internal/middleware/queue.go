package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"iwut-smartclass-backend/internal/infrastructure/config"
	loggerPkg "iwut-smartclass-backend/internal/infrastructure/logger"
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
	logger         loggerPkg.Logger     // 日志
}

type JobLoader func([]byte, *config.Config) (Job, error)

var (
	queues        = make(map[string]*WorkQueue)
	queueMutex    sync.Mutex
	globalLoaders = make(map[string]JobLoader)
)

func NewWorkQueue(name string, workerCount int, queueSize int, logger loggerPkg.Logger) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if q, exists := queues[name]; exists {
		return q
	}

	// 创建持久化目录
	persistenceDir := filepath.Join("data", "queues", name)
	if err := os.MkdirAll(persistenceDir, 0755); err != nil {
		logger.Error("failed to create persistence dir", loggerPkg.String("error", fmt.Sprintf("%v", err)))
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
		logger:         logger,
	}

	// 注册全局 Loaders
	for jobType, loader := range globalLoaders {
		queue.jobLoaders[jobType] = loader
	}

	queues[name] = queue
	logger.Info("created work queue", loggerPkg.String("name", name), loggerPkg.String("workers", fmt.Sprintf("%d", workerCount)))

	return queue
}

func RegisterGlobalLoader(jobType string, loader JobLoader) {
	globalLoaders[jobType] = loader
}

func (q *WorkQueue) RegisterLoader(jobType string, loader JobLoader) {
	q.jobLoaders[jobType] = loader
}

func (q *WorkQueue) Start(cfg *config.Config) {
	q.logger.Info("starting work queue", loggerPkg.String("name", q.name))

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
		q.logger.Error("failed to read persistence dir", loggerPkg.String("error", err.Error()))
		return
	}

	count := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(q.persistenceDir, file.Name()))
		if err != nil {
			q.logger.Error("failed to read job file", loggerPkg.String("file", file.Name()), loggerPkg.String("error", err.Error()))
			continue
		}

		var wrapper struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(content, &wrapper); err != nil {
			q.logger.Error("failed to unmarshal job wrapper", loggerPkg.String("file", file.Name()), loggerPkg.String("error", err.Error()))
			continue
		}

		loader, ok := q.jobLoaders[wrapper.Type]
		if !ok {
			q.logger.Warn("no loader found for job type", loggerPkg.String("type", wrapper.Type))
			continue
		}

		job, err := loader(wrapper.Data, cfg)
		if err != nil {
			q.logger.Error("failed to load job", loggerPkg.String("file", file.Name()), loggerPkg.String("error", err.Error()))
			continue
		}

		q.jobQueue <- job
		count++
	}
	if count > 0 {
		q.logger.Info("recovered jobs", loggerPkg.String("count", fmt.Sprintf("%d", count)), loggerPkg.String("queue", q.name))
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
	q.logger.Debug("started worker", loggerPkg.String("worker", workerName))

	for {
		select {
		case <-q.ctx.Done():
			q.logger.Debug("shutting down worker", loggerPkg.String("worker", workerName))
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
				q.logger.Error("job failed", loggerPkg.String("worker", workerName), loggerPkg.String("duration", duration.String()), loggerPkg.String("error", err.Error()))
			} else {
				q.logger.Debug("job completed", loggerPkg.String("worker", workerName), loggerPkg.String("duration", duration.String()))
				// 任务成功完成后删除持久化文件
				if deleteErr := q.deleteJob(job); deleteErr != nil {
					q.logger.Warn("failed to delete persisted job", loggerPkg.String("error", deleteErr.Error()))
				}
			}
		}
	}
}

func (q *WorkQueue) AddJob(job Job) {
	select {
	case <-q.ctx.Done():
		q.logger.Warn("attempting to add job to stopped queue", loggerPkg.String("queue", q.name))
		return
	default:
		// 持久化任务
		if err := q.saveJob(job); err != nil {
			q.logger.Error("failed to persist job", loggerPkg.String("error", err.Error()))
		}
		q.jobQueue <- job
	}
}

func (q *WorkQueue) Stop() {
	q.logger.Info("stopping queue", loggerPkg.String("queue", q.name))
	q.cancelFunc()
	close(q.jobQueue)
	q.wg.Wait()
	close(q.shutdownChan)
	q.logger.Info("queue stopped", loggerPkg.String("queue", q.name))
}

func GetQueue(name string) *WorkQueue {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return queues[name]
}

func InitQueues(cfg *config.Config, logger loggerPkg.Logger) {
	// Summary Service
	summaryQueue := NewWorkQueue("SummaryServiceQueue", cfg.SummaryWorkerCount, cfg.SummaryQueueSize, logger)
	summaryQueue.Start(cfg)
}
