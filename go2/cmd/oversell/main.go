// Bloque 3: la prueba de carga. Crea un producto nuevo con 10 unidades, lanza N
// POST /reserve simultáneos contra él, de una unidad cada uno, repartidos por
// turnos entre una o varias instancias de la API en marcha, y después lee el
// stock. Se esperan 10 creadas, el resto rechazadas con 409, y stock 0. Como
// el producto es nuevo en cada ejecución, se puede repetir sin reiniciar nada.
//
//	go run ./cmd/oversell -n 100 http://localhost:8080
//	go run ./cmd/oversell -n 100 http://localhost:8080 http://localhost:8081
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	n := flag.Int("n", 100, "number of concurrent reservations")
	stock := flag.Int("stock", 10, "stock of the product created for this run")
	product := flag.String("product", "", "reserve this existing product instead of creating one")
	flag.Parse()
	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "usage: oversell [-n 100] [-stock 10] [-product ID] URL [URL...]")
		os.Exit(2)
	}

	// Un producto nuevo por ejecución: así la carga no depende de lo que haya
	// quedado en la base de datos de la vez anterior.
	if *product == "" {
		*product = fmt.Sprintf("load-%d", time.Now().UnixMilli())
		form := url.Values{"name": {"Load test"}, "stock": {fmt.Sprint(*stock)}}
		resp, err := http.PostForm(urls[0]+"/products/"+*product, form)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			fmt.Fprintf(os.Stderr, "could not create the product: status %d: %s\n", resp.StatusCode, strings.TrimSpace(string(b)))
			os.Exit(1)
		}
		fmt.Printf("product %s created with stock %d\n", *product, *stock)
	}

	body := fmt.Sprintf(`{"product_id": %q, "quantity": 1}`, *product)
	counts := make(map[int]int)     // código de estado -> cuántas
	samples := make(map[int]string) // código de estado -> un cuerpo de ejemplo
	var mu sync.Mutex               // en counts y samples escriben muchas goroutines
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go func(base string) {
			defer wg.Done()
			<-start
			resp, err := http.Post(base+"/reserve", "application/json", strings.NewReader(body))
			code := 0 // 0 quiere decir que ha fallado la propia petición
			sample := ""
			if err == nil {
				b, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				code = resp.StatusCode
				sample = strings.TrimSpace(string(b))
			} else {
				sample = err.Error()
			}
			mu.Lock()
			counts[code]++
			if _, seen := samples[code]; !seen {
				samples[code] = sample
			}
			mu.Unlock()
		}(urls[i%len(urls)])
	}
	close(start)
	wg.Wait()

	fmt.Printf("%d requests: %d created (201), %d out of stock (409)", *n, counts[201], counts[409])
	fmt.Println()
	for code, c := range counts {
		if code != 201 && code != 409 {
			fmt.Printf("  %d with status %d, for example: %s\n", c, code, samples[code])
		}
	}

	resp, err := http.Get(urls[0] + "/products/" + *product)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	left, _ := io.ReadAll(resp.Body)
	fmt.Printf("stock now: %s", left)
}
