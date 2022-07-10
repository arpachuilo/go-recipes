// strike out text
function strike(el) {
  el.classList.toggle("strike");
}

window.strike = strike;

// image preview
function setImageSrc(el, imgID, fallback) {
  const img = document.getElementById(imgID);
  const [file] = el.files;
  if (file) {
    img.src = URL.createObjectURL(file);
  } else {
    img.src = fallback;
  }
}

window.setImageSrc = setImageSrc;

// focus search
function slashToFocus(el) {
  window.addEventListener("keydown", (e) => {
    if (document.activeElement === el) return;
    if (e.code === "Slash") {
      el.focus();
      e.preventDefault();
    }
  });
}

window.slashToFocus = slashToFocus;

// auto search limit size
function setSearchLimit() {
  let el = document.getElementById("search-results");
  if (el === null) return;
  let name = "search_limit";
  let value = Math.max(Math.floor(Math.min(1320, el.offsetWidth) / 200), 2) * 5;
  let expiry = new Date();
  let domain = window.location.hostname;
  expiry.setMonth(expiry.getMonth() + 2);
  let cookie = `${name}=${value};expires=${expiry};domain=${domain};path=/`;
  document.cookie = cookie;
}

window.onload = function () {
  if (!navigator.cookieEnabled) return;
  const urlSearchParams = new URLSearchParams(window.location.search);
  const params = Object.fromEntries(urlSearchParams.entries());
  let el = document.getElementById("limit");
  if ("limit" in params) {
    el.value = params.limit;
  } else {
    if (document.cookie.match(/^(.*;)?\s*search_limit\s*=\s*[^;]+(.*)?$/)) {
      el.value = "auto";
    }
  }

  setSearchLimit();
};

if (navigator.cookieEnabled) {
  window.onresize = setSearchLimit;
}

// nice
if ("scrollRestoration" in history) {
  history.scrollRestoration = "auto";
}

// scroll sync
function syncSrolls(_tx, _rx) {
  const tx = document.querySelector(_tx);
  const rx = document.querySelector(_rx);

  let hovered = false;
  let offset = 0;
  let scrollOfset = 0;
  const sync = () => {
    // only apply if below certain width
    if (window.innerWidth < 480) {
      offset = 0;
      scrollOfset = 0;
      return;
    }

    // compute max
    let max = tx.offsetHeight - rx.offsetHeight;
    max = Math.max(0, max);

    // sticky scroll
    let rxb = rx.getBoundingClientRect();
    if (hovered) {
      scrollOfset = Math.min(rxb.top, 0);
    }

    // reset sticky scroll
    if (rxb.top > 0) {
      scrollOfset = 0;
    }

    // no need to move
    if (offset < 0 && rx.style.marginTop == "0px") {
      return;
    }

    offset = tx.getBoundingClientRect().top * -1 + scrollOfset;
    offset = Math.max(0, offset);
    offset = Math.min(offset, max);
    rx.style.marginTop = offset + "px";
  };

  rx.addEventListener("mouseenter", (_) => {
    hovered = true;
  });

  rx.addEventListener("mouseleave", (_) => {
    hovered = false;
  });

  window.addEventListener("scroll", (_) => {
    window.requestAnimationFrame(sync);
  });
}

window.syncSrolls = syncSrolls;
