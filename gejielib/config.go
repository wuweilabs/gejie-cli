package gejie

type BrowserType string

const (
	BrowserTypeChromium BrowserType = "chromium"
	BrowserTypeFirefox  BrowserType = "firefox"
)

type GejieConfig struct {
	BrowserHeadlessMode bool
	BrowserTimeout      float64
	BrowserType         BrowserType
}

func DefaultGejieConfig() *GejieConfig {
	return &GejieConfig{
		BrowserHeadlessMode: true,
		BrowserTimeout:      15000,
		BrowserType:         BrowserTypeFirefox,
	}
}

const UserAgentChrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"

// stealth browser options

type BrowserFlag string

const disableBlinkFeaturesAutomationControlled BrowserFlag = "--disable-blink-features=AutomationControlled"
const noSandbox BrowserFlag = "--no-sandbox"
