package main

import (
	"net/http"
	"net/url"
	"strings"
)

// ---------- i18n ----------

const defaultLang = "nl"

var translations = map[string]map[string]string{
	"nl": {
		"title":               "Is deze domeinnaam te koop? · _for-sale checker",
		"nav.home":            "Home",
		"nav.about":           "Over",
		"theme.toggle.title":  "Thema wisselen",
		"footer.example_impl": "Voorbeeldimplementatie",
		"footer.based_on":     "op basis van de",
		"modal.title":         "Externe link openen?",
		"modal.body":          "Je staat op het punt een externe link te openen. Controleer de bestemming voordat je doorgaat.",
		"modal.cancel":        "Annuleren",
		"modal.continue":      "Doorgaan",

		"index.h1":           "Is deze domeinnaam te koop?",
		"index.lead_pre":     "Vul een domeinnaam in (bijv.",
		"index.lead_mid":     "). We controleren of er een",
		"index.lead_post":    "TXT-record aanwezig is en tonen relevante gegevens.",
		"index.domain_label": "Domeinnaam",
		"index.submit":       "Controleren",
		"index.hint":         "Werkt met Unicode domeinen (IDN). We tonen zowel Unicode als Punycode om misverstanden te voorkomen.",

		"result.for_title":           "Resultaat voor:",
		"result.punycode":            "Punycode:",
		"result.badge.forsale":       "Te koop",
		"result.badge.forsale.title": "Het lijkt erop dat de domeinnaam te koop is.",
		"result.badge.not":           "Niet te koop",
		"result.badge.not.title":     "We konden niet vaststellen of de domeinnaam te koop staat.",
		"result.price_title":         "Indicatieve prijs:",
		"result.price_disclaimer":    "Prijs is indicatief - verifieer bij verkoper.",
		"result.contact_title":       "Contact & meer informatie:",
		"result.unverifiable":        "oncontroleerbaar",
		"result.contact_disclaimer":  "We redirecten niet automatisch - jouw bevestiging is vereist.",
		"result.extra_title":         "Extra informatie:",
		"result.fcod_title":          "Verwerkingscode(s):",
		"result.fcod_hint":           "Deze codes zijn doorgaans alleen maar betekenisvol binnen een ecosysteem (bijv. registrar/registry).",
		"result.fcod_hint2":          "Als we een door SIDN en haar registrars gebruikte code voor landingspagina's vinden, tonen we die.",
		"result.empty_title":         "Geen geldig",
		"result.empty_title2":        "record gevonden.",
		"result.empty_body":          "Deze domeinnaam lijkt niet expliciet als te koop gemarkeerd.",
		"result.empty_whois":         "Je kunt informatie proberen te vinden via WHOIS/RDAP.",
		"result.empty_visit_pre":     "Of bezoek",
		"result.empty_visit_post":    "om te zien of daar meer informatie is te vinden.",
		"result.raw_summary":         "Geparseerde TXT-record(s)",
		"result.new_check":           "Nieuwe check",

		"js.aria_to_dark":  "Schakel naar donker thema",
		"js.aria_to_light": "Schakel naar licht thema",
		"js.title_dark":    "Donker thema",
		"js.title_light":   "Licht thema",
		"js.about":         "Deze tool controleert _for-sale TXT-records conform RFC 10023.\n\nEigenschappen: geen automatische redirects, duidelijke waarschuwingen en IDN-bewuste weergave.\n\nDISCLAIMER: Dit is een demo-applicatie! Aan de weergegeven resultaten kunnen GEEN rechten worden ontleend.",

		"reason.arpa":                 "Domein valt onder .arpa infrastructuur en wordt genegeerd.",
		"reason.no_txt":               "Geen TXT-records gevonden of lookup-fout.",
		"reason.empty_record":         "Leeg 'te koop' record (alleen versie-tag) aangetroffen.",
		"reason.empty_ftxt":           "Lege ftxt=-waarde",
		"reason.invalid_fval":         "Ongeldige fval-structuur; verwacht CUR123[.45]:",
		"reason.empty_fcod":           "Lege fcod=-waarde",
		"reason.unknown_tag":          "Onbekende of ongeldige content-tag:",
		"reason.invalid_syntax":       "Ongeldige domeinnaam-syntax.",
		"reason.no_valid_indicator":   "Geen geldig _for-sale-indicator aangetroffen (geen versie-tag).",
		"warning.invalid_furi":        "Aangetroffen furi is mogelijk ongeldig:",
		"note.uri_parse_failed":       "Kon URI niet parsen",
		"note.uri_missing_scheme":     "(URI ontbreekt of mist scheme http/https/mailto/tel)",
		"note.uri_unsupported_scheme": "(Niet-ondersteunde URI-scheme)",
		"error.template":              "Template-fout: ",
	},
	"en": {
		"title":               "Is this domain name for sale? · _for-sale checker",
		"nav.home":            "Home",
		"nav.about":           "About",
		"theme.toggle.title":  "Toggle theme",
		"footer.example_impl": "Example implementation",
		"footer.based_on":     "based on the",
		"modal.title":         "Open external link?",
		"modal.body":          "You are about to open an external link. Check the destination before continuing.",
		"modal.cancel":        "Cancel",
		"modal.continue":      "Continue",

		"index.h1":           "Is this domain name for sale?",
		"index.lead_pre":     "Enter a domain name (e.g.",
		"index.lead_mid":     "). We check whether a",
		"index.lead_post":    "TXT record is present and show the relevant details.",
		"index.domain_label": "Domain name",
		"index.submit":       "Check",
		"index.hint":         "Works with Unicode domains (IDN). We show both Unicode and Punycode to avoid confusion.",

		"result.for_title":           "Result for:",
		"result.punycode":            "Punycode:",
		"result.badge.forsale":       "For sale",
		"result.badge.forsale.title": "It looks like this domain name is for sale.",
		"result.badge.not":           "Not for sale",
		"result.badge.not.title":     "We could not determine whether the domain name is for sale.",
		"result.price_title":         "Indicative price:",
		"result.price_disclaimer":    "Price is indicative - verify with the seller.",
		"result.contact_title":       "Contact & more information:",
		"result.unverifiable":        "unverifiable",
		"result.contact_disclaimer":  "We do not redirect automatically - your confirmation is required.",
		"result.extra_title":         "Additional information:",
		"result.fcod_title":          "Processing code(s):",
		"result.fcod_hint":           "These codes are usually only meaningful within an ecosystem (e.g. registrar/registry).",
		"result.fcod_hint2":          "If we find a landing-page code used by SIDN and its registrars, we show it here.",
		"result.empty_title":         "No valid",
		"result.empty_title2":        "record found.",
		"result.empty_body":          "This domain name does not appear to be explicitly marked for sale.",
		"result.empty_whois":         "You can try to find information via WHOIS/RDAP.",
		"result.empty_visit_pre":     "Or visit",
		"result.empty_visit_post":    "to see if more information is available there.",
		"result.raw_summary":         "Parsed TXT record(s)",
		"result.new_check":           "New check",

		"js.aria_to_dark":  "Switch to dark theme",
		"js.aria_to_light": "Switch to light theme",
		"js.title_dark":    "Dark theme",
		"js.title_light":   "Light theme",
		"js.about":         "This tool checks _for-sale TXT records conform RFC 10023.\n\nFeatures: no automatic redirects, clear warnings, and IDN-aware display.\n\nDISCLAIMER: This is a demo application! No rights can be derived from the results shown.",

		"reason.arpa":                 "Domain falls under .arpa infrastructure and is ignored.",
		"reason.no_txt":               "No TXT records found or lookup error.",
		"reason.empty_record":         "Empty 'for sale' record (version tag only) found.",
		"reason.empty_ftxt":           "Empty ftxt= value",
		"reason.invalid_fval":         "Invalid fval structure; expected CUR123[.45]:",
		"reason.empty_fcod":           "Empty fcod= value",
		"reason.unknown_tag":          "Unrecognized or invalid content tag:",
		"reason.invalid_syntax":       "Invalid domain name syntax.",
		"reason.no_valid_indicator":   "No valid _for-sale indicator found (no version tag).",
		"warning.invalid_furi":        "Found furi may be invalid:",
		"note.uri_parse_failed":       "Could not parse URI",
		"note.uri_missing_scheme":     "(URI missing or missing scheme http/https/mailto/tel)",
		"note.uri_unsupported_scheme": "(Unsupported URI scheme)",
		"error.template":              "Template error: ",
	},
}

// T vertaalt een sleutel naar de opgegeven taal, met terugval op het Nederlands
// en als laatste redmiddel de sleutel zelf (zichtbaar in de UI bij een ontbrekende vertaling,
// zodat het meteen opvalt in plaats van stil te falen).
func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := translations[defaultLang][key]; ok {
		return v
	}
	return key
}

// detectLang leest de taal uit ?lang=; alles behalve "en" valt terug op Nederlands.
func detectLang(r *http.Request) string {
	if r.URL.Query().Get("lang") == "en" {
		return "en"
	}
	return defaultLang
}

// AddLang voegt (indien nodig) ?lang=en of &lang=en toe aan een pad, voor gebruik
// in templates zodat interne links de taalkeuze vasthouden (stateless via de URL).
func AddLang(lang, path string) string {
	if lang != "en" {
		return path
	}
	if strings.Contains(path, "?") {
		return path + "&lang=en"
	}
	return path + "?lang=en"
}

// langSwitchURL bouwt de URL voor de taal-wisselknop: huidige pad + query,
// met de taal omgezet naar de andere waarde.
func langSwitchURL(r *http.Request, lang string) string {
	q := r.URL.Query()
	if lang == "en" {
		q.Del("lang")
	} else {
		q.Set("lang", "en")
	}
	u := url.URL{Path: r.URL.Path, RawQuery: q.Encode()}
	return u.String()
}
