package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"shark_bot/pkg/logger"
	"strings"
	"time"

	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var captPattern = regexp.MustCompile(`What is (\d+)\s*\+\s*(\d+)\s*=\s*\?\s*:`)
var ajaxSourcePattern = regexp.MustCompile(`"sAjaxSource":\s*"(res/data_smscdr\.php\?[^"]+)"`)

type Scraper struct {
	client   *http.Client
	loginURL string
	smsURL   string
	username string
	password string
	log      *slog.Logger
}

type SMSResult struct {
	DateTime    string
	Number      string
	Service     string
	Message     string
	OTPCode     string
	ShortCode   string
	FlagEmoji   string
	ServiceIcon string
}

func NewScraper(loginURL, smsURL, username, password string) (*Scraper, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Scraper{
		client:   &http.Client{Jar: jar, Timeout: 30 * time.Second},
		loginURL: loginURL,
		smsURL:   smsURL,
		username: username,
		password: password,
		log:      logger.New("scraper"),
	}, nil
}

func (s *Scraper) Login() error {
	if s.username == "" || s.password == "" {
		s.log.Error("login skipped: credentials missing", "username", s.username != "", "password", s.password != "")
		return fmt.Errorf("username or password not provided")
	}

	s.log.Info("attempting login", "url", s.loginURL, "user", s.username)

	// 1. Get login page to extract CAPTCHA
	resp, err := s.client.Get(s.loginURL)
	if err != nil {
		s.log.Error("failed to get login page", "err", err)
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		s.log.Error("failed to parse login page", "err", err)
		return err
	}

	// 2. Extract and solve CAPTCHA
	var captAnswer int
	foundCapt := false
	pageText := doc.Text()
	matches := captPattern.FindStringSubmatch(pageText)
	if len(matches) == 3 {
		n1, _ := strconv.Atoi(matches[1])
		n2, _ := strconv.Atoi(matches[2])
		captAnswer = n1 + n2
		foundCapt = true
		s.log.Info("solved CAPTCHA", "q", fmt.Sprintf("%d + %d", n1, n2), "ans", captAnswer)
	}

	if !foundCapt {
		s.log.Warn("CAPTCHA pattern not found in login page, attempting without it")
	}

	// 3. Post login credentials + CAPTCHA to the action URL
	// The form action is 'signin', so we join it with the base of loginURL
	base := s.loginURL
	if lastSlash := strings.LastIndex(base, "/"); lastSlash != -1 {
		base = base[:lastSlash+1]
	}
	signinURL := base + "signin"
	s.log.Info("posting to signin URL", "url", signinURL)

	data := url.Values{}
	data.Set("username", s.username)
	data.Set("password", s.password)
	if foundCapt {
		data.Set("capt", strconv.Itoa(captAnswer))
	}

	loginResp, err := s.client.PostForm(signinURL, data)
	if err != nil {
		s.log.Error("login post failed", "err", err)
		return err
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		s.log.Error("login failed status", "status", loginResp.Status)
		return fmt.Errorf("login failed with status: %d", loginResp.StatusCode)
	}

	// 4. Verify login by checking if "Logout" or specific dashboard elements exist
	// Or check final URL if it redirected
	finalURL := loginResp.Request.URL.String()
	s.log.Info("login post complete", "final_url", finalURL, "status", loginResp.Status)

	if strings.Contains(strings.ToLower(finalURL), "login") || strings.Contains(strings.ToLower(finalURL), "signin") {
		// If still on login/signin page, it likely failed
		s.log.Error("still on login/signin page after post, login likely failed", "url", finalURL)
		return fmt.Errorf("login failed: redirected back to login/signin page")
	}

	s.log.Info("login successful", "redirected_to", finalURL)
	return nil
}

func (s *Scraper) FetchSMS() ([]SMSResult, error) {
	s.log.Info("fetching SMS report page to extract AJAX source", "url", s.smsURL)

	// 1. Get the report page to find the current sAjaxSource (which includes a sesskey)
	resp, err := s.client.Get(s.smsURL)
	if err != nil {
		s.log.Error("failed to get report page", "err", err)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// 2. Extract sAjaxSource from JavaScript
	matches := ajaxSourcePattern.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		s.log.Warn("AJAX source not found in report page, table might be empty or session expired")
		return nil, nil
	}
	ajaxPath := matches[1]

	// 3. Construct full AJAX URL and encode properly
	baseURL := s.smsURL[:strings.LastIndex(s.smsURL, "/")+1]
	ajaxURL := baseURL + ajaxPath
	ajaxURL = strings.ReplaceAll(ajaxURL, " ", "%20")

	s.log.Info("polling AJAX source", "url", ajaxURL)

	// 4. Perform the AJAX request
	req, err := http.NewRequest("GET", ajaxURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", s.smsURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	ajaxResp, err := s.client.Do(req)
	if err != nil {
		s.log.Error("AJAX fetch failed", "err", err)
		return nil, err
	}
	defer ajaxResp.Body.Close()

	if ajaxResp.StatusCode != http.StatusOK {
		s.log.Error("AJAX fetch failed status", "status", ajaxResp.Status)
		return nil, fmt.Errorf("failed to fetch AJAX SMS: %d", ajaxResp.StatusCode)
	}

	// 5. Parse JSON Result
	var dtResult struct {
		AAData [][]interface{} `json:"aaData"`
	}
	if err := json.NewDecoder(ajaxResp.Body).Decode(&dtResult); err != nil {
		s.log.Error("failed to decode AJAX JSON", "err", err)
		return nil, err
	}

	const timeLayout = "2006-01-02 15:04:05"

	var results []SMSResult
	for _, row := range dtResult.AAData {
		if len(row) < 9 {
			continue
		}

		// Helper to safely get string from interface
		getStr := func(idx int) string {
			if idx >= len(row) || row[idx] == nil {
				return ""
			}
			switch v := row[idx].(type) {
			case string:
				return strings.TrimSpace(v)
			case float64:
				return fmt.Sprintf("%.0f", v)
			case int:
				return strconv.Itoa(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}

		dateTimeStr := getStr(0)
		number := getStr(2)
		serviceCol := getStr(3)
		message := getStr(5)

		if message == "" || number == "" || message == "0" || number == "0" || strings.Contains(dateTimeStr, "Total") {
			continue
		}


		results = append(results, SMSResult{
			DateTime: dateTimeStr,
			Number:   number,
			Service:  serviceCol,
			Message:  message,
		})
	}

	s.log.Info("fetch complete", "records_found", len(results))
	return results, nil
}
