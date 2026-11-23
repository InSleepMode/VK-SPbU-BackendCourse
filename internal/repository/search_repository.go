package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
	"github.com/redis/go-redis/v9"
)

func tokenize(text string) []string {
	lowered := strings.ToLower(text)
	words := strings.FieldsFunc(lowered, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return words

}

type SearchRepository struct {
	client *redis.Client
}

func NewSearchRepository(client *redis.Client) *SearchRepository {
	return &SearchRepository{client: client}
}

func (r *SearchRepository) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64 = 0

	for i := 0; ; i++ {
		keys, curs, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			err := r.client.Del(ctx, keys...).Err()
			if err != nil {
				return err
			}
		}
		cursor = curs
		if curs == 0 {
			break
		}
	}
	return nil
}

func (r *SearchRepository) BuildIndex(ctx context.Context, movies []domain.Movie) error {
	err := r.deleteByPattern(ctx, "index:word:*") //сбросим старый индекс
	if err != nil {
		return err
	}
	err = r.deleteByPattern(ctx, "cache:search:*") //сбросим кэш
	if err != nil {
		return err
	}
	if len(movies) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, m := range movies {
		var words []string
		words = tokenize(m.Title + " " + m.Director)
		for _, word := range words {
			pipe.SAdd(ctx, "index:word:"+word, m.ID)
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

func stringsToInts(strings []string) ([]int, error) {
	intNums := make([]int, len(strings))
	for i, str := range strings {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("element [%d] %q is not a number: %w", i, str, err)
		}
		intNums[i] = num
	}
	return intNums, nil
}

func (r *SearchRepository) SearchIDs(ctx context.Context, query string) ([]int, error) {
	//отсечем пустые запросы
	words := tokenize(query)
	if len(words) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(words))
	for _, w := range words {
		keys = append(keys, "index:word:"+w)
	}

	rawIDs, err := r.client.SInter(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	idsForAns, err := stringsToInts(rawIDs)
	if err != nil {
		return nil, err
	}
	return idsForAns, nil
}

func buildCacheKey(query string) string {
	return "cache:search:" + strings.Join(tokenize(query), "_")
}

func (r *SearchRepository) SetCachedSearch(ctx context.Context, query string, movies []domain.Movie) error {
	key := buildCacheKey(query)
	data, err := json.Marshal(movies)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, time.Hour).Err()
}

func (r *SearchRepository) GetCachedSearch(ctx context.Context, query string) ([]domain.Movie, error) {
	key := buildCacheKey(query)
	data, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var movies []domain.Movie
	err = json.Unmarshal(data, &movies)
	if err != nil {
		return nil, err
	}
	return movies, nil
}
