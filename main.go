package main

import (
    "encoding/json"
    "html/template"
    "log"
    "net"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
    "time"

    "golang.org/x/net/idna"
)

/*
  Forsale Web
  -----------
  Volledige webserver die _for-sale TXT-records controleert conform draft-davids-forsalereg-18.

  Belangrijkste ontwerpkeuzes:
  - Veilig: geen auto-redirect; links vragen bevestiging (frontend).
  - Net.LookupTXT voor DNS; ondersteunt IDN via golang.org/x/net/idna.
  - Robust parsing van "v=FORSALE1;" + één content-tag per record.
  - Presentatie voor niet-technische gebruikers; moderne UI.
*/

// ---------- Datatypen ----------

type ValidatedURI struct {
    URI    string `json:"uri"`
    Scheme string `json:"scheme"`
    Valid  bool   `json:"valid"`
    Note   string `json:"note"`
}

type Price struct {
    Currency      string  `json:"currency"`
    AmountString  string  `json:"amountString"`
    AmountFloat   float64 `json:"amountFloat,omitempty"`
    FormattedNice string  `json:"formattedNice"`
}

type SaleInfo struct {
    DomainInput string `json:"domainInput"`
    Unicode     string `json:"unicode"`
    Punycode    string `json:"punycode"`

    ForSale bool     `json:"forSale"`
    Reasons []string `json:"dismissReasons"`

    FTxt  []string       `json:"ftxt"`
    FUri  []ValidatedURI `json:"furi"`
    FVal  []Price        `json:"fval"`
    FCod  []string       `json:"fcod"`
    RawRR []string       `json:"rawRR"`

    Warnings []string `json:"warnings"`
}

// ---------- Helpers ----------

func isLikelyDomain(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" {
        return false
    }
    // simpele check: minimaal één punt, geen spaties of protocol
    if strings.ContainsAny(s, " /\\") {
        return false
    }
    return strings.Contains(s, ".")
}

func toASCII(domain string) (string, error) {
    return idna.Lookup.ToASCII(strings.TrimSpace(domain))
}
func toUnicode(domain string) (string, error) {
    return idna.Lookup.ToUnicode(strings.TrimSpace(domain))
}

func sanitizeText(s string) string {
    s = strings.ReplaceAll(s, "\t", " ")
    s = strings.ReplaceAll(s, "\n", " ")
    s = strings.ReplaceAll(s, "\r", " ")
    return strings.TrimSpace(s)
}

var reFval = regexp.MustCompile(`^([A-Z]+)(\d+(?:\.\d+)?)$`)

func formatPrice(cur, amt string) Price {
    p := Price{Currency: cur, AmountString: amt, FormattedNice: cur + " " + amt}
    symbol := map[string]string{
        "EUR": "€", "USD": "$", "GBP": "£", "JPY": "¥",
        "CHF": "CHF", "AUD": "A$", "CAD": "C$", "CNY": "¥", "INR": "₹",
    }
    if sym, ok := symbol[cur]; ok {
        p.FormattedNice = sym + " " + amt
    }
    return p
}

func validateURI(u string) ValidatedURI {
    v := ValidatedURI{URI: u, Valid: false, Scheme: "", Note: ""}

//    // spaties zijn niet toegestaan in URIs (moeten percent-encoded zijn)
//    if strings.Contains(u, " ") {
//        v.Note = "Bevat spaties; URI zou percent-encoded moeten zijn (%20)."
//        return v
//    }
    parsed, err := url.Parse(u)
    if err != nil {
        v.Note = "Kon URI niet parsen"
        return v
    }
    if parsed.Scheme == "" {
        v.Note = "(URI ontbreekt of mist scheme http/https/mailto/tel)"
        return v
    }
    allowed := map[string]bool{"http": true, "https": true, "mailto": true, "tel": true}
    if !allowed[strings.ToLower(parsed.Scheme)] {
        v.Note = "(Niet-ondersteunde URI-scheme)"
        return v
    }
    v.Valid = true
    v.Scheme = strings.ToLower(parsed.Scheme)
    return v
}

// verplichte versie-tag aan het begin
func hasVersionPrefix(rr string) (bool, string) {
    const tag = "v=FORSALE1;"
    rr = strings.TrimSpace(rr)
    if !strings.HasPrefix(rr, tag) {
        return false, ""
    }
    rest := rr[len(tag):]
    // Robustness (§3.6): tolereer spaties direct na de ;
    rest = strings.TrimLeft(rest, " ")
    return true, rest
}

// parse één TXT-record (na versie-tag) voor de content-tag
func parseForsaleRR(content string, info *SaleInfo) {
    content = strings.TrimSpace(content)
    if content == "" {
        info.ForSale = true
        info.Reasons = append(info.Reasons, "Leeg 'te koop' record (alleen versie-tag) aangetroffen.") // Geldige versie-tag zonder inhoud (SHOULD be te koop)
        return
    }

    switch {
    case strings.HasPrefix(content, "ftxt="):
        val := sanitizeText(content[len("ftxt="):])
        if val != "" {
            info.FTxt = append(info.FTxt, val)
            info.ForSale = true
        } else {
            info.Reasons = append(info.Reasons, "Lege ftxt=-waarde")
            info.ForSale = true // versie was geldig; nog steeds te koop (SHOULD)
        }
    case strings.HasPrefix(content, "furi="):
        val := strings.TrimSpace(content[len("furi="):])
        v := validateURI(val)
        info.FUri = append(info.FUri, v)
        info.ForSale = true
        if !v.Valid {
            info.Warnings = append(info.Warnings, "Aangetroffen furi is mogelijk ongeldig: "+val+" "+v.Note)
        }
    case strings.HasPrefix(content, "fval="):
        val := strings.TrimSpace(content[len("fval="):])
        m := reFval.FindStringSubmatch(val)
        if len(m) == 3 {
            price := formatPrice(m[1], m[2])
            info.FVal = append(info.FVal, price)
            info.ForSale = true
        } else {
            info.Reasons = append(info.Reasons, "Ongeldige fval-structuur; verwacht CUR123[.45]: "+val)
            info.ForSale = true
        }
    case strings.HasPrefix(content, "fcod="):
        val := sanitizeText(content[len("fcod="):])
        if val != "" {
            info.FCod = append(info.FCod, val)
            info.ForSale = true
        } else {
            info.Reasons = append(info.Reasons, "Lege fcod=-waarde")
            info.ForSale = true
        }
    default:
        // Niet-herkende content-tag; versie was geldig -> te koop (SHOULD).
        info.Reasons = append(info.Reasons, "Onbekende of ongeldige content-tag: "+content)
        info.ForSale = true
    }
}

// ---------- Kernlogica ----------

func checkDomain(input string) SaleInfo {
    info := SaleInfo{
        DomainInput: strings.TrimSpace(input),
        Reasons:     []string{},
        Warnings:    []string{},
    }

    if !isLikelyDomain(info.DomainInput) {
        info.Warnings = append(info.Warnings, "Ongeldige domeinnaam-syntax.")
        return info
    }

    // IDN verwerken
    var err error
    info.Unicode, err = toUnicode(info.DomainInput)
    if err != nil {
        info.Unicode = info.DomainInput // fallback
    }
    info.Punycode, err = toASCII(info.DomainInput)
    if err != nil {
        info.Punycode = info.DomainInput // fallback
    }

    // ---- Infrastructuur-TLD check (draft §2.6, Note 2) ----
    // _for-sale onder .arpa moet genegeerd worden.
    lc := strings.ToLower(info.Punycode)
    lc = strings.TrimSuffix(lc, ".") // verwijder trailing dot
    if strings.HasSuffix(lc, ".arpa") {
        info.Reasons = append(info.Reasons, "Domein valt onder .arpa infrastructuur en wordt genegeerd.")
        // Niet te koop (ForSale blijft false), maar wel context tonen.
        return info
    }

    leaf := "_for-sale." + info.Punycode

    // DNS lookup TXT
    txts, err := net.LookupTXT(leaf)
    if err != nil {
        info.Reasons = append(info.Reasons, "Geen TXT-records gevonden of lookup-fout.")
        return info
    }

    // Parseer alle TXT-records
    for _, rr := range txts {
        info.RawRR = append(info.RawRR, rr)
        ok, content := hasVersionPrefix(rr)
        if !ok {
            // §2.1: zonder versie-tag niet interpreteren als geldige indicator
            continue
        }
        parseForsaleRR(content, &info)
    }

    // Sorteer output voor consistente presentatie
    sort.Strings(info.FTxt)
    sort.Strings(info.FCod)
    sort.SliceStable(info.FUri, func(i, j int) bool { return info.FUri[i].URI < info.FUri[j].URI })
    sort.SliceStable(info.FVal, func(i, j int) bool { return info.FVal[i].FormattedNice < info.FVal[j].FormattedNice })

    // Indien niets geldig gevonden -> niet te koop
    if len(info.FTxt) == 0 && len(info.FUri) == 0 && len(info.FVal) == 0 && len(info.FCod) == 0 && info.ForSale == false {
        info.Reasons = append(info.Reasons, "Geen geldig _for-sale-indicator aangetroffen (geen versie-tag).")
    }

    return info
}

// ---------- HTTP ----------

var (
    tplLayout *template.Template
    tplIndex  *template.Template
    tplResult *template.Template
)

func mustParseTemplates() {
    base := filepath.Join("templates", "layout.gohtml")
    index := filepath.Join("templates", "index.gohtml")
    result := filepath.Join("templates", "result.gohtml")

    var err error
    tplLayout, err = template.ParseFiles(base)
    if err != nil {
        log.Fatalf("Kon layout.gohtml niet parsen: %v", err)
    }
    tplIndex, err = template.Must(tplLayout.Clone()).ParseFiles(index)
    if err != nil {
        log.Fatalf("Kon index.gohtml niet parsen: %v", err)
    }
    tplResult, err = template.Must(tplLayout.Clone()).ParseFiles(result)
    if err != nil {
        log.Fatalf("Kon result.gohtml niet parsen: %v", err)
    }
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
    type data struct {
        Now         time.Time
        DefaultDemo string
    }
    d := data{Now: time.Now(), DefaultDemo: "example.nl"}
    if err := tplIndex.ExecuteTemplate(w, "layout", d); err != nil {
        log.Printf("template index render error: %v", err)
        http.Error(w, "Template-fout: "+err.Error(), http.StatusInternalServerError)
    }
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
    domain := strings.TrimSpace(r.FormValue("domain"))
    if domain == "" {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
    info := checkDomain(domain)

    type data struct {
        Info SaleInfo
        Now  time.Time
    }
    d := data{Info: info, Now: time.Now()}
    if err := tplResult.ExecuteTemplate(w, "layout", d); err != nil {
        log.Printf("template result render error: %v", err)
        http.Error(w, "Template-fout: "+err.Error(), http.StatusInternalServerError)
    }
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
    domain := strings.TrimSpace(r.URL.Query().Get("domain"))
    if domain == "" {
        http.Error(w, "Missing ?domain=", http.StatusBadRequest)
        return
    }
    info := checkDomain(domain)
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    _ = enc.Encode(info)
}

func main() {
    // Logging
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    // Templates laden
    mustParseTemplates()

    // Static files
    fs := http.FileServer(http.Dir("static"))

    // Routes
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/check", handleCheck)
    http.HandleFunc("/api/check", handleAPI)
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Poort via env PORT of default 8080
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Forsale-web gestart op :%s", port)
    log.Printf("Open http://localhost:%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}
