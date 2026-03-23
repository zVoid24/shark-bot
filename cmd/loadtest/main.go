package main

import (
	"flag"
	"fmt"
	"log"
	"shark_bot/config"
	"shark_bot/infra/db"
	"shark_bot/infra/repository"
	"shark_bot/internal/activenumber"
	"shark_bot/internal/number"
	"shark_bot/internal/seennumber"
	"shark_bot/internal/user"
	"sync"
	"time"
)

func main() {
	numUsers := flag.Int("users", 100, "Number of concurrent users to simulate")
	platform := flag.String("platform", "test", "Platform to test")
	country := flag.String("country", "testing", "Country to test")
	seed := flag.Bool("seed", false, "Seed the database with test numbers before running")
	flag.Parse()

	fmt.Printf("🚀 Starting Load Test with %d users on %s/%s\n", *numUsers, *platform, *country)

	// 1. Load config
	cnf := config.Load()

	// 2. Connect to DB
	dbConn, err := db.NewConnection(&cnf.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	// 3. Setup Repositories
	userRepo := repository.NewUserRepo(dbConn)
	numberRepo := repository.NewNumberRepo(dbConn)
	activeRepo := repository.NewActiveNumberRepo(dbConn)
	seenRepo := repository.NewSeenNumberRepo(dbConn)

	// 4. Setup Services
	numberSvc := number.NewService(numberRepo)
	activeSvc := activenumber.NewService(activeRepo)
	seenSvc := seennumber.NewService(seenRepo)
	userSvc := user.NewService(userRepo)

	// Seeding logic
	if *seed {
		fmt.Printf("🌱 Seeding %d numbers for %s/%s...\n", *numUsers, *platform, *country)
		dummyNums := make([]string, *numUsers)
		for i := 0; i < *numUsers; i++ {
			dummyNums[i] = fmt.Sprintf("+1000%07d", i)
		}
		_, err := numberSvc.BulkInsert(*platform, *country, dummyNums)
		if err != nil {
			log.Fatal("Failed to seed numbers:", err)
		}
		fmt.Println("✅ Seeding complete.")
	}

	// Auto-detect platform/country if not provided (fallback)
	if *platform == "" {
		platforms, _ := numberSvc.GetPlatforms()
		if len(platforms) > 0 {
			*platform = platforms[0]
			countries, _ := numberSvc.GetCountries(*platform)
			if len(countries) > 0 {
				*country = countries[0]
			}
		}
	}

	// Pre-requisite check: Ensure there are numbers available
	count, _ := numberSvc.CountAvailable(*platform, *country)
	if count < *numUsers {
		fmt.Printf("⚠️  Warning: Only %d numbers available, but trying to simulate %d users. Some will fail.\n", count, *numUsers)
	}

	var wg sync.WaitGroup
	start := time.Now()

	results := make(chan error, *numUsers)

	for i := 0; i < *numUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			userIDStr := fmt.Sprintf("test_user_%d", userID)

			// Simulate IsBlocked check
			if blocked, _ := userSvc.IsBlocked(userIDStr); blocked {
				results <- fmt.Errorf("user %d is blocked", userID)
				return
			}

			// Simulate getNumbers
			numbers, err := numberSvc.GetNumbers(*platform, *country, userIDStr, nil, 1)
			if err != nil {
				results <- fmt.Errorf("user %d GetNumbers error: %v", userID, err)
				return
			}
			if len(numbers) == 0 {
				results <- fmt.Errorf("user %d: no numbers available", userID)
				return
			}

			// Simulate Insert Active
			num := numbers[0]
			an := activenumber.ActiveNumber{
				Number:    num,
				UserID:    userIDStr,
				Timestamp: time.Now(),
				MessageID: 12345, // dummy
				Platform:  *platform,
				Country:   *country,
			}
			if err := activeSvc.Insert(an); err != nil {
				results <- fmt.Errorf("user %d Insert error: %v", userID, err)
				return
			}

			// Simulate Seen
			if err := seenSvc.Add(userIDStr, num, *country); err != nil {
				results <- fmt.Errorf("user %d Seen error: %v", userID, err)
				return
			}

			results <- nil
		}(i)
	}

	wg.Wait()
	close(results)

	duration := time.Since(start)

	success := 0
	failed := 0
	for err := range results {
		if err != nil {
			failed++
		} else {
			success++
		}
	}

	fmt.Printf("\n🏁 Load Test Finished in %v\n", duration)
	fmt.Printf("✅ Successful: %d\n", success)
	fmt.Printf("❌ Failed:     %d\n", failed)
	fmt.Printf("📈 Avg Thrput: %.2f req/s\n", float64(*numUsers)/duration.Seconds())
}
