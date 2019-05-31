package dratelimiter

import (
	"github.com/go-redis/redis"
	"sync"
	"testing"
	"time"
)


func createLimiter() (*Limiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	l := Limiter{}
	return &l, l.NewLimiter(client, 10)

}
func TestLimiter_NewLimiter(t *testing.T) {
	c, err := createLimiter()
	if err != nil {
		t.Error(err)
	}
	defer c.Done()
}
func TestLimiter_Serial(t *testing.T) {

	c, err := createLimiter()
	if err != nil {
		t.Error(err)
	}
	defer c.Done()



	for i := 0; i < 100; i++ {

		if !c.Allow() {
			t.Fatalf("Unexpected, rate limit was triggered")
		}
		time.Sleep(time.Millisecond * 100)

	}

	rateLimitRaised := false
	for i := 0; i < 100; i++ {

		if !c.Allow() {
			rateLimitRaised = true
		}
		time.Sleep(time.Millisecond * 98)

	}

	if !rateLimitRaised {
		t.Error("Unexpected, rate limit was not triggered")
	}

}
func TestLimiter_Parallel(t *testing.T) {
	c, err := createLimiter()
	if err != nil {
		t.Error(err)
	}
	defer c.Done()


	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func(w *sync.WaitGroup){
		for i := 0; i < 100; i++ {
			if !c.Allow() {
				w.Done()
				t.Fatalf("Unexpected, rate limit was triggered")
			}
			time.Sleep(time.Millisecond * 200)

		}
		w.Done()
	}(wg)

	go func(w *sync.WaitGroup){
		for i := 0; i < 100; i++ {
			if !c.Allow() {
				w.Done()
				t.Fatalf("Unexpected, rate limit was triggered")
			}
			time.Sleep(time.Millisecond * 200)

		}
		w.Done()
	}(wg)
	wg.Wait()



	wg.Add(2)
	triggered := false

	go func(w *sync.WaitGroup){
		for i := 0; i < 100; i++ {
			if !c.Allow() {
				triggered = true
				w.Done()
				return
			}
			time.Sleep(time.Millisecond * 198)

		}
		w.Done()
	}(wg)

	go func(w *sync.WaitGroup){
		for i := 0; i < 100; i++ {
			if !c.Allow() {
				triggered = true
				w.Done()
				return
			}
			time.Sleep(time.Millisecond * 198)
		}
		w.Done()
	}(wg)

	wg.Wait()


	if !triggered {
		t.Error("Unexpected, rate limit was not triggered")
	}


}
