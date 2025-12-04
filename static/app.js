// ===========================
// Thema-toggle (dark/light)
// ===========================
const THEME_KEY = 'forsale.theme';

function applyTheme(theme){
  document.documentElement.setAttribute('data-theme', theme);
  const btn = document.getElementById('themeToggle');
  if (btn){
    const isLight = theme === 'light';
    // Icoon + ARIA
    //btn.textContent = isLight ? '🌞' : '🌙';
    btn.textContent = isLight ? '🌓' : '🌙';
    btn.setAttribute('aria-label', isLight ? 'Schakel naar donker thema' : 'Schakel naar licht thema');
    btn.title = isLight ? 'Donker thema' : 'Licht thema';
  }
}

function detectSystemPref(){
  try {
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
  } catch {
    return 'dark';
  }
}

function initTheme(){
  // Lees eventuele opgeslagen keuze
  const saved = (() => {
    try { return localStorage.getItem(THEME_KEY); } catch { return null; }
  })();

  const theme = (saved === 'light' || saved === 'dark') ? saved : detectSystemPref();

  // Subtiele init: korte delay zodat transitie niet "flasht" bij init
  requestAnimationFrame(() => {
    applyTheme(theme);
    document.body.classList.add('theme-ready'); // optioneel: hook voor extra effecten
  });
}

function toggleTheme(){
  const current = document.documentElement.getAttribute('data-theme') || detectSystemPref();
  const next = current === 'light' ? 'dark' : 'light';
  applyTheme(next);
  try { localStorage.setItem(THEME_KEY, next); } catch {}
}

// Init bij load
document.addEventListener('DOMContentLoaded', initTheme);

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
  alert("Deze tool controleert _for-sale TXT-records conform draft-davids-forsalereg-18.\n\nEigenschappen: geen automatische redirects, duidelijke waarschuwingen en IDN-bewuste weergave.");
}

// Publiceer functies op window (voor inline onclick handlers)
window.toggleTheme = toggleTheme;
window.confirmOpen = confirmOpen;
window.closeModal = closeModal;
window.openAbout = openAbout;
