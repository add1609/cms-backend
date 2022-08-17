package src

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
)

// Set this value before serving anything.
var originHost string

func CheckOrigin(req *http.Request) bool {
	origin := req.Header["Origin"]
	if len(origin) == 0 {
		return false
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	return u.Host == originHost
}

func redircetHTTP(w http.ResponseWriter, req *http.Request) {
	if CheckOrigin(req) {
		host := strings.Split(req.Host, ":")[0]
		target := "https://" + strings.TrimPrefix(host, "www.") + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		log.Printf("[INFO] Redirecting client to HTTPS: %s, (%s), UserAgent: %s", target, host, req.UserAgent())
		http.Redirect(w, req, target, http.StatusPermanentRedirect)
	}
}

func SetupTLS(envDomain, envCacheDir, envEmail, envHTTP, envOriginHost string) {
	var (
		certReloader *simplecert.CertReloader
		numRenews    int
		ctx, cancel  = context.WithCancel(context.Background())
		tlsConf      = tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)
		makeServer   = func() *http.Server {
			return &http.Server{
				Addr:      ":443",
				TLSConfig: tlsConf,
			}
		}
		srv = makeServer()
		cfg = simplecert.Default
	)

	// configure
	cfg.Domains = []string{envDomain}
	cfg.CacheDir = envCacheDir
	cfg.SSLEmail = envEmail
	cfg.HTTPAddress = envHTTP

	// this function will be called just before certificate renewal starts and is used to gracefully stop the service
	// (we need to temporarily free port 443 in order to complete the TLS challenge)
	cfg.WillRenewCertificate = func() {
		// stop server
		cancel()
	}

	// this function will be called after the certificate has been renewed, and is used to restart your service.
	cfg.DidRenewCertificate = func() {
		numRenews++
		// restart server: both context and server instance need to be recreated!
		ctx, cancel = context.WithCancel(context.Background())
		srv = makeServer()
		// force reload the updated cert from disk
		certReloader.ReloadNow()
		// here we go again
		go serve(ctx, srv)
	}

	// init simplecert configuration
	// this will block initially until the certificate has been obtained for the first time.
	// on subsequent runs, simplecert will load the certificate from the cache directory on disk.
	certReloader, err := simplecert.Init(cfg, func() {
		os.Exit(0)
	})
	if err != nil {
		log.Fatalf("[FATAL ERROR] in simplecert.Init(): %v", err)
	}

	// redirect HTTP to HTTPS
	log.Printf("[INFO] Starting HTTP Listener on Port 80")
	go func() {
		err := http.ListenAndServe(":80", http.HandlerFunc(redircetHTTP))
		if err != nil {
			log.Printf("[ERROR] on redirect: %v", err)
		}
	}()

	// enable hot reload
	tlsConf.GetCertificate = certReloader.GetCertificateFunc()

	// start serving
	log.Printf("[INFO] Serving at: https://%s", cfg.Domains[0])
	serve(ctx, srv)

	log.Println("[INFO] Waiting forever")
	<-make(chan bool)
}

func serve(ctx context.Context, srv *http.Server) {

	// lets go
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL ERROR] ListenAndServeTLS(): %v", err)
		}
	}()

	log.Printf("[INFO] Server started")
	<-ctx.Done()
	log.Printf("[INFO] Server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	err := srv.Shutdown(ctxShutDown)
	if err == http.ErrServerClosed {
		log.Printf("[INFO] Server exited properly")
	} else if err != nil {
		log.Printf("[ERROR] Server encountered an error on exit: %v", err)
	}
}
