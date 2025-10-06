package browser

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/playwright-community/playwright-go"
)

var jlog = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type BrowserManager struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
}

type BrowserOptions struct {
	Headless    bool
	BlockImages bool
	BlockMedia  bool
	BlockFonts  bool
	UserAgent   string
	Timeout     float64
}

type GejieConfig struct {
	BrowserHeadlessMode bool
	BrowserTimeout      float64
	BrowserType         BrowserType
}

func DefaultGejieConfig() *GejieConfig {
	return &GejieConfig{
		BrowserHeadlessMode: true,
		BrowserTimeout:      15000,
		BrowserType:         BrowserTypeChromium,
	}
}

var gejieConfig = DefaultGejieConfig()

func DefaultBrowserOptions() *BrowserOptions {
	return &BrowserOptions{
		Headless:    gejieConfig.BrowserHeadlessMode,
		BlockImages: true,
		BlockMedia:  true,
		BlockFonts:  true,
		UserAgent:   UserAgentChrome,
		// UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout: 15000,
	}
}

func CreateBrowser(pw *playwright.Playwright, opts *BrowserOptions, cmdOpts CmdOptions) (playwright.Browser, error) {
	var browser playwright.Browser
	var err error
	launchOpts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(opts.Headless),
		Timeout:  playwright.Float(opts.Timeout),
		Args: []string{
			string(disableBlinkFeaturesAutomationControlled),
			string(noSandbox),
		},
	}

	switch cmdOpts.Browser {
	case BrowserTypeFirefox:
		browser, err = pw.Firefox.Launch(launchOpts)
		if err != nil {
			return nil, err
		}
	case BrowserTypeChromium:
		if opts.Headless {
			// always use Firefox if headless mode since Chrome does not work headless
			slog.Info("running browser in Firefox headless mode")
			browser, err = pw.Firefox.Launch(launchOpts)
			if err != nil {
				return nil, err
			}
			// launchOpts.Args = append(launchOpts.Args, string("--headless=new"))
			// launchOpts.Args = append(launchOpts.Args, string("--disable-dev-shm-usage"))
		} else {
			browser, err = pw.Chromium.Launch(launchOpts)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("browser type not supported: %s", cmdOpts.Browser)
	}
	return browser, nil
}

type SupportedCommerceSite int

const (
	NotSupported SupportedCommerceSite = iota
	MercadoLibre
	DHGate
)

func (scs SupportedCommerceSite) String() string {
	notSupported := ""
	return [...]string{
		notSupported,
		"meli",
		"dhgate",
	}[scs]
}

func ParseCommerceSite(site string) SupportedCommerceSite {
	switch site {
	case "meli":
		return MercadoLibre
	case "dhgate":
		return DHGate
	default:
		return NotSupported
	}
}

type CmdOptions struct {
	Site         SupportedCommerceSite
	MaxItems     int
	OnlyImages   bool
	CreateCsv    bool
	HeadlessMode bool
	Workers      int
	Browser      BrowserType
}

func NewBrowserManager(opts *BrowserOptions, cmdOpts CmdOptions) (*BrowserManager, error) {
	if opts == nil {
		opts = DefaultBrowserOptions()
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	newBrowser, err := CreateBrowser(pw, opts, cmdOpts)
	if err != nil {
		pw.Stop()
		return nil, err
	}

	context, err := newBrowser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(opts.UserAgent),
	})
	if err != nil {
		newBrowser.Close()
		pw.Stop()
		return nil, err
	}

	context.On("request", func(req playwright.Request) {
		jlog.Info("on request information", "method", req.Method(), "url", req.URL(), "headers", req.Headers())
	})
	context.On("response", func(res playwright.Response) {
		jlog.Info("on response information", "response.status",
			res.Status(), "response.url", res.URL(), "resp.headers", res.Headers())
	})

	if opts.BlockImages || opts.BlockMedia || opts.BlockFonts {
		context.Route("**/*", func(route playwright.Route) {
			rt := route.Request().ResourceType()
			switch rt {
			case "image":
				if opts.BlockImages {
					route.Abort()
					return
				}
			case "media":
				if opts.BlockMedia {
					route.Abort()
					return
				}
			case "font":
				if opts.BlockFonts {
					route.Abort()
					return
				}
			}
			route.Continue()
		})
	}

	return &BrowserManager{
		pw:      pw,
		browser: newBrowser,
		context: context,
	}, nil
}

func (bm *BrowserManager) NewPage() (playwright.Page, error) {
	return bm.context.NewPage()
}

func (bm *BrowserManager) ClosePage(page playwright.Page) {
	if page != nil {
		page.Close()
	}
}

func (bm *BrowserManager) Close() {
	if bm.context != nil {
		bm.context.Close()
	}
	if bm.browser != nil {
		bm.browser.Close()
	}
	if bm.pw != nil {
		bm.pw.Stop()
	}
}

func (bm *BrowserManager) GetContext() playwright.BrowserContext {
	return bm.context
}

func (bm *BrowserManager) GetBrowser() playwright.Browser {
	return bm.browser
}
