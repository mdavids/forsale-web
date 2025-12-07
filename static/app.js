// ===========================
// Thema-toggle (dark/light)
// ===========================
const THEME_KEY = 'forsale.theme';

function applyTheme(theme){
    // Dit stelt het data-theme attribuut onmiddellijk in op <html>
    document.documentElement.setAttribute('data-theme', theme);

    // De knop-update kan in applyTheme blijven, de 'if (btn)' zorgt dat het niet crasht
    // als de knop nog niet in de DOM zit bij het initieel laden in de <head>.
    const btn = document.getElementById('themeToggle');
    if (btn){
        const isLight = theme === 'light';
        // Icoon + ARIA
        btn.textContent = isLight ? '🌓' : '🌙';
        btn.setAttribute('aria-label', isLight ? 'Schakel naar donker thema' : 'Schakel naar licht thema');
        btn.title = isLight ? 'Donker thema' : 'Licht thema';
    }
}

function detectSystemPref(){
    try {
        // Controleert de voorkeur van het besturingssysteem
        return window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
    } catch {
        return 'dark';
    }
}

function initTheme(){
    // 1. Lees eventuele opgeslagen keuze uit localStorage
    const saved = (() => {
        try { return localStorage.getItem(THEME_KEY); } catch { return null; }
    })();

    // 2. Bepaal het uiteindelijke thema: Opgeslagen > Systeemvoorkeur > Dark (als fallback)
    const theme = (saved === 'light' || saved === 'dark') ? saved : detectSystemPref();

    // 3. PAS HET THEMA DIRECT TOE ZONDER WACHTEN
    applyTheme(theme);
}

function toggleTheme(){
    // Haal het huidige thema op, of gebruik Systeemvoorkeur als fallback
    const current = document.documentElement.getAttribute('data-theme') || detectSystemPref();
    const next = current === 'light' ? 'dark' : 'light';
    applyTheme(next);
    try { localStorage.setItem(THEME_KEY, next); } catch {}
}

// 🎉 BELANGRIJK: ROEP DE FUNCTIE DIRECT AAN
// Dit zorgt ervoor dat het thema wordt ingesteld VOORDAT de browser de pagina rendert.
initTheme(); 


// ===========================
// Modal (bevestiging externe link)
// De rest van de code blijft ongewijzigd
// ===========================
function confirmOpen(href){
    const modal = document.getElementById('linkModal');
    const modalLink = document.getElementById('modalLink');
    if (!modal || !modalLink) return;
    modalLink.setAttribute('href', href);
    modal.classList.remove('hidden');
}

function closeModal(){
    const modal = document.getElementById('linkModal');
    if (!modal) return;
    modal.classList.add('hidden');
}

function openAbout(){
    alert("Deze tool controleert _for-sale TXT-records conform draft-davids-forsalereg-18.\n\nEigenschappen: geen automatische redirects, duidelijke waarschuwingen en IDN-bewuste weergave.\n\nDISCLAIMER: Dit is een demo-applicatie! Aan de weergegeven resultaten kunnen GEEN rechten worden ontleend.");
}

// Publiceer functies op window (voor inline onclick handlers)
window.toggleTheme = toggleTheme;
window.confirmOpen = confirmOpen;
window.closeModal = closeModal;
window.openAbout = openAbout;
