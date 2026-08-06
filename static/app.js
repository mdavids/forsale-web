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
        const i18n = window.I18N || {};
        // Icoon + ARIA
        btn.textContent = isLight ? '🌓' : '🌙';
        btn.setAttribute('aria-label', isLight ? i18n.ariaToDark : i18n.ariaToLight);
        btn.title = isLight ? i18n.titleDark : i18n.titleLight;
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

    // 3. Pas het thema direct aan zonder te wachten
    applyTheme(theme);
}

function toggleTheme(){
    // Haal het huidige thema op, of gebruik Systeemvoorkeur als fallback
    const current = document.documentElement.getAttribute('data-theme') || detectSystemPref();
    const next = current === 'light' ? 'dark' : 'light';
    applyTheme(next);
    try { localStorage.setItem(THEME_KEY, next); } catch {}
}

// Dit zorgt ervoor dat het thema wordt ingesteld VOORDAT de browser de pagina rendert.
initTheme(); 


// ===========================
// Modal (bevestiging externe link)
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
    alert((window.I18N && window.I18N.about) || '');
}

// Publiceer functies op window (voor inline onclick handlers)
window.toggleTheme = toggleTheme;
window.confirmOpen = confirmOpen;
window.closeModal = closeModal;
window.openAbout = openAbout;


// Confetti - https://github.com/catdad/canvas-confetti
document.addEventListener('DOMContentLoaded', () => {
    // Deze code wordt pas uitgevoerd NADAT alle HTML is geparsd
    // en de elementen met de class .js-trigger-confetti beschikbaar zijn.

    const confettiElement = document.querySelector('.js-trigger-confetti');

    if (confettiElement) {
        // We weten nu zeker dat de confetti functie (van de CDN) bestaat
        // en het HTML-element is gevonden. Tijd voor confetti!
        
        // Gebruik 'load' is hier niet per se nodig, omdat we al DOMContentLoaded gebruiken.
        // We kunnen de confetti direct starten, of met een kleine vertraging.
        
        confetti({
            particleCount: 200,
            spread: 90,
            origin: { y: 0.65 }
        });
    }

    // Andere code van app.js die met themes/CSS werkt kan hier ook staan.
});
