package redis
import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"

	"example.com/go_practics/internal/model"
)

type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(addr, password string, db int, ttl time.Duration) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Cache{rdb: rdb, ttl: ttl}, nil
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}

func keyNote(id int64) string { return "note:" + itoa(id) }

func (c *Cache) GetNote(ctx context.Context, id int64) (model.Note, bool, error) {
	val, err := c.rdb.Get(ctx, keyNote(id)).Result()
	if err == redis.Nil {
		return model.Note{}, false, nil
	}
	if err != nil {
		return model.Note{}, false, err
	}
	var n model.Note
	if err := json.Unmarshal([]byte(val), &n); err != nil {
		return model.Note{}, false, err
	}
	return n, true, nil
}

func (c *Cache) SetNote(ctx context.Context, n model.Note) error {
	b, err := json.Marshal(n)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyNote(n.ID), b, c.ttl).Err()
}

func (c *Cache) DelNote(ctx context.Context, id int64) error {
	return c.rdb.Del(ctx, keyNote(id)).Err()
}

// tiny int64 -> string without fmt
func itoa(x int64) string {
	if x == 0 {
		return "0"
	}
	neg := x < 0
	if neg {
		x = -x
	}
	var buf [32]byte
	i := len(buf)
	for x > 0 {
		i--
		buf[i] = byte('0' + x%10)
		x /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}