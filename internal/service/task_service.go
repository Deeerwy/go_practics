package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/task"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *task.Repo
	redis *redis.Client
	cfg   config.Config
}

func NewTaskService(repo *task.Repo, redisClient *redis.Client, cfg config.Config) *TaskService {
	return &TaskService{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (task.Task, error) {
	key := cache.TaskByIDKey(id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var t task.Task
			if err := json.Unmarshal([]byte(cached), &t); err == nil {
				log.Printf("[CACHE HIT] key=%s", key)   //  cache hit
				return t, nil
			}
			log.Printf("[CACHE ERROR] decode failed for key=%s: %v", key, err)
		} else if !errors.Is(err, redis.Nil) {
			log.Printf("[CACHE ERROR] redis read failed for key=%s: %v", key, err)
		} else {
			log.Printf("[CACHE MISS] key=%s", key)      //  cache miss
		}
	}

	t, err := s.repo.GetByID(id)
	if err != nil {
		return task.Task{}, err
	}

	if s.redis != nil {
		bytes, err := json.Marshal(t)
		if err != nil {
			log.Printf("[CACHE ERROR] encode failed for key=%s: %v", key, err)
			return t, nil
		}

		ttl := cache.TTLWithJitter(s.cfg.CacheTTL, s.cfg.CacheTTLJitter)
		if err := s.redis.Set(ctx, key, bytes, ttl).Err(); err != nil {
			log.Printf("[CACHE ERROR] redis write failed for key=%s: %v", key, err)
		} else {
			log.Printf("[CACHE SET] key=%s ttl=%s", key, ttl)   //  cache set
		}
	}

	return t, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, t task.Task) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(t.ID)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Printf("[CACHE ERROR] redis delete failed for key=%s: %v", key, err)
		} else {
			log.Printf("[CACHE INVALIDATED] key=%s (reason: update)", key)   //  cache invalidated
		}
	}

	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(id)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Printf("[CACHE ERROR] redis delete failed for key=%s: %v", key, err)
		} else {
			log.Printf("[CACHE INVALIDATED] key=%s (reason: delete)", key)   // cache invalidated
		}
	}

	return nil
}
