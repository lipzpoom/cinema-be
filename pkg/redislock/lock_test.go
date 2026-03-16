package redislock_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gin-quickstart/pkg/redislock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// ข้อมูลเชื่อมต่อ Redis สำหรับเทส แนะนำให้รัน redis:6379 จริงๆ ทิ้งไว้
func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // เปลี่ยนเป็น URI รูปแบบที่คุณใช้งานอยู่
	})
}

func TestConcurrentAcquireMultipleLocks(t *testing.T) {
	client := setupRedis()
	defer client.Close()

	// ล้าง cache เพื่อความเคลียร์ก่อนรัน
	client.FlushDB(context.Background())

	locker := redislock.NewRedisLocker(client)

	ctx := context.Background()
	prefix := "test_theater"
	seatsToLock := []string{"A1", "A2"}

	// ตั้งค่าจำนวนคนที่กดยิงเข้ามาพร้อมๆ กัน
	concurrentUsers := 10

	var wg sync.WaitGroup
	var successCount int
	var faillCount int
	var mu sync.Mutex

	// จำลองคน 10 คนกดซื้อที่นั่ง A1, A2 พร้อมกัน
	wg.Add(concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			// รันคนละ Go Routine
			success, _, err := locker.AcquireMultipleLocks(ctx, prefix, seatsToLock, 1*time.Minute)

			mu.Lock()
			if success && err == nil {
				successCount++
				fmt.Printf("User %d successfully acquired locks!\n", userID)
			} else {
				faillCount++
				fmt.Printf("User %d failed to acquire locks (already locked)\n", userID)
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// ASSERTIONS: ต้องมีคนเดียวเท่านั้นที่ Success (Lock ผ่าน)
	// คนที่เหลืออีก 9 คน ต้อง Fail (Lock ไม่ผ่าน)
	assert.Equal(t, 1, successCount, "Only 1 user should be able to acquire the lock")
	assert.Equal(t, concurrentUsers-1, faillCount, "Other users should fail to acquire the lock")

	// หลังทดสอบเสร็จ ลอง Release ดูครับ
	success, _, _ := locker.AcquireMultipleLocks(ctx, prefix, seatsToLock, 1*time.Minute)
	assert.False(t, success, "Should still be locked before release")
}
