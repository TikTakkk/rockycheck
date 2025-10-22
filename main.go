package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const appName = "ROCKYCHECK (lite v1.5)"

// ⚠️ Placeholders pour publication GitHub (aucune clé active ici)
const (
	reverseIPAPI     = "https://api.example.com/reverseip?api_key=YOUR_KEY&ip={ip}"
	reverseDomainAPI = "https://api.example.com/reversedomain?api_key=YOUR_KEY&domain={domain}"
	contactLink      = "https://t.me/tomaMA212"
)

var (
	flagAuth = flag.Bool("auth", false, "Confirme que vous avez une autorisation explicite")
)

// ========= UI de base =========

func clearScreen() {
	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "cls").Run()
	} else {
		_ = exec.Command("clear").Run()
	}
}

func pause() {
	fmt.Print("\nAppuyez sur Entrée pour continuer...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func header() {
	clearScreen()
	fmt.Println("\033[1;34m  ____   ____  _   _ _  __  ____  _  __\033[0m")
	fmt.Println("\033[1;34m |  _ \\ / __ \\| \\ | (_)/ _|/ __ \\| |/ /\033[0m")
	fmt.Println("\033[1;34m | |_) | |  | |  \\| | | |_ | |  | | ' / \033[0m")
	fmt.Println("\033[1;34m |  _ <| |  | | . ` | |  _|| |  | |  <  \033[0m")
	fmt.Println("\033[1;34m |_| \\_\\____/|_|\\_ |_|_|   \\____/|_|\\_\\ \033[0m")
	fmt.Printf("\n\033[1;33m# %s - Reverse tools (IP->domain, domain->subdomains)\033[0m\n\n", appName)
	fmt.Println("⚠️  UTILISATION LÉGALE UNIQUEMENT — n'effectuez des scans que sur des cibles pour lesquelles vous avez une autorisation explicite.")
	fmt.Println()
}

// ========= Garde-fous & vérifs =========

// On considère l’API “ok” si 200 et JSON contenant des clés typiques
func checkAPI(url string) bool {
	testURL := strings.ReplaceAll(url, "{domain}", "google.com")
	testURL = strings.ReplaceAll(testURL, "{ip}", "8.8.8.8")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	content := string(body)
	return strings.Contains(content, `"result"`) || strings.Contains(content, `"domains"`)
}

func mustBeAuthorized() {
	envOK := strings.EqualFold(os.Getenv("I_AM_AUTHORIZED"), "true") || os.Getenv("I_AM_AUTHORIZED") == "1"
	if !*flagAuth && !envOK {
		fmt.Println("❌ I_AM_AUTHORIZED est False.")
		fmt.Println("   Lancez avec le flag --auth OU exportez I_AM_AUTHORIZED=true")
		fmt.Println("   Exemple :")
		fmt.Println("     - Windows : set I_AM_AUTHORIZED=true && main.exe")
		fmt.Println("     - Linux/Mac : I_AM_AUTHORIZED=true ./main")
		os.Exit(1)
	}
}

func menu() {
	header()

	// Vérification API (PLACEHOLDER -> affichera ton contact par design)
	fmt.Print("🔍 Vérification des APIs... ")
	if !checkAPI(reverseDomainAPI) || !checkAPI(reverseIPAPI) {
		fmt.Printf("\n\n❌ Aucune API valide détectée.\nVeuillez me contacter en MP : \033[1;36m%s\033[0m\n", contactLink)
		os.Exit(0)
	}
	fmt.Println("✅ OK")

	fmt.Println("\n\033[1;32m1)\033[0m Reverse IP -> domaine (API lookup)")
	fmt.Println("\033[1;32m2)\033[0m Reverse domaine -> sous-domaines (API)")
	fmt.Println("\033[1;31m0)\033[0m Quitter")
	fmt.Print("\nChoix: ")
}

// ========= Core: Reverse IP -> Domains =========

func reverseIPToDomain() {
	mustBeAuthorized()
	header()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Fichier d'IPs (ou Entrée pour coller manuellement): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	var ips []string
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Erreur lecture fichier:", err)
			pause()
			return
		}
		for _, l := range strings.Split(string(b), "\n") {
			l = strings.TrimSpace(l)
			if l != "" {
				ips = append(ips, l)
			}
		}
	} else {
		fmt.Println("Collez les IPs (une par ligne, vide pour finir):")
		for {
			txt, _ := reader.ReadString('\n')
			txt = strings.TrimSpace(txt)
			if txt == "" {
				break
			}
			ips = append(ips, txt)
		}
	}

	if len(ips) == 0 {
		fmt.Println("Aucune IP fournie.")
		pause()
		return
	}

	fmt.Printf("Nombre d'IPs: %d\n", len(ips))
	fmt.Print("Threads (défaut 50): ")
	th, _ := reader.ReadString('\n')
	th = strings.TrimSpace(th)
	threads := 50
	if v, err := strconv.Atoi(th); err == nil && v > 0 {
		threads = v
	}

	outFile := "iptodomains.txt"
	f, _ := os.Create(outFile)
	defer f.Close()

	client := &http.Client{Timeout: 30 * time.Second} // ⏱️ timeout augmenté
	jobs := make(chan string, len(ips))
	subCh := make(chan string, 1000)
	var processed, unique uint64
	start := time.Now()

	go progressBar(&processed, &unique, len(ips), start, "IPs", "domains")
	go writer(subCh, &unique, f)

	var wg sync.WaitGroup
	for w := 0; w < threads; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reDomain := regexp.MustCompile(`([a-z0-9\-]+\.)+[a-z]{2,}`)
			for ip := range jobs {
				url := strings.ReplaceAll(reverseIPAPI, "{ip}", ip)

				var body []byte
				// 🔁 retries + backoff
				for attempt := 1; attempt <= 3; attempt++ {
					req, _ := http.NewRequest("GET", url, nil)
					req.Header.Set("User-Agent", appName)
					resp, err := client.Do(req)
					if err != nil {
						time.Sleep(time.Duration(attempt) * 1 * time.Second)
						continue
					}
					body, _ = io.ReadAll(io.LimitReader(resp.Body, 400*1024))
					resp.Body.Close()
					if resp.StatusCode >= 500 {
						time.Sleep(time.Duration(attempt) * 1 * time.Second)
						continue
					}
					break
				}

				if len(body) > 0 {
					// Parse flexible: result.domains OU result (array) OU fallback regex
					var root map[string]interface{}
					if err := json.Unmarshal(body, &root); err == nil {
						// 1) result.domains (objet)
						if resObj, ok := root["result"].(map[string]interface{}); ok {
							if arr, ok := resObj["domains"].([]interface{}); ok {
								for _, v := range arr {
									if s, ok := v.(string); ok {
										subCh <- strings.TrimSpace(s)
									}
								}
							}
						}
						// 2) result []string (legacy)
						if arr, ok := root["result"].([]interface{}); ok {
							for _, v := range arr {
								if s, ok := v.(string); ok {
									subCh <- strings.TrimSpace(s)
								}
							}
						}
					} else {
						// 3) fallback regex
						for _, s := range reDomain.FindAllString(string(body), -1) {
							subCh <- s
						}
					}
				}

				atomic.AddUint64(&processed, 1)
			}
		}()
	}

	for _, ip := range ips {
		jobs <- ip
	}
	close(jobs)
	wg.Wait()
	close(subCh)

	fmt.Printf("\nTerminé. Résultats dans %s\n", outFile)
	pause()
}

// ========= Core: Reverse Domain -> Subdomains =========

func reverseDomainToSubdomains() {
	mustBeAuthorized()
	header()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Fichier de domaines (ou Entrée pour coller manuellement): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	var domains []string
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Erreur lecture fichier:", err)
			pause()
			return
		}
		for _, l := range strings.Split(string(b), "\n") {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			l = sanitizeDomain(l)
			domains = append(domains, l)
		}
	} else {
		fmt.Println("Collez les domaines (une par ligne, vide pour finir):")
		for {
			txt, _ := reader.ReadString('\n')
			txt = sanitizeDomain(strings.TrimSpace(txt))
			if txt == "" {
				break
			}
			domains = append(domains, txt)
		}
	}

	if len(domains) == 0 {
		fmt.Println("Aucun domaine fourni.")
		pause()
		return
	}

	fmt.Printf("Nombre de domaines: %d\n", len(domains))
	fmt.Print("Threads (défaut 50): ")
	th, _ := reader.ReadString('\n')
	th = strings.TrimSpace(th)
	threads := 50
	if v, err := strconv.Atoi(th); err == nil && v > 0 {
		threads = v
	}

	outFile := "domainstosubdomains.txt"
	f, _ := os.Create(outFile)
	defer f.Close()
	_ = os.MkdirAll("debug_responses", 0755)

	client := &http.Client{Timeout: 30 * time.Second}
	jobs := make(chan string, len(domains))
	subCh := make(chan string, 1000)
	var processed, unique uint64
	start := time.Now()

	go progressBar(&processed, &unique, len(domains), start, "Domains", "subdomains")
	go writer(subCh, &unique, f)

	var wg sync.WaitGroup
	for w := 0; w < threads; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reDomain := regexp.MustCompile(`([a-z0-9\-]+\.)+[a-z]{2,}`)
			for d := range jobs {
				url := strings.ReplaceAll(reverseDomainAPI, "{domain}", d)

				var body []byte
				for attempt := 1; attempt <= 3; attempt++ {
					req, _ := http.NewRequest("GET", url, nil)
					req.Header.Set("User-Agent", appName)
					resp, err := client.Do(req)
					if err != nil {
						time.Sleep(time.Duration(attempt) * 1 * time.Second)
						continue
					}
					body, _ = io.ReadAll(io.LimitReader(resp.Body, 800*1024))
					resp.Body.Close()
					if resp.StatusCode >= 500 {
						time.Sleep(time.Duration(attempt) * 1 * time.Second)
						continue
					}
					break
				}

				if len(body) > 0 {
					var parsed map[string]interface{}
					if err := json.Unmarshal(body, &parsed); err == nil {
						if res, ok := parsed["result"].(map[string]interface{}); ok {
							if arr, ok := res["domains"].([]interface{}); ok {
								for _, v := range arr {
									if s, ok := v.(string); ok {
										subCh <- strings.TrimSpace(s)
									}
								}
							}
						}
					} else {
						found := reDomain.FindAllString(string(body), -1)
						for _, s := range found {
							subCh <- s
						}
						if len(found) == 0 {
							_ = os.WriteFile("debug_responses/"+d+".txt", body, 0644)
						}
					}
				}

				atomic.AddUint64(&processed, 1)
			}
		}()
	}

	for _, d := range domains {
		jobs <- d
	}
	close(jobs)
	wg.Wait()
	close(subCh)

	fmt.Printf("\nTerminé. Résultats dans %s\n", outFile)
	pause()
}

// ========= Helpers =========

func sanitizeDomain(s string) string {
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(s, "http://"), "https://"))
	return strings.TrimRight(s, "/")
}

func progressBar(proc, uniq *uint64, total int, start time.Time, label, label2 string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	var last uint64
	for range ticker.C {
		p := atomic.LoadUint64(proc)
		u := atomic.LoadUint64(uniq)
		elapsed := time.Since(start)
		delta := p - last
		last = p
		instant := float64(delta) * 60
		avg := float64(p) / max(0.001, elapsed.Minutes())
		fmt.Printf("%s: %d/%d | Unique %s: %d | CPM: %.0f (inst) / %.0f (avg)\r",
			label, p, total, label2, u, instant, avg)
		if p >= uint64(total) {
			return
		}
	}
}

func writer(ch chan string, uniq *uint64, f *os.File) {
	seen := make(map[string]struct{})
	for s := range ch {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		_, _ = f.WriteString(s + "\n")
		atomic.AddUint64(uniq, 1)
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ========= MAIN =========

func main() {
	flag.Parse()
	for {
		menu()
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			reverseIPToDomain()
		case "2":
			reverseDomainToSubdomains()
		case "0":
			fmt.Println("Au revoir.")
			return
		default:
			fmt.Println("Choix invalide.")
			pause()
		}
	}
}
