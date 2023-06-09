package main

import (
	"encoding/json"
	"fmt"
	"github.com/secim/src"
	"github.com/secim/src/client"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	c := client.From(client.NewHTTPClient())
	wg := sync.WaitGroup{}
	// klasorleri olustur
	if err := os.MkdirAll("cache/", 0o777); err != nil {
		log.Fatalf("Onbellek dizini olusturulamiyor! (cache/)")
	} else if err = os.MkdirAll("output/", 0o777); err != nil {
		log.Fatalf("Cikti dizini olusturulamiyor! (output/)")
	} else if err = os.MkdirAll("temp/", 0o777); err != nil {
		log.Fatalf("Temp dizini olusturulamiyor! (temp/)")
	}
	// cb ve mv icin ic / dis fetch paralel baslat
	for _, isCB := range []bool{false, true} {
		wg.Add(1)
		go icSandik(c, &wg, isCB)
		wg.Add(1)
		go disTemsSandik(c, &wg, isCB)
		wg.Add(1)
		go cezaeviSandik(c, &wg, isCB)
		wg.Add(1)
		go gumrukSandik(c, &wg, isCB)
	}
	// tum goroutine'leri bekle
	wg.Wait()
	fmt.Println("DONE.")
}

// region utils

// hata alirsak dogrudan programi kapatalim diye tembellik util'i
func must(_ int, err error) {
	if err != nil {
		log.Fatalf("cannot write to file: %v", err)
	}
}

// 8 = cumhurbaskanligi secimi, 9 = parlamento secimi
func secimTurID(isCB bool) int {
	if isCB {
		return 9
	}
	return 8
}

// timestamp'ler icin sabit konum: Europe/Istanbul (UTC+3)
// makinenin saati bozuk oldugu icin bunu enforce etmek gerekli
var loc = time.FixedZone("UTC+3", 3*60*60)

// ilgili csv dosyasini olustur, defer edilecek fonksiyonla beraber don
func openFile(title string, isCB bool) (io.Writer, func()) {
	prefix := "MV"
	if isCB {
		prefix = "CB"
	}
	// ornek: temp/sandiklarCB-14-05-2023-23-04.csv
	tm := time.Now().In(loc).Format("02-01-2006-15-04")
	fn := fmt.Sprintf("temp/%s%s-%s.csv", title, prefix, tm)
	lastFn := fmt.Sprintf("output/%s%s-%s.csv", title, prefix, tm)
	w, err := os.Create(fn)
	if err != nil {
		log.Fatalf("cannot open file: %v\n", err)
	}
	// dosya adini yazdir
	fmt.Printf("Dosya olusturuluyor: %s\n", fn)
	return w, func() {
		// close fonksiyonu
		if er := w.Close(); er != nil {
			log.Printf("cannot close file: %v\n", err)
		} else if er = os.Rename(fn, lastFn); er != nil {
			log.Printf("cannot move file to %s: %v\n", lastFn, err)
		}
	}
}

// csv icin string type'lari quote icine alma
func quoteVal(a any) (s string) {
	switch v := a.(type) {
	case bool:
		s = fmt.Sprintf("%v", v)
	case int:
		s = fmt.Sprintf("%d", v)
	case int32:
		s = fmt.Sprintf("%d", v)
	case int64:
		s = fmt.Sprintf("%d", v)
	case float32:
		s = fmt.Sprintf("%d", int(v))
	case float64:
		s = fmt.Sprintf("%d", int(v))
	default:
		s = fmt.Sprintf("%q", fmt.Sprintf("%v", a))
	}
	return
}

// turOrd sonuc turune gore siralamak icin order index verir
// (once ittifak, sonra parti, sonra bagimsiz sonuclar)
func turOrd(n string) int {
	if strings.HasPrefix(n, "ittifak") {
		return -3
	} else if strings.HasPrefix(n, "parti") {
		return -2
	} else if strings.HasPrefix(n, "bagimsiz") {
		return -1
	}
	// varsa ekstra alanlar
	return 0
}

func colNameBaslikMap(basliklar []src.SecimSonucBaslik, uniq bool) map[string]src.SecimSonucBaslik {
	m := make(map[string]src.SecimSonucBaslik)
	for _, baslik := range basliklar {
		if baslik.SiraNO == 0 {
			// backendde bazi adaylar kasten boyle skip edilmis
			continue
		}
		if uniq {
			if b, ok := m[baslik.ColumnNAME]; ok && (b.Ad != baslik.Ad) {
				log.Fatalf("sutun adi onceden girilmis! %s (%s, %s)",
					baslik.ColumnNAME, b.Ad, baslik.Ad)
			}
		}
		m[baslik.ColumnNAME] = baslik
	}
	return m
}

func adBaslikMap(basliklar []src.SecimSonucBaslik, uniq bool) map[string]src.SecimSonucBaslik {
	m := make(map[string]src.SecimSonucBaslik)
	for _, baslik := range basliklar {
		if baslik.SiraNO == 0 {
			// backendde bazi adaylar kasten boyle skip edilmis
			continue
		}
		if uniq {
			if b, ok := m[baslik.Ad]; ok && (b.ColumnNAME != baslik.ColumnNAME) {
				log.Fatalf("ad onceden girilmis! %s (%s, %s)",
					baslik.Ad, b.ColumnNAME, baslik.ColumnNAME)
			}
		}
		m[baslik.Ad] = baslik
	}
	return m
}

// adBaslikMap veya colNameBaslikMap ciktisi alabilir
func toOrdSutunlar(m map[string]src.SecimSonucBaslik) []src.SecimSonucBaslik {
	// degerleri listeye diz
	sutunlar := make([]src.SecimSonucBaslik, 0, len(m))
	for _, v := range m {
		sutunlar = append(sutunlar, v)
	}
	// listeyi sirala
	sort.Slice(sutunlar, func(i, j int) bool {
		l, r := sutunlar[i], sutunlar[j]
		// farkli turleri (ittifak, parti, bagimsiz) grupla
		if tl, tr := turOrd(l.ColumnNAME), turOrd(r.ColumnNAME); tl != tr {
			return tl < tr
		}
		// gruplari kendi icinde sira nosuna gore sirala
		if l.SiraNO != r.SiraNO {
			return l.SiraNO < r.SiraNO
		}
		// sira nolari eslesen olursa alfabetik sirala
		return l.Ad < r.Ad
	})
	return sutunlar
}

var units = []string{"B", "kB", "MB", "GB"}

func memUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	i, a := 0, float64(m.Alloc)
	for a >= 1024 && i < len(units) {
		i, a = i+1, a/1024
	}
	return fmt.Sprintf("%f %s", a, units[i])
}

// endregion utils

type SutunBilgi struct {
	Names map[string]src.SecimSonucBaslik `json:"names"`
}

type PrintCtx struct {
	ordCols        []src.SecimSonucBaslik
	skippedColumns map[int]bool
	i              int
}

func (sb *SutunBilgi) addRow(colNames map[string]src.SecimSonucBaslik, sonuc map[string]any) map[string]any {
	m := make(map[string]any)
	for colName, v := range sonuc {
		col, foundCol := colNames[colName]
		if !foundCol {
			// boyle bir sutun gorulmemis, uydurup sona ekle!
			col = src.SecimSonucBaslik{SiraNO: 9999, ColumnNAME: colName,
				Ad: strings.ToUpper(strings.ReplaceAll(colName, "_", " "))}
		}
		if _, foundName := sb.Names[col.Ad]; !foundName {
			sb.Names[col.Ad] = col
		}
		// map'e adlarla kaydet; col name ile degil!
		m[col.Ad] = v
	}
	return m
}

func (sb *SutunBilgi) FprintHeader(w io.Writer, isSkipColumn func(src.SecimSonucBaslik) bool) *PrintCtx {
	must(fmt.Fprint(w, "#"))
	pc := &PrintCtx{ordCols: toOrdSutunlar(sb.Names), skippedColumns: make(map[int]bool)}
	for i, sutun := range pc.ordCols {
		if isSkipColumn != nil && isSkipColumn(sutun) {
			pc.skippedColumns[i] = true
		} else {
			must(fmt.Fprintf(w, ",%q", sutun.Ad))
		}
	}
	must(fmt.Fprintln(w))
	return pc
}

func (pc *PrintCtx) FprintRow(w io.Writer, row map[string]any) {
	pc.i++
	must(fmt.Fprintf(w, "%d", pc.i))
	for j, sutun := range pc.ordCols {
		if pc.skippedColumns[j] {
			// skip this column
		} else if v, ok := row[sutun.Ad]; ok && v != nil {
			must(fmt.Fprintf(w, ",%s", quoteVal(v)))
		} else {
			must(fmt.Fprint(w, ","))
		}
	}
	must(fmt.Fprintln(w))
}

func skippedColumnsFn(isCB bool) func(src.SecimSonucBaslik) bool {
	if isCB {
		return nil
	}
	// mv secimleri icin bagimsiz adaylari skip et
	return func(baslik src.SecimSonucBaslik) bool {
		return strings.HasPrefix(baslik.ColumnNAME, "bagimsiz")
	}
}

func disTemsSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	defer wg.Done()
	st := secimTurID(isCB)

	// basliklari cek
	ulkeler := src.UlkeListesi(c)
	fmt.Printf("Yurt disi sandik basliklari cekiliyor (cb=%v) [%s]\n", isCB, memUsage())
	baslikList := src.YurtdisiSecimSonucBaslikListesi(c, st)
	// tek scope; tum column name'ler unique olmali
	colNames := colNameBaslikMap(baslikList, true)

	var sb SutunBilgi
	cacheFilename := fmt.Sprintf("cache/__disTemsSandiklar%d.cache", st)
	if getSutunBilgiFromCache(cacheFilename, &sb) {
		fmt.Printf("Yurt disi sandik sutun bilgileri onbellekten kullaniliyor (cb=%v) [%s]\n", isCB, memUsage())
	} else {
		// tek scope; tum adlar unique olmali
		sb = SutunBilgi{Names: adBaslikMap(baslikList, true)}
		fmt.Printf("Yurt disi sandik basliklari cekildi (cb=%v), %d sutun var [%s]\n",
			isCB, len(sb.Names), memUsage())
		for ulkeIdx, ulke := range ulkeler {
			fmt.Printf("Yurt disi sandik verileri cekiliyor (cb=%v) (%d / %d ulke) %s [%s]\n",
				isCB, ulkeIdx+1, len(ulkeler), ulke.UlkeADI, memUsage())
			for _, dt := range src.DisTemsilcilikListesi(c, ulke) {
				for _, sonuc := range src.SecimSandikSonucListesi(c, src.DisTemsSonucParams(dt, st)) {
					// tum row'lari fetch et
					sb.addRow(colNames, sonuc)
				}
			}
		}
		cacheSutunBilgi(cacheFilename, &sb)
	}

	// siralanmis basliklarla print
	w, closeFile := openFile("disTemsSandiklar", isCB)
	defer closeFile()
	pc := sb.FprintHeader(w, skippedColumnsFn(isCB))
	for ulkeIdx, ulke := range ulkeler {
		fmt.Printf("Yurt disi sandik verileri yaziliyor (cb=%v) (%d / %d ulke) %s [%s]\n",
			isCB, ulkeIdx+1, len(ulkeler), ulke.UlkeADI, memUsage())
		for _, dt := range src.DisTemsilcilikListesi(c, ulke) {
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.DisTemsSonucParams(dt, st)) {
				pc.FprintRow(w, sb.addRow(colNames, sonuc))
			}
		}
	}
	fmt.Printf("Yurt disi sandik verileri dosyaya yazildi (cb=%v) [%s].\n", isCB, memUsage())
}

func gumrukSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	defer wg.Done()
	st := secimTurID(isCB)

	// basliklari cek
	gumrukler := src.GumrukListesi(c)
	fmt.Printf("Gumruk sandik basliklari cekiliyor (cb=%v) [%s]\n", isCB, memUsage())
	baslikList := src.YurtdisiSecimSonucBaslikListesi(c, st)
	// tek scope; tum column name'ler unique olmali
	colNames := colNameBaslikMap(baslikList, true)

	var sb SutunBilgi
	cacheFilename := fmt.Sprintf("cache/__gumrukSandiklar%d.cache", st)
	if getSutunBilgiFromCache(cacheFilename, &sb) {
		fmt.Printf("Gumruk sandik sutun bilgileri onbellekten kullaniliyor (cb=%v) [%s]\n", isCB, memUsage())
	} else {
		// tek scope; tum adlar unique olmali
		sb = SutunBilgi{Names: adBaslikMap(baslikList, true)}
		fmt.Printf("Gumruk sandik basliklari cekildi (cb=%v), %d sutun var [%s]\n",
			isCB, len(sb.Names), memUsage())
		for gumrukIdx, gumruk := range gumrukler {
			fmt.Printf("Gumruk sandik verileri cekiliyor (cb=%v) (%d / %d gumruk) %s [%s]\n",
				isCB, gumrukIdx+1, len(gumrukler), gumruk.GumrukADI, memUsage())
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.GumrukSonucParams(gumruk, st)) {
				// tum row'lari fetch et
				sb.addRow(colNames, sonuc)
			}
		}
		cacheSutunBilgi(cacheFilename, &sb)
	}

	// siralanmis basliklarla print
	w, closeFile := openFile("gumrukSandiklar", isCB)
	defer closeFile()
	pc := sb.FprintHeader(w, skippedColumnsFn(isCB))
	for gumrukIdx, gumruk := range gumrukler {
		fmt.Printf("Gumruk sandik verileri yaziliyor (cb=%v) (%d / %d gumruk) %s [%s]\n",
			isCB, gumrukIdx+1, len(gumrukler), gumruk.GumrukADI, memUsage())
		for _, sonuc := range src.SecimSandikSonucListesi(c, src.GumrukSonucParams(gumruk, st)) {
			pc.FprintRow(w, sb.addRow(colNames, sonuc))
		}
	}
	fmt.Printf("Gumruk sandik verileri dosyaya yazildi (cb=%v) [%s].\n", isCB, memUsage())
}

func icSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	defer wg.Done()
	st := secimTurID(isCB)

	fmt.Printf("Yurt ici sandik basliklari cekiliyor (cb=%v) [%s]\n", isCB, memUsage())
	cevreler := src.IlListesi(c, st, 0)
	cevBas := make([][]src.SecimSonucBaslik, 0, len(cevreler))
	for _, cvr := range cevreler {
		cevBas = append(cevBas, src.SecimSonucBaslikListesi(c, cvr, st))
	}

	var sb SutunBilgi
	cacheFilename := fmt.Sprintf("cache/__yurticiSandiklar%d.cache", st)
	if getSutunBilgiFromCache(cacheFilename, &sb) {
		fmt.Printf("Yurt ici sandik sutun bilgileri onbellekten kullaniliyor (cb=%v) [%s]\n", isCB, memUsage())
	} else {
		// adBaslikMap tum basliklarin union'ını verir; uniq = false olmali
		var basTmp []src.SecimSonucBaslik
		for _, bas := range cevBas {
			basTmp = append(basTmp, bas...)
		}
		sb = SutunBilgi{Names: adBaslikMap(basTmp, false)}
		fmt.Printf("Yurt ici sandik basliklari cekildi (cb=%v), %d sutun var [%s]\n",
			isCB, len(sb.Names), memUsage())

		for cevIdx, cev := range cevreler {
			fmt.Printf("Yurt ici sandik verileri cekiliyor (cb=%v) (%d / %d secim cevresi) %s [%s]\n",
				isCB, cevIdx+1, len(cevreler), cev.IlADI, memUsage())
			cevColNameBaslikMap := colNameBaslikMap(cevBas[cevIdx], true)
			for _, ilce := range src.IlceListesi(c, cev, st, 0) {
				for _, sonuc := range src.SecimSandikSonucListesi(c, src.IlceSonucParams(ilce, st)) {
					// her cevrenin sonuclarini kendi column name'leriyle map'le
					sb.addRow(cevColNameBaslikMap, sonuc)
				}
			}
		}
		cacheSutunBilgi(cacheFilename, &sb)
	}

	// siralanmis basliklarla print
	w, closeFile := openFile("sandiklar", isCB)
	defer closeFile()
	pc := sb.FprintHeader(w, skippedColumnsFn(isCB))
	for cevIdx, cev := range cevreler {
		fmt.Printf("Yurt ici sandik verileri yaziliyor (cb=%v) (%d / %d secim cevresi) %s [%s]\n",
			isCB, cevIdx+1, len(cevreler), cev.IlADI, memUsage())
		cevColNameBaslikMap := colNameBaslikMap(cevBas[cevIdx], true)
		for _, ilce := range src.IlceListesi(c, cev, st, 0) {
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.IlceSonucParams(ilce, st)) {
				pc.FprintRow(w, sb.addRow(cevColNameBaslikMap, sonuc))
			}
		}
	}
	fmt.Printf("Yurt ici sandik verileri dosyaya yazildi (cb=%v).\n", isCB)
}

func cezaeviSandik(c client.Client, wg *sync.WaitGroup, isCB bool) {
	const cezaeviSandikTuru = 2
	defer wg.Done()
	st := secimTurID(isCB)

	fmt.Printf("Cezaevi sandik basliklari cekiliyor (cb=%v) [%s]\n", isCB, memUsage())
	cevreler := src.IlListesi(c, st, cezaeviSandikTuru)
	cevBas := make([][]src.SecimSonucBaslik, 0, len(cevreler))
	for _, cvr := range cevreler {
		cevBas = append(cevBas, src.SecimSonucBaslikListesi(c, cvr, st))
	}

	var sb SutunBilgi
	cacheFilename := fmt.Sprintf("cache/__cezaeviSandiklar%d.cache", st)
	if getSutunBilgiFromCache(cacheFilename, &sb) {
		fmt.Printf("Cezaevi sandik sutun bilgileri onbellekten kullaniliyor (cb=%v) [%s]\n", isCB, memUsage())
	} else {
		// adBaslikMap tum basliklarin union'ını verir; uniq = false olmali
		var basTmp []src.SecimSonucBaslik
		for _, bas := range cevBas {
			basTmp = append(basTmp, bas...)
		}
		sb = SutunBilgi{Names: adBaslikMap(basTmp, false)}
		fmt.Printf("Cezaevi sandik basliklari cekildi (cb=%v), %d sutun var [%s]\n",
			isCB, len(sb.Names), memUsage())

		for cevIdx, cev := range cevreler {
			fmt.Printf("Cezaevi sandik verileri cekiliyor (cb=%v) (%d / %d secim cevresi) %s [%s]\n",
				isCB, cevIdx+1, len(cevreler), cev.IlADI, memUsage())
			cevColNameBaslikMap := colNameBaslikMap(cevBas[cevIdx], true)
			for _, ilce := range src.IlceListesi(c, cev, st, cezaeviSandikTuru) {
				for _, sonuc := range src.SecimSandikSonucListesi(c, src.CezaeviSonucParams(ilce, st)) {
					// her cevrenin sonuclarini kendi column name'leriyle map'le
					sb.addRow(cevColNameBaslikMap, sonuc)
				}
			}
		}
		cacheSutunBilgi(cacheFilename, &sb)
	}

	// siralanmis basliklarla print
	w, closeFile := openFile("cezaeviSandiklar", isCB)
	defer closeFile()
	pc := sb.FprintHeader(w, skippedColumnsFn(isCB))
	for cevIdx, cev := range cevreler {
		fmt.Printf("Cezaevi sandik verileri yaziliyor (cb=%v) (%d / %d secim cevresi) %s [%s]\n",
			isCB, cevIdx+1, len(cevreler), cev.IlADI, memUsage())
		cevColNameBaslikMap := colNameBaslikMap(cevBas[cevIdx], true)
		for _, ilce := range src.IlceListesi(c, cev, st, cezaeviSandikTuru) {
			for _, sonuc := range src.SecimSandikSonucListesi(c, src.CezaeviSonucParams(ilce, st)) {
				pc.FprintRow(w, sb.addRow(cevColNameBaslikMap, sonuc))
			}
		}
	}
	fmt.Printf("Cezaevi sandik verileri dosyaya yazildi (cb=%v).\n", isCB)
}

func getSutunBilgiFromCache(fn string, sb *SutunBilgi) bool {
	b, err := os.ReadFile(fn)
	return err == nil && json.Unmarshal(b, &sb) == nil && len(sb.Names) != 0
}

func cacheSutunBilgi(fn string, sb *SutunBilgi) {
	if b, err := json.Marshal(sb); err != nil {
		log.Printf("WARN: cannot marshal sutunBilgi: %v\n", err)
	} else if err = os.WriteFile(fn, b, 0o666); err != nil {
		log.Printf("WARN: cannot write cache file %s: %v\n", fn, err)
	}
}
