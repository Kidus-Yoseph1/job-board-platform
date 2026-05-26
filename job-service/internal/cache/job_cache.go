package cache

import (
	"context"
	"time"

	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/redis/go-redis/v9"
)

const (
	jobKeyPrefix  = "job:"
	defaultJobTTL = 10 * time.Minute
)

type JobCache struct {
	cache *RedisCache
}

func NewJobCache(redisCache *RedisCache) *JobCache {
	return &JobCache{cache: redisCache}
}

func (jc *JobCache) GetJob(ctx context.Context, jobID string) (*db.Job, error) {
	key := jobKeyPrefix + jobID
	var job db.Job

	err := jc.cache.Get(ctx, key, &job)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (jc *JobCache) SetJob(ctx context.Context, job *db.Job) error {
	key := jobKeyPrefix + job.ID.String()
	return jc.cache.Set(ctx, key, job, defaultJobTTL)
}

func (jc *JobCache) InvalidateJob(ctx context.Context, jobID string) error {
	key := jobKeyPrefix + jobID
	return jc.cache.Delete(ctx, key)
}
