package deezer

import (
	"fmt"
	"log/slog"

	"github.com/playwright-community/playwright-go"
)

const (
	loginURL    = "https://account.deezer.com/en/login/"
	loggedInURL = "https://www.deezer.com/"
)

func fillAndSubmit(page playwright.Page, email, password string) error {
	if err := page.Locator(`input[name="email"], input[type="email"]`).Fill(email); err != nil {
		return fmt.Errorf("filling email: %w", err)
	}
	if err := page.Locator(`input[name="password"], input[type="password"]`).Fill(password); err != nil {
		return fmt.Errorf("filling password: %w", err)
	}
	if err := page.Locator(`button[type="submit"]`).Click(); err != nil {
		return fmt.Errorf("clicking submit: %w", err)
	}
	return nil
}

func Login(logger *slog.Logger, email, password string) (string, error) {
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("starting playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("launching browser: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		return "", fmt.Errorf("creating context: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return "", fmt.Errorf("creating page: %w", err)
	}

	logger.Info("loading login page...")
	if _, err := page.Goto(loginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return "", fmt.Errorf("navigating to login: %w", err)
	}

	// dismiss GDPR banner if present
	gdpr := page.Locator(`#gdpr-btn-accept-all, [data-testid="gdpr-btn-accept-all"]`)
	if err := gdpr.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		logger.Debug("no GDPR banner found", "error", err)
	}

	logger.Info("submitting credentials...")
	if err := fillAndSubmit(page, email, password); err != nil {
		return "", err
	}

	if err := page.WaitForURL(loggedInURL+"**", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		logger.Debug("initial redirect did not complete", "error", err)
	}

	if url := page.URL(); len(url) < len(loggedInURL) || url[:len(loggedInURL)] != loggedInURL {
		logger.Info("security check triggered, waiting up to 45s...")
		page.WaitForTimeout(40000)

		if err := page.Locator(`button[type="submit"]`).WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(10000),
		}); err != nil {
			return "", fmt.Errorf("waiting for submit button after security check: %w", err)
		}

		logger.Info("re-submitting credentials...")
		if err := fillAndSubmit(page, email, password); err != nil {
			return "", err
		}

		if err := page.WaitForURL(loggedInURL+"**", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(15000),
		}); err != nil {
			return "", fmt.Errorf("timed out after re-submit, current url: %s", page.URL())
		}
	}

	cookies, err := context.Cookies(loggedInURL)
	if err != nil {
		return "", fmt.Errorf("getting cookies: %w", err)
	}

	for _, c := range cookies {
		if c.Name == "arl" {
			logger.Info("arl cookie retrieved", "length", len(c.Value))
			return c.Value, nil
		}
	}

	return "", fmt.Errorf("no arl cookie found")
}
