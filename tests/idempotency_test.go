package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/models"
	"github.com/barrukurniawan/multi_user/core/repositories"
	"github.com/barrukurniawan/multi_user/core/services"
	"github.com/google/uuid"
)

type mockTaskRepo struct {
	repositories.TaskRepository
	mu    sync.Mutex
	count int
}

func (m *mockTaskRepo) Create(ctx context.Context, task *models.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	task.ID = fmt.Sprintf("task-%d", m.count)
	return nil
}

type mockIdempotencyRepo struct {
	mu   sync.Mutex
	keys map[string]models.IdempotencyKey
}

func (m *mockIdempotencyRepo) Find(ctx context.Context, key string) (*models.IdempotencyKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.keys[key]
	if !ok {
		return nil, nil
	}
	return &k, nil
}

func (m *mockIdempotencyRepo) Reserve(ctx context.Context, key, userID string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.keys[key]; exists {
		return false, nil
	}
	m.keys[key] = models.IdempotencyKey{UserID: userID}
	return true, nil
}

func (m *mockIdempotencyRepo) SaveResponse(ctx context.Context, key, userID string, response []byte, statusCode int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[key] = models.IdempotencyKey{UserID: userID, StatusCode: statusCode, Response: response}
	return nil
}

func (m *mockIdempotencyRepo) Release(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, key)
	return nil
}

func newService() (services.TaskService, *mockTaskRepo) {
	tasks := &mockTaskRepo{}
	idem := &mockIdempotencyRepo{keys: map[string]models.IdempotencyKey{}}
	return services.NewTaskService(tasks, nil, idem, nil, nil, nil), tasks
}

func TestIdempotencySequential(t *testing.T) {
	svc, tasks := newService()
	key := uuid.NewString()
	req := dto.CreateTaskRequest{Title: "Belajar Go"}

	first, code1, err := svc.Create(context.Background(), "user-1", key, req)
	if err != nil || code1 != 201 {
		t.Fatalf("request pertama: code=%d err=%v, seharusnya 201", code1, err)
	}

	second, code2, err := svc.Create(context.Background(), "user-1", key, req)
	if err != nil || code2 != 201 {
		t.Fatalf("request kedua: code=%d err=%v, seharusnya 201", code2, err)
	}

	if tasks.count != 1 {
		t.Fatalf("task dibuat %d kali, seharusnya 1", tasks.count)
	}

	body1, _ := json.Marshal(first)
	body2, _ := json.Marshal(second)
	if string(body1) != string(body2) {
		t.Fatalf("response tidak identik:\n%s\n%s", body1, body2)
	}
}

func TestIdempotencyConcurrentDuplicate(t *testing.T) {
	svc, tasks := newService()
	key := uuid.NewString()
	const n = 50

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			svc.Create(context.Background(), "user-1", key, dto.CreateTaskRequest{Title: "Task kembar"})
		}()
	}

	close(start)
	wg.Wait()

	if tasks.count != 1 {
		t.Fatalf("task dibuat %d kali, seharusnya tepat 1", tasks.count)
	}
}
