package src

import (
	"context"
	"fmt"
	"github.com/secim/src/client"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"
)

func Get[T any](
	ctx context.Context, c client.Client, u string, m map[string]any,
) (t T, err error) {
	if !strings.HasPrefix(u, "https://") {
		u = "https://sspskokpit.ysk.gov.tr/api/ssps/" + strings.Trim(u, "/")
	}
	if len(m) != 0 {
		vals := url.Values{}
		for k, v := range m {
			vals[k] = []string{fmt.Sprintf("%v", v)}
		}
		u += "?" + vals.Encode()
	}
	err = c.Request(ctx, u, &t)
	return
}

func MustGet[T any](c client.Client, endpoint string, m map[string]any) T {
	ctx, cf := context.WithTimeout(context.Background(), time.Minute)
	defer cf()
	if m == nil {
		m = make(map[string]any)
	}
	t, err := Get[T](ctx, c, endpoint, m)
	if err != nil {
		log.Fatalf("failed req to %s: %v\n", endpoint, err)
	}
	return t
}

const (
	secimID = 60792
	//	secimTuru = 9 // 8 == mv, 9 = cb
)

// region IlListesi

func IlListesi(c client.Client, secimTuru, sandikTuru int) []Il {
	return MustGet[[]Il](c, "getIlList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": sandikTuru, "yurtIciDisi": 1,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getIlList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=0
//		&yurtIciDisi=1

type Il struct {
	SecimDETAYID        int    `json:"secim_DETAY_ID"`
	SecilecekADAYSAYISI int    `json:"secilecek_ADAY_SAYISI"`
	IlADI               string `json:"il_ADI"`
	IlID                int    `json:"il_ID"`
	SecimCEVRESIID      int    `json:"secim_CEVRESI_ID"`
	YerlesimYERITURU    int    `json:"yerlesim_YERI_TURU"`
	Id                  string `json:"id"`
}

// endregion
// region IlceListesi

func IlceListesi(c client.Client, i Il, secimTuru, sandikTuru int) []Ilce {
	return MustGet[[]Ilce](c, "getIlceList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": sandikTuru, "yurtIciDisi": 1,
		"ilId": i.IlID, "secimCevresiId": i.SecimCEVRESIID,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getIlceList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=0
//		&yurtIciDisi=1
//		&ilId=6
//		&secimCevresiId=404520

type Ilce struct {
	IlceID              int    `json:"ilce_ID"`
	IlceADI             string `json:"ilce_ADI"`
	BeldeID             int    `json:"belde_ID"`
	Id                  string `json:"id"`
	BirimID             int    `json:"birim_ID"`
	SecimDETAYID        int    `json:"secim_DETAY_ID"`
	SecilecekADAYSAYISI int    `json:"secilecek_ADAY_SAYISI"`
	IlADI               string `json:"il_ADI"`
	IlID                int    `json:"il_ID"`
	SecimCEVRESIID      int    `json:"secim_CEVRESI_ID"`
	YerlesimYERITURU    int    `json:"yerlesim_YERI_TURU"`
}

// endregion
// region MuhtarlikListesi

func MuhtarlikListesi(c client.Client, i Ilce, secimTuru, sandikTuru int) []Muh {
	return MustGet[[]Muh](c, "getMuhtarlikList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": sandikTuru, "yurtIciDisi": 1,
		"ilceId": i.IlceID, "beldeId": i.BeldeID, "birimId": i.BirimID, "secimCevresiId": i.SecimCEVRESIID,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getMuhtarlikList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=0
//		&yurtIciDisi=1
//		&ilceId=815
//		&beldeId=0
//		&birimId=3744
//		&secimCevresiId=404520

type Muh struct {
	MinSANDIKNO         string      `json:"min_SANDIK_NO"`
	MaxSANDIKNO         string      `json:"max_SANDIK_NO"`
	MuhtarlikID         int         `json:"muhtarlik_ID"`
	MuhtarlikADI        string      `json:"muhtarlik_ADI"`
	CezaeviID           int         `json:"cezaevi_ID"`
	IlceID              int         `json:"ilce_ID"`
	IlceADI             interface{} `json:"ilce_ADI"`
	BeldeID             int         `json:"belde_ID"`
	Id                  string      `json:"id"`
	BirimID             int         `json:"birim_ID"`
	SecimDETAYID        int         `json:"secim_DETAY_ID"`
	SecilecekADAYSAYISI int         `json:"secilecek_ADAY_SAYISI"`
	IlADI               interface{} `json:"il_ADI"`
	IlID                int         `json:"il_ID"`
	SecimCEVRESIID      int         `json:"secim_CEVRESI_ID"`
	YerlesimYERITURU    int         `json:"yerlesim_YERI_TURU"`
}

// endregion
// region GumrukListesi

func GumrukListesi(c client.Client) []Gumruk {
	return MustGet[[]Gumruk](c, "getGumrukList", map[string]any{
		"secimId": secimID,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getGumrukList
//		?secimId=60792

type Gumruk struct {
	GumrukID       int    `json:"gumruk_ID"`
	GumrukADI      string `json:"gumruk_ADI"`
	IlceID         int    `json:"ilce_ID"`
	MinSANDIKNO    string `json:"min_SANDIK_NO"`
	MaxSANDIKNO    string `json:"max_SANDIK_NO"`
	MaxSANDIKRUMUZ string `json:"max_SANDIK_RUMUZ"`
	MinSANDIKRUMUZ string `json:"min_SANDIK_RUMUZ"`
}

// endregion
// region UlkeListesi

func UlkeListesi(c client.Client) []Ulke {
	return MustGet[[]Ulke](c, "getUlkeList", map[string]any{
		"secimId": secimID,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getUlkeList
//		?secimId=60792

type Ulke struct {
	UlkeADI string `json:"ulke_ADI"`
	UlkeID  int    `json:"ulke_ID"`
}

// endregion
// region DisTemsilcilikListesi

func DisTemsilcilikListesi(c client.Client, u Ulke) []DisTemsilcilik {
	return MustGet[[]DisTemsilcilik](c, "getDisTemsilcilikList", map[string]any{
		"secimId": secimID, "ulkeId": u.UlkeID,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getDisTemsilcilikList
//		?secimId=60792
//		&ulkeId=9988

type DisTemsilcilik struct {
	MinSANDIKNO       string `json:"min_SANDIK_NO"`
	MaxSANDIKNO       string `json:"max_SANDIK_NO"`
	MaxSANDIKRUMUZ    string `json:"max_SANDIK_RUMUZ"`
	MinSANDIKRUMUZ    string `json:"min_SANDIK_RUMUZ"`
	DisTEMSILCILIKID  int    `json:"dis_TEMSILCILIK_ID"`
	DisTEMSILCILIKADI string `json:"dis_TEMSILCILIK_ADI"`
	UlkeADI           string `json:"ulke_ADI"`
	UlkeID            int    `json:"ulke_ID"`
}

// endregion

// region SecimSonucListesi

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSecimSonucList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=0
//		&yurtIciDisi=1
//		&ilId=6
//		&ilceId=815
//		&beldeId=0
//		&birimId=3744
//		&muhtarlikId=
//		&cezaeviId=
//		&sandikNoIlk=
//		&sandikNoSon=
//		&ulkeId=
//		&disTemsilcilikId=
//		&gumrukId=
//		&sandikRumuzIlk=
//		&sandikRumuzSon=
//		&secimCevresiId=404520
//		&sandikId=

func SecimSonucListesi(c client.Client, i Ilce, secimTuru int) []SecimSonuc {
	return MustGet[[]SecimSonuc](c, "getSecimSonucList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": 0, "yurtIciDisi": 1, "sandikId": "",
		"ilId": i.IlID, "ilceId": i.IlceID, "beldeId": i.BeldeID, "birimId": i.BirimID, "muhtarlikId": "",
		"cezaeviId": "", "sandikNoIlk": "", "sandikNoSon": "", "ulkeId": "", "disTemsilcilikId": "",
		"gumrukId": "", "sandikRumuzIlk": "", "sandikRumuzSon": "", "secimCevresiId": i.SecimCEVRESIID,
	})
}

type SecimSonuc struct {
	SecmenSAYISI                      int `json:"secmen_SAYISI"`
	ToplamSANDIKSAYISI                int `json:"toplam_SANDIK_SAYISI"`
	AcilanSANDIKSAYISI                int `json:"acilan_SANDIK_SAYISI"`
	AcilanSECMENSAYISI                int `json:"acilan_SECMEN_SAYISI"`
	OyKULLANANSECMENSAYISI            int `json:"oy_KULLANAN_SECMEN_SAYISI"`
	ItirazliGECERLIOYSAYISI           int `json:"itirazli_GECERLI_OY_SAYISI"`
	BirlestirmeTUTANAGITUMDUNYA       int `json:"birlestirme_TUTANAGI_TUMDUNYA"`
	BirlestirmeTUTANAGIDISTEMSILCILIK int `json:"birlestirme_TUTANAGI_DIS_TEMSILCILIK"`
	BirlestirmeTUTANAGIGUMRUKILCE     int `json:"birlestirme_TUTANAGI_GUMRUK_ILCE"`
	BirlestirmeTUTANAGIGUMRUKKURUL    int `json:"birlestirme_TUTANAGI_GUMRUK_KURUL"`
	BirlestirmeTUTANAGIGUMRUKRUMUZ    int `json:"birlestirme_TUTANAGI_GUMRUK_RUMUZ"`
	BirlestirmeTUTANAGIULKELER        int `json:"birlestirme_TUTANAGI_ULKELER"`
	BirlestirmeTUTANAGIGUMRUKLER      int `json:"birlestirme_TUTANAGI_GUMRUKLER"`
	ItirazsizGECERLIOYSAYISI          int `json:"itirazsiz_GECERLI_OY_SAYISI"`
	GecerliOYTOPLAMI                  int `json:"gecerli_OY_TOPLAMI"`
	GecersizOYTOPLAMI                 int `json:"gecersiz_OY_TOPLAMI"`
	BirlestirmeTUTANAGIIL             int `json:"birlestirme_TUTANAGI_IL"`
	BirlestirmeTUTANAGIILCE           int `json:"birlestirme_TUTANAGI_ILCE"`
	BirlestirmeTUTANAGIKURUL          int `json:"birlestirme_TUTANAGI_KURUL"`
	BirlestirmeTUTANAGIULKE           int `json:"birlestirme_TUTANAGI_ULKE"`
	SecilecekADAYSAYISI               int `json:"secilecek_ADAY_SAYISI"`
}

// endregion
// region SecimSandikSonucListesi

func GumrukSonucParams(g Gumruk, secimTuru int) map[string]any {
	return map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": 1, "yurtIciDisi": 2, "ulkeId": "",
		"gumrukId": g.GumrukID, "ilId": "", "ilceId": g.IlceID, "beldeId": "", "birimId": "",
		"muhtarlikId": "", "cezaeviId": "", "sandikNoIlk": "", "sandikNoSon": "", "disTemsilcilikId": "",
		"sandikRumuzIlk": "", "sandikRumuzSon": "", "secimCevresiId": "", "sandikId": "",
	}
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSecimSandikSonucList
//		?secimId=60792
//		&secimTuru=9
//		&sandikTuru=1
//		&yurtIciDisi=2
//		&ilId=
//		&ilceId=901
//		&beldeId=
//		&birimId=
//		&muhtarlikId=
//		&cezaeviId=
//		&sandikNoIlk=
//		&sandikNoSon=
//		&ulkeId=
//		&disTemsilcilikId=
//		&gumrukId=7
//		&sandikRumuzIlk=
//		&sandikRumuzSon=
//		&secimCevresiId=
//		&sandikId=

func CezaeviSonucParams(i Ilce, secimTuru int) map[string]any {
	return map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": 2, "yurtIciDisi": 1, "ulkeId": "",
		"disTemsilcilikId": "", "ilId": i.IlID, "ilceId": i.IlceID, "beldeId": i.BeldeID, "birimId": i.BirimID,
		"muhtarlikId": "", "cezaeviId": "", "sandikNoIlk": "", "sandikNoSon": "", "gumrukId": "",
		"sandikRumuzIlk": "", "sandikRumuzSon": "", "secimCevresiId": i.SecimCEVRESIID, "sandikId": "",
	}
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSecimSandikSonucList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=2
//		&yurtIciDisi=1
//		&ilId=1
//		&ilceId=473
//		&beldeId=0
//		&birimId=0
//		&muhtarlikId=
//		&cezaeviId=
//		&sandikNoIlk=
//		&sandikNoSon=
//		&ulkeId=
//		&disTemsilcilikId=
//		&gumrukId=
//		&sandikRumuzIlk=
//		&sandikRumuzSon=
//		&secimCevresiId=404480
//		&sandikId=

func DisTemsSonucParams(d DisTemsilcilik, secimTuru int) map[string]any {
	return map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": 3, "yurtIciDisi": 2, "ulkeId": d.UlkeID,
		"disTemsilcilikId": d.DisTEMSILCILIKID, "ilId": "", "ilceId": "", "beldeId": "", "birimId": "",
		"muhtarlikId": "", "cezaeviId": "", "sandikNoIlk": "", "sandikNoSon": "", "gumrukId": "",
		"sandikRumuzIlk": "", "sandikRumuzSon": "", "secimCevresiId": "", "sandikId": "",
	}
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSecimSandikSonucList
//		?secimId=60792
//		&secimTuru=9
//		&sandikTuru=3
//		&yurtIciDisi=2
//		&ilId=
//		&ilceId=
//		&beldeId=
//		&birimId=
//		&muhtarlikId=
//		&cezaeviId=
//		&sandikNoIlk=
//		&sandikNoSon=
//		&ulkeId=9893
//		&disTemsilcilikId=14
//		&gumrukId=
//		&sandikRumuzIlk=
//		&sandikRumuzSon=
//		&secimCevresiId=
//		&sandikId=

func IlceSonucParams(i Ilce, secimTuru int) map[string]any {
	return map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "sandikTuru": 0, "yurtIciDisi": 1, "sandikId": "",
		"ilId": i.IlID, "ilceId": i.IlceID, "beldeId": i.BeldeID, "birimId": i.BirimID, "muhtarlikId": "",
		"cezaeviId": "", "sandikNoIlk": "", "sandikNoSon": "", "ulkeId": "", "disTemsilcilikId": "",
		"gumrukId": "", "sandikRumuzIlk": "", "sandikRumuzSon": "", "secimCevresiId": i.SecimCEVRESIID,
	}
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSecimSandikSonucList
//		?secimId=60792
//		&secimTuru=8
//		&sandikTuru=0
//		&yurtIciDisi=1
//		&ilId=6
//		&ilceId=815
//		&beldeId=0
//		&birimId=3744
//		&muhtarlikId=
//		&cezaeviId=
//		&sandikNoIlk=
//		&sandikNoSon=
//		&ulkeId=
//		&disTemsilcilikId=
//		&gumrukId=
//		&sandikRumuzIlk=
//		&sandikRumuzSon=
//		&secimCevresiId=404520
//		&sandikId=

func SecimSandikSonucListesi(c client.Client, p map[string]any) []map[string]any {
	return MustGet[[]map[string]any](c, "getSecimSandikSonucList", p)
}

// endregion
// region SecimSonucBaslikListesi

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSandikSecimSonucBaslikList
//		?secimId=60792
//		&secimTuru=8
//		&yurtIciDisi=1
//		&secimCevresiId=404520
//		&ilId=6
//		&bagimsiz=1

func SecimSonucBaslikListesi(c client.Client, i Il, secimTuru int) []SecimSonucBaslik {
	return MustGet[[]SecimSonucBaslik](c, "getSandikSecimSonucBaslikList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "yurtIciDisi": 1,
		"secimCevresiId": i.SecimCEVRESIID, "ilId": i.IlID, "bagimsiz": 1,
	})
}

// https://sspskokpit.ysk.gov.tr/api/ssps/
//	getSandikSecimSonucBaslikList
//		?secimId=60792
//		&secimTuru=8
//		&yurtIciDisi=2
//		&secimCevresiId=
//		&ilId=
//		&bagimsiz=1

func YurtdisiSecimSonucBaslikListesi(c client.Client, secimTuru int) []SecimSonucBaslik {
	return MustGet[[]SecimSonucBaslik](c, "getSandikSecimSonucBaslikList", map[string]any{
		"secimId": secimID, "secimTuru": secimTuru, "yurtIciDisi": 2,
		"secimCevresiId": "", "ilId": "", "bagimsiz": 1,
	})
}

type SecimSonucBaslik struct {
	SiraNO     int    `json:"sira_NO"`
	Ad         string `json:"ad"`
	ColumnNAME string `json:"column_NAME"`
}

// endregion
// region MVSonucListesi

func GenelMVSonuclar(c client.Client) MVSonuc {
	return MustGet[MVSonuc](c,
		"https://sspskokpit.ysk.gov.tr/api/milletvekili/indexpagedata",
		map[string]any{"cacheSlayer": time.Now().UnixMilli()})
}

// https://sspskokpit.ysk.gov.tr/api/milletvekili/birim/SECIM_CEVRESI/404520?cacheSlayer=1684072407851

func CevreMVSonuclar(c client.Client, cevreID int) DVOData {
	dd := MustGet[DVOData](c, fmt.Sprintf(
		"https://sspskokpit.ysk.gov.tr/api/milletvekili/birim/SECIM_CEVRESI/%d", cevreID,
	), map[string]any{"cacheSlayer": time.Now().UnixMilli()})
	sort.Slice(dd.PartiDVOs, func(i, j int) bool {
		return dd.PartiDVOs[i].PartiSira < dd.PartiDVOs[j].PartiSira
	})
	return dd
}

type IttifakData struct {
	YURTDISIOY    int    `json:"YURTDISI_OY"`
	SIRA          int    `json:"SIRA"`
	CEVREOY       int    `json:"CEVRE_OY"`
	ITTIFAKUNVANI string `json:"ITTIFAK_UNVANI"`
	COLOR         string `json:"COLOR"`
	ULKEOY        int    `json:"ULKE_OY"`
	GUMRUKOY      int    `json:"GUMRUK_OY"`
}

type PartiDVOData struct {
	PartiAdi                 string `json:"parti_adi"`
	PartiKisaAdi             string `json:"parti_kisa_adi"`
	PartiSira                int    `json:"parti_sira"`
	Oy                       int    `json:"oy"`
	YurtdisindanYansiyacakOy int    `json:"yurtdisindanYansiyacakOy"`
	KazanacakMVSayisi        int    `json:"kazanacakMVSayisi"`
	PartiSecimId             int    `json:"parti_secim_id"`
	//	AdayDVOs                 []any  `json:"adayDVOs"`
	//	Color                    string `json:"color"`
}

type IttifakDVOData struct {
	IttifakUnvani string      `json:"ittifak_unvani"`
	IttifakSiraNo int         `json:"ittifak_sira_no"`
	IttifakId     int         `json:"ittifak_id"`
	Id            interface{} `json:"id"`
	Oy            int         `json:"oy"`
	Color         string      `json:"color"`
}

type DVOData struct {
	PartiDVOs                 []PartiDVOData   `json:"partiDVOs"`
	AltBirimDVOs              []DVOData        `json:"altBirimDVOs"`
	IttifakDVOs               []IttifakDVOData `json:"ittifakDVOs"`
	BagimsizDVOs              []any            `json:"bagimsizDVOs"`
	ToplamSandikSayisi        int              `json:"toplamSandikSayisi"`
	AcilanSandikSayisi        int              `json:"acilanSandikSayisi"`
	KayitliSecmenSayisi       int              `json:"kayitliSecmenSayisi"`
	AcilanKayitliSecmenSayisi int              `json:"acilanKayitliSecmenSayisi"`
	OyKullananSecmenSayisi    int              `json:"oyKullananSecmenSayisi"`
	GecerliOyToplami          int              `json:"gecerliOyToplami"`
	GecersizOyToplami         int              `json:"gecersizOyToplami"`
	SecimCevresiNo            int              `json:"secim_cevresi_no"`
	YerlesimYeriId            int              `json:"yerlesim_yeri_id"`
	SecilecekAdaySayisi       int              `json:"secilecek_aday_sayisi"`
	TarananTutanak            int              `json:"tarananTutanak"`
	TarananCetvel             int              `json:"tarananCetvel"`
	TutukluSecmenSayisi       int              `json:"tutukluSecmenSayisi"`
	Version                   string           `json:"version"`
	BirimID                   int              `json:"birim_ID"`
	BirimADI                  string           `json:"birim_ADI"`
	Turu                      string           `json:"turu"`
	UstBIRIM                  any              `json:"ust_BIRIM"`
	UstBIRIMTURU              any              `json:"ust_BIRIM_TURU"`
}

type MVSonuc struct {
	//	Ittifaklar []IttifakData `json:"ittifaklar"`
	Yurtdisi DVOData `json:"yurtdisi"`
	Turkiye  DVOData `json:"turkiye"`
	/*
		Partiler              []struct {
			PARTI        string `json:"PARTI"`
			YURTDISIOY   int    `json:"YURTDISI_OY"`
			SIRA         int    `json:"SIRA"`
			CEVREOY      int    `json:"CEVRE_OY"`
			COLOR        string `json:"COLOR"`
			CMVS         int    `json:"CMVS"`
			ULKEOY       int    `json:"ULKE_OY"`
			PartiKISAADI string `json:"parti_KISA_ADI"`
			GUMRUKOY     int    `json:"GUMRUK_OY"`
		} `json:"partiler"`
	*/
}

// endregion
