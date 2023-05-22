package main

import (
	"fmt"
	"github.com/secim/src"
	"github.com/secim/src/client"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func turOrd(n string) int {
	if strings.HasPrefix(n, "ittifak") {
		return -2
	} else if strings.HasPrefix(n, "parti") {
		return -1
	}
	return 0
}

func must(_ int, err error) {
	if err != nil {
		log.Fatalf("cannot write to file: %v", err)
	}
}

func secimTurID(isCB bool) int {
	if isCB {
		return 9
	}
	return 8
}

func main() {
	c := client.From(client.NewHTTPClient())
	wg := sync.WaitGroup{}
	for _, isCB := range []bool{false, true} {
		wg.Add(1)
		go icSandik(c, &wg, isCB)
		wg.Add(1)
		go disTemsSandik(c, &wg, isCB)
	}
	wg.Wait()
	fmt.Println("DONE.")
}

var loc = time.FixedZone("UTC+3", 3*60*60)

//goland:noinspection GoBoolExpressions
func disTemsSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	defer wg.Done()

	w, cls := openFile("disTemsSandiklar", isCB)
	defer cls()

	st := secimTurID(isCB)

	partiAdlari := make(map[string]src.SecimSonucBaslik)
	for _, ssb := range src.YurtdisiSecimSonucBaslikListesi(c, st) {
		if strings.HasSuffix(ssb.ColumnNAME, "_ALDIGI_OY") &&
			(isCB || !strings.HasPrefix(ssb.ColumnNAME, "bagimsiz")) {
			partiAdlari[ssb.ColumnNAME] = ssb
		}
	}
	var sutunlar []src.SecimSonucBaslik
	for _, v := range sutunlar {
		sutunlar = append(sutunlar, v)
	}
	sort.Slice(sutunlar, func(i, j int) bool {
		return sutunlar[i].SiraNO < sutunlar[j].SiraNO
	})

	must(fmt.Fprint(w, "ULKE,KONSOLOSLUK,\"SANDIK NO\",\"SANDIK RUMUZ\""))
	for _, parti := range sutunlar {
		must(fmt.Fprintf(w, ",%q", parti.Ad))
	}
	must(fmt.Fprintln(w))

	for _, ulke := range src.UlkeListesi(c) {
		for _, dt := range src.DisTemsilcilikListesi(c, ulke) {
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.DisTemsSonucParams(dt, st)) {

				oylar := make(map[string]int)
				for k, v := range sonuc {
					if partiBaslik, ok := partiAdlari[k]; ok {
						oylar[partiBaslik.Ad] = convInt(v)
					}
				}

				must(fmt.Fprintf(w, "%s,%s,%s,%s",
					sonuc["il_ADI"], sonuc["ilce_ADI"], sonuc["sandik_NO"], sonuc["sandik_RUMUZ"]))
				for _, parti := range sutunlar {
					if oy, ok := oylar[parti.Ad]; ok {
						must(fmt.Fprintf(w, ",%d", oy))
					} else {
						must(fmt.Fprint(w, ","))
					}
				}
				must(fmt.Fprintln(w))
			}

		}
	}

}

//goland:noinspection GoBoolExpressions
func icSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	defer wg.Done()

	w, cls := openFile("sandiklar", isCB)
	defer cls()

	st := secimTurID(isCB)
	cevreler := src.IlListesi(c, st)

	// m: parti adindan veriye map
	cevreBasliklari := make(map[int][]src.SecimSonucBaslik)
	partiAdlari := make(map[string]int)
	for _, cvr := range cevreler {
		bas := src.SecimSonucBaslikListesi(c, cvr, st)
		cevreBasliklari[cvr.SecimCEVRESIID] = bas
		for _, ssb := range bas {
			if strings.HasSuffix(ssb.ColumnNAME, "_ALDIGI_OY") &&
				(isCB || !strings.HasPrefix(ssb.ColumnNAME, "bagimsiz")) {
				partiAdlari[ssb.Ad] = turOrd(ssb.ColumnNAME)
			}
		}
	}
	sutunlar := make([]string, 0, len(partiAdlari))
	for k := range partiAdlari {
		sutunlar = append(sutunlar, k)
	}
	sort.Slice(sutunlar, func(i, j int) bool {
		l, r := sutunlar[i], sutunlar[j]
		// farkli turleri (ittifak, parti, bagimsiz) grupla
		if tl, tr := partiAdlari[l], partiAdlari[r]; tl != tr {
			return tl < tr
		}
		// her birini kendi icinde alfabetik sirala
		return l < r
	})

	// baslik satiri yaz
	must(fmt.Fprint(w, "ZAMAN,IL,ILCE,MUHTARLIK,SANDIK"))
	for _, sutun := range sutunlar {
		must(fmt.Fprintf(w, ",%q", sutun))
	}
	must(fmt.Fprintln(w))

	for i, cvr := range cevreler {

		fmt.Printf("- %s (%d/%d)\n", cvr.IlADI, 1+i, len(cevreler))

		// sutun adindan baslik bilgisine
		m := make(map[string]src.SecimSonucBaslik)
		for _, ssb := range cevreBasliklari[cvr.SecimCEVRESIID] {
			m[ssb.ColumnNAME] = ssb
		}

		// her ilce icin
		ilceler := src.IlceListesi(c, cvr, st)
		for j, ilce := range ilceler {

			fmt.Printf("  - %s (%d/%d)\n", ilce.IlceADI, 1+j, len(ilceler))

			// satirlar
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.IlceSonucParams(ilce, st)) {

				// bu satirdaki oy sutunlari
				oylar := make(map[string]int)
				for k, v := range sonuc {
					if partiBaslik, ok := m[k]; ok {
						oylar[partiBaslik.Ad] = convInt(v)
					}
				}

				// sira nosuna gore basliklari iterate et
				must(fmt.Fprintf(w, "%q,%q,%q,%q,%v", time.Now().In(loc).Format(time.DateTime),
					ilce.IlADI, ilce.IlceADI, sonuc["muhtarlik_ADI"], sonuc["sandik_NO"]))
				for _, ssb := range sutunlar {
					if oy, ok := oylar[ssb]; ok {
						must(fmt.Fprintf(w, ",%d", oy))
					} else {
						must(fmt.Fprint(w, ","))
					}
				}
				must(fmt.Fprintln(w))
			}
		}
	}
}

func openFile(title string, isCB bool) (io.Writer, func()) {
	prefix := "MV"
	if isCB {
		prefix = "CB"
	}
	fn := fmt.Sprintf("%s%s-%s.csv", title, prefix,
		time.Now().In(loc).Format("02-01-2006-15-04"))
	w, err := os.Create(fn)
	if err != nil {
		log.Fatalf("cannot open file: %v\n", err)
	}
	fmt.Printf("[%s] FILE: %s\n", prefix, fn)
	return w, func() {
		if er := w.Close(); er != nil {
			log.Printf("cannot close file: %v\n", err)
		}
	}
}

func convInt(a any) int {
	switch v := a.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		log.Fatalf("wtf %T %v", a, a)
		return 0
	}
}
